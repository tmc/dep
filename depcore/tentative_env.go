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

// checks if the updated version of package pkgPath is in conflict with
// packages of the original environment
//func (tent *tentativeEnvironment) checkConflicts(pkgPath string) (errs map[string][3]string) {
func (tent *tentativeEnvironment) checkConflicts(pkg *gdf.Package) (errs map[string][3]string) {
	//if !tent.Original.PkgExists(pkgPath) {
	//	panic(fmt.Sprintf("package %s is not installed in %s/src", pkgPath, tent.Original.GOPATH))
	//}
	//p, err := tent.Pkg(pkgPath)
	//if err != nil {
	//	panic(fmt.Sprintf("can't check conflicts for package %s: %s", pkgPath, err))
	//}
	return tent.Original.db.hasConflict(pkg)
}

// closes db and remove the temporary gopath
func (tent *tentativeEnvironment) Close() {
	tent.Environment.Close()
	os.RemoveAll(tent.GOPATH)
}

// returns the repos that are candidates for the real update
// that will be a movement
func (tempEnv *tentativeEnvironment) getCandidates() (pkgs []*gdf.Package) {
	o := tempEnv.Original
	//ps := tempEnv.allPackages()
	skip := map[string]bool{}
	pkgs = []*gdf.Package{}

	for _, p := range tempEnv.allPackages() {
		r := _repoRoot(p.Dir)
		if skip[r] {
			continue
		}

		dirInOriginal, _, dirErr := o.PkgDir(p.Path)

		if dirErr == nil {
			revNew := tempEnv.getRevision(p.Dir, "")
			revOld := o.getRevision(dirInOriginal, "")
			if revNew.Rev == revOld.Rev {
				skip[r] = true
				continue
			}
		}
		pkgs = append(pkgs, p)
	}

	return
}

func moveToBackup(dir string) (err error) {
	backup := dir + fmt.Sprintf("_backup_of_dep_update_%v", time.Now().UnixNano())
	err = os.Rename(dir, backup)
	return
}

/*
	TODO

	make tests for
		- a package within a repo that is a subsubdir of GOPATH/src
		- a package within a repo that is a subdir of GOPATH/src
		- a package that is a repo that is a subsubdir of GOPATH/src
		- a package that is a repo that is a subdir of GOPATH/src
*/
// returns a path of absolute path that is relative to the given parent path
func relativePath(parentPath, childPath string) (rel string, err error) {
	if !strings.Contains(childPath, parentPath) {
		err = fmt.Errorf("%s is not within %s", childPath, parentPath)
		return
	}
	rel, err = filepath.Rel(parentPath, childPath)
	return
}

func (tempEnv *tentativeEnvironment) moveCandidatesToGOPATH(pkgs ...*gdf.Package) (err error) {
	//func (tempEnv *tentativeEnvironment) moveCandidatesToGOPATH(pkgs ...string) (err error) {
	o := tempEnv.Original
	visited := map[string]bool{}

	for _, pkg := range pkgs {
		//revBefore := o.getRevisionGit(o.PkgDir(pkg.Path))
		//fmt.Printf("rev before: %#v\n", revBefore)
		//dir := tempEnv.PkgDir(pkg.Path)
		//dir := tempEnv.PkgDir(pkg)
		dir := pkg.Dir
		//fmt.Printf("rev in temp: %#v\n", tempEnv.getRevisionGit(dir))
		_, e := os.Stat(dir)

		// already moved
		if e != nil {
			// fmt.Printf("already moved: %#v\n", dir)
			continue
		}

		r := _repoRoot(dir)
		if visited[r] {
			continue
		}
		visited[r] = true
		var relRepoPath string
		relRepoPath, err = relativePath(path.Join(tempEnv.GOPATH, "src"), r)
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
		//revAfter := o.getRevisionGit(o.PkgDir(pkg.Path))
		//fmt.Printf("rev after: %#v\n", revAfter)
	}
	return
}

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

// return no errors for conflicts, only for severe errors
func (tentative *tentativeEnvironment) updatePackage(pkg string, confirmation func(candidates ...*gdf.Package) bool) (conflicts map[string]map[string][3]string, err error) {
	//conflicts = map[string]map[string][3]string{}
	err = tentative.getPackage(pkg)
	if err != nil {
		return
	}

	//tempEnv := NewEnv(tmpDir)
	err = tentative.checkoutTrackedImports(pkg)

	if err != nil {
		return
	}

	//tentative.mkdb()
	conflicts = tentative.Init()
	/*
	   err = createDB(tentative.GOPATH)
	   if err != nil {
	       return
	   }
	*/

	//conflicts = tentative.CheckIntegrity()
	if len(conflicts) > 0 {
		// fmt.Println("NO INTEGRITY for TENTATIVE")
		err = fmt.Errorf("tentative GOPATH %s is not integer", tentative.GOPATH)
		return
	}

	candidates := tentative.getCandidates()

	for _, candidate := range candidates {
		//fmt.Printf("candidate: %v\n", candidate.Path)
		//if tentative.Original.PkgExists(candidate.Path) {
		if tentative.Original.PkgExists(candidate.Path) {
			//	fmt.Println("exists")
			//errs := tentative.checkConflicts(candidate.Path)
			errs := tentative.checkConflicts(candidate)
			if len(errs) > 0 {
				//		fmt.Printf("errors: %v\n", len(errs))
				//conflicts[candidate.Path] = errs
				conflicts[candidate.Path] = errs
			}
		}
	}
	if len(conflicts) == 0 {
		if confirmation(candidates...) {
			err = tentative.moveCandidatesToGOPATH(candidates...)
		}
	}
	return
}
