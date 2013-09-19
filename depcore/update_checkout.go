package depcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	"io/ioutil"
	"path"
	"path/filepath"
	// "runtime"
	"time"
	// "github.com/metakeule/cli"

	"os"
	"os/exec"
)

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
func (tentative *TentativEnvironment) _updatePackage(pkg string) (conflicts map[string]map[string][3]string, err error) {
	conflicts = map[string]map[string][3]string{}
	err = tentative.Original.goGetPackages(tentative.GOPATH, pkg)
	if err != nil {
		return
	}

	//tempEnv := NewEnv(tmpDir)
	err = tentative.Original.checkoutDependanciesByRevFile(tentative.GOPATH, pkg)

	if err != nil {
		return
	}

	tentative.mkdb()
	/*
		err = createDB(tentative.GOPATH)
		if err != nil {
			return
		}
	*/

	err = tentative.checkIntegrity()
	if err != nil {
		return
	}

	candidates := tentative.getCandidatesForMovement()

	for _, candidate := range candidates {
		fmt.Printf("candidate: %v\n", candidate.Path)
		if tentative.Original.PkgExists(candidate.Path) {
			fmt.Println("exists")
			errs := tentative.checkConflicts(candidate)
			if len(errs) > 0 {
				fmt.Printf("errors: %v\n", len(errs))
				conflicts[candidate.Path] = errs
			}
		}
	}
	if len(conflicts) == 0 {
		err = tentative.moveCandidatesToGOPATH(candidates...)
	}
	return
}

func (env *Environment) goGetPackages(tmpDir string, pkg string) error {
	//args := []string{"get", "-u", pkg}
	// With get -d we don't install the packages
	// TODO check if we want to install them at a later point
	// e.g. after the dependant packages have been checked out
	// to the correct revisions
	args := []string{"get", "-u", "-d", pkg}
	cmd := exec.Command("go", args...)
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, tmpDir),
		fmt.Sprintf(`GOROOT=%s`, env.GOROOT),
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error while go getting %#v:%s\n", pkg, stdout.String()+"\n"+stderr.String())
	}
	return nil
}

func (env *Environment) checkoutRevision(r string, rev revision) {
	var checkoutErr error
	switch rev.VCM {
	case "bzr":
		checkoutErr = env.checkoutBzr(r, rev.Rev)
	case "git":
		checkoutErr = env.checkoutGit(r, rev.Rev)
	case "hg":
		checkoutErr = env.checkoutHg(r, rev.Rev)
	case "svn":
		panic("unsupported VCM svn for repository " + r)
	default:
		panic("unsupported VCM " + rev.VCM + " for repository " + r)
	}

	if checkoutErr != nil {
		panic("can't checkout " + r + " rev " + rev.Rev + ":\n" + checkoutErr.Error())
	}
}

func (env *Environment) mkdb() {
	// func getDB(gopath string) *db.DB {
	_, dbFileerr := os.Stat(dEP(env.GOPATH))
	dB, err := db.Open(dEP(env.GOPATH))
	if err != nil {
		panic(err.Error())
	}
	if dbFileerr != nil {
		// fmt.Println(dbFileerr)
		db.CreateTables(dB)
	}
	d := &DB{}
	d.Environment = env
	d.DB = dB
	env.DB = d
}

func (env *Environment) checkIntegrity() (err error) {
	env.mkdb()
	//defer dB.Close()
	conflicts := map[string]map[string][3]string{}

	ps := env.allPackages()
	// fmt.Printf("all packages: %v\n", len(ps))
	env.DB.registerPackages(ps...)
	for _, p := range ps {
		// TODO: do something like hasConflict() but for a new db
		errs := env.DB.hasConflict(p)
		if len(errs) > 0 {
			conflicts[p.Path] = errs
		}
		/*
			res, e := runGoTest(o, env.GOPATH, path.Join(env.GOPATH, "src", p.Path))
			if e != nil {
				panic(fmt.Sprintf("Error while running test for package %s in tempdir:\n%s\n", p.Path, res))
			}
		*/
	}
	if len(conflicts) > 0 {
		b, e := json.MarshalIndent(conflicts, "", "  ")
		if e != nil {
			panic(e.Error())
		}
		fmt.Printf("%s\n", b)
		//Exit(UpdateConflict)
		return fmt.Errorf("integrity conflict in GOPATH %s", env.GOPATH)
		// panic("update conflict")
	}
	return nil
}

