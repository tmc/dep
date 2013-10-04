package depcore

import (
	"fmt"
	"github.com/metakeule/gdf"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// environment during a tentative update
type tentativeEnvironment struct {
	*Environment
	Original *Environment
}

// closes db and remove the temporary gopath
func (tent *tentativeEnvironment) Close() {
	tent.Environment.Close()
	os.RemoveAll(tent.GOPATH)
}

// returns the repos that are candidates for the real update that will be a movement
func (tent *tentativeEnvironment) getCandidates() (pkgs []*gdf.Package) {
	skip := map[string]bool{}
	pkgs = []*gdf.Package{}

	for _, p := range tent.allPackages() {
		r := _repoRoot(p.Dir)
		if skip[r] {
			continue
		}
		origDir, _, err := tent.Original.PkgDir(p.Path)
		// package is updated
		if err == nil {
			// package is updated only, if the revision changed
			if tent.getRevision(p.Dir, "").Rev == tent.Original.getRevision(origDir, "").Rev {
				skip[r] = true
				continue
			}
		}

		pkgs = append(pkgs, p)
	}
	return
}

// move directory to a backup
func moveToBackup(dir string) (err error) {
	backup := dir + fmt.Sprintf("_%v"+backupString, time.Now().UnixNano())
	err = os.Rename(dir, backup)
	return
}

// returns a path of absolute path that is relative to the given parent path
func relativePath(parentPath, childPath string) (rel string, err error) {
	if !strings.Contains(childPath, parentPath) {
		err = fmt.Errorf("%s is not within %s", childPath, parentPath)
		return
	}
	rel, err = filepath.Rel(parentPath, childPath)
	return
}

func (tent *tentativeEnvironment) movePackages(pkgs ...*gdf.Package) (err error) {
	o := tent.Original
	visited := map[string]bool{}

	for _, pkg := range pkgs {
		dir := pkg.Dir
		_, e := os.Stat(dir)

		// already moved
		if e != nil {
			continue
		}

		r := _repoRoot(dir)
		if visited[r] {
			continue
		}
		visited[r] = true
		var relRepoPath string
		relRepoPath, err = relativePath(path.Join(tent.GOPATH, "src"), r)
		if err != nil {
			return
		}

		target := path.Join(o.GOPATH, "src", relRepoPath)
		if _, errExists := os.Stat(target); errExists == nil {
			err = moveToBackup(target)
			if err != nil {
				return
			}
		}
		if _, errExists := os.Stat(filepath.Dir(target)); errExists != nil {
			os.MkdirAll(filepath.Dir(target), 0755)
		}

		err = os.Rename(r, target)
		if err != nil {
			return
		}
	}
	return
}

/*
// todo make it work
func hasConflict(pkg *gdf.Package, override *gdf.Package, ignoring map[string]bool) (errors map[string][3]string) {
	errors = map[string][3]string{}

	imp, err := dB.GetImported(pkg.Path)
	if err != nil {
		errors[pkg.Path] = [3]string{"error", err.Error(), ""}
		return
	}

	for _, im := range imp {
		if ignoring[im.Package] {
			continue
		}
		key := fmt.Sprintf("%s:%s", im.Package, im.Name)
		if val, exists := pkg.Exports[im.Name]; exists {
			if val != im.Value {
				errors[key] = [3]string{"changed", im.Value, val}
			}
			continue
		}
		//fmt.Printf("package %s \n\timports %s of %s, \n\tbut that has \n\t%s\n\n", im.Package, im.Name, pkg.Path, strings.Join(mapkeys(pkg.Exports), "\n\t"))
		errors[key] = [3]string{"removed", im.Value, ""}
	}
	return
}
*/

/*
   TODO

    (checkout / update is only for one package at a time)
   1. go get package into tempdir
   2. (for all packages in tempdir/package/dep-rev.json): checkout revisions into tempdir (each repo only once)
   3. check integrity for all packages in tempdir, run to test on each
   4. get candidates for movement to GOPATH: all repos in tempdir, that either aren't in GOPATH or have different revisions
   5. (for all candidates) check if updates packages won't break packages in GOPATH, if so return errors
   6. move candicate repos to path and go install them, update the registry
*/
/*
   tentative.Original.db
*/
// return no errors for conflicts, only for severe errors
func (tentative *tentativeEnvironment) updatePackage(pkg string, overrides []*gdf.Package, confirmation func(candidates ...*gdf.Package) bool) (conflicts map[string]map[string][3]string, err error) {
	g := newPackageGetter(tentative.Environment, pkg)
	err = g.get()
	if err != nil {
		return
	}

	conflicts = tentative.Init()
	if VERBOSE {
		fmt.Println("tentative GOPATH initialized")
	}

	if len(conflicts) > 0 {
		err = fmt.Errorf("tentative GOPATH %s is not integer", tentative.GOPATH)
		return
	}

	dB := tentative.Original.db

	// if we have overrides, make a copy of the db
	// and change the overrides in the new db
	if len(overrides) > 0 {
		dB, err = tentative.Original.cpdb(tentative.Environment, "temp.db")
		if err != nil {
			return
		}

		defer func() {
			dB.Close()
		}()

		dB.registerPackages(false, overrides...)
	}

	candidates := tentative.getCandidates()

	// to ignore conflicts of dependencies between the
	// packages that are all to be updated, ignore them
	// this should be save, since their compatibility has already been
	// checked in the tentative GOPATH
	ignoring := map[string]bool{}
	for _, candidate := range candidates {
		ignoring[candidate.Path] = true
	}

	for _, candidate := range candidates {
		if tentative.Original.PkgExists(candidate.Path) {
			errs := tentative.Original.db.hasConflict(candidate, ignoring)
			if len(errs) > 0 {
				conflicts[candidate.Path] = errs
			}
		}
	}
	if len(conflicts) == 0 {
		if confirmation(candidates...) {
			err = tentative.movePackages(candidates...)
			tentative.Original.db.registerPackages(false, candidates...)
		}
	}
	return
}