func (tempEnv *TentativEnvironment) getCandidatesForMovement() (pkgs []*exports.Package) {
	// TODO for all packages in tempEnv: check if they are in GOPATH and if the revision of the repo is the same
	o := tempEnv.Original
	skip := map[string]bool{}
	ps := tempEnv.allPackages()
	pkgs = []*exports.Package{}

	for _, p := range ps {
		//fmt.Printf("a package in ", ...)
		dir, _ := p.Dir()
		r := _repoRoot(dir)
		if skip[r] {
			continue
		}

		_, err := os.Stat(o.PkgDir(p.Path))
		/*
			if err != nil {
				fmt.Printf("can't find: %s\n", path.Join(o.GOPATH, "src", p.Path))
			}
		*/
		if err == nil {
			revNew := o.pkgRevision(dir, "")
			revOld := o.pkgRevision(o.PkgDir(p.Path), "")
			if revNew.Rev == revOld.Rev {
				skip[r] = true
				continue
			}
		}

		// fmt.Printf("add candidate: %s\n", p.Path)
		pkgs = append(pkgs, p)
	}

	return
}

func (tempEnv *TentativEnvironment) moveCandidatesToGOPATH(pkgs ...*exports.Package) (err error) {
	o := tempEnv.Original
	visited := map[string]bool{}

	for _, pkg := range pkgs {
		revBefore := o.getRevisionGit(o.PkgDir(pkg.Path))
		fmt.Printf("rev before: %#v\n", revBefore)
		dir := tempEnv.PkgDir(pkg.Path)
		_, e := os.Stat(dir)

		// already moved
		if e != nil {
			fmt.Printf("already moved: %#v\n", dir)
			continue
		}

		r := _repoRoot(dir)
		if visited[r] {
			continue
		}
		visited[r] = true
		target := _repoRoot(o.PkgDir(pkg.Path))
		backup := target + fmt.Sprintf("_backup_of_dep_update_%v", time.Now().UnixNano())
		err = os.Rename(target, backup)
		if err != nil {
			fmt.Printf("can't make backup: %s\n", backup)
			return
			//panic("can't make backup: " + backup)
		}
		err = os.Rename(r, target)
		fmt.Printf("try to move  %#v to %#v\n", r, target)
		if err != nil {
			fmt.Printf("can't move  %#v to %#v\n", r, target)
			return
			//panic(err.Error())
		}
		revAfter := o.getRevisionGit(o.PkgDir(pkg.Path))
		fmt.Printf("rev after: %#v\n", revAfter)
	}
	return
}

func (env *Environment) getDependancyRevisions(pkg string) (r map[string]revision, err error) {
	var data []byte
	data, err = ioutil.ReadFile(path.Join(env.GOPATH, "src", pkg, revFileName))
	if err != nil {
		return
	}

	r = map[string]revision{}
	err = json.Unmarshal(data, &r)
	return
}

func (env *Environment) checkoutDependanciesByRevFile(gopath string, pkg string) error {
	revisions, err := env.getDependancyRevisions(pkg)

	if err != nil {
		return err
	}

	visited := map[string]bool{}

	for p, rev := range revisions {
		dir := path.Join(gopath, "src", p)
		r := _repoRoot(dir)
		// fmt.Printf("repoROOT is: %#v\n", r)
		if visited[r] {
			continue
		}
		visited[r] = true
		// fmt.Printf("checking out: \n\tpkg %v\n\trev: %s\n\n", p, rev.Rev)
		env.checkoutRevision(r, rev)
	}
	return nil
}

func (dB *DB) uPdatePackage(pkg string) error {
	//tmpDir := o.mkdirTempDir()
	tentative := dB.Environment.NewTentative()
	//conflicts, err := o._updatePackage(tmpDir, dB, pkg)
	conflicts, err := tentative._updatePackage(pkg)

	if err != nil {
		return err
	}

	if len(conflicts) > 0 {
		b, e := json.MarshalIndent(conflicts, "", "  ")
		if e != nil {
			panic(e.Error())
		}
		fmt.Printf("%s\n", b)
		return fmt.Errorf("update conflict")
	}
	return nil
}

func (env *Environment) allPackages() (a []*exports.Package) {
	a = []*exports.Package{}
	// prs := &allPkgParser{map[string]bool{}}
	prs := newAllPkgParser(env.Environment)
	err := filepath.Walk(path.Join(env.GOPATH, "src"), prs.Walker)
	if err != nil {
		panic(err.Error())
	}

	for fp, _ := range prs.packages {
		//pkg := o.Env.Pkg(fp)
		pkg := env.Pkg(fp)
		if pkg.Internal {
			continue
		}
		a = append(a, pkg)
	}
	return
}
