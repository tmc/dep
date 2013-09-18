package dep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metakeule/dep/db"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime"
	"time"
	// "github.com/metakeule/cli"
	"github.com/metakeule/exports"
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

func GoGetPackages(o *Options, tmpDir string, pkg string) error {
	//args := []string{"get", "-u", pkg}
	// With get -d we don't install the packages
	// TODO check if we want to install them at a later point
	// e.g. after the dependant packages have been checked out
	// to the correct revisions
	args := []string{"get", "-u", "-d", pkg}
	cmd := exec.Command("go", args...)
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, tmpDir),
		fmt.Sprintf(`GOROOT=%s`, o.GOROOT),
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

func CheckoutRevision(o *Options, r string, rev Revision) {
	var checkoutErr error
	switch rev.VCM {
	case "bzr":
		checkoutErr = checkoutBzr(o, r, rev.Rev)
	case "git":
		checkoutErr = checkoutGit(o, r, rev.Rev)
	case "hg":
		checkoutErr = checkoutHg(o, r, rev.Rev)
	case "svn":
		panic("unsupported VCM svn for repository " + r)
	default:
		panic("unsupported VCM " + rev.VCM + " for repository " + r)
	}

	if checkoutErr != nil {
		panic("can't checkout " + r + " rev " + rev.Rev + ":\n" + checkoutErr.Error())
	}
}

func CheckIntegrity(o *Options, env *exports.Environment) (err error) {
	var dB *db.DB
	dB, err = db.Open(path.Join(env.GOPATH, "dep.db"))
	if err != nil {
		return
	}

	defer dB.Close()
	conflicts := map[string]map[string][3]string{}

	ps := allPackages(env)
	// fmt.Printf("all packages: %v\n", len(ps))
	registerPackages(env, dB, ps...)
	for _, p := range ps {
		// TODO: do something like hasConflict() but for a new db
		errs := hasConflict(dB, p)
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

func getCandidatesForMovement(o *Options, tempEnv *exports.Environment) (pkgs []*exports.Package) {
	// TODO for all packages in tempEnv: check if they are in GOPATH and if the revision of the repo is the same
	skip := map[string]bool{}
	ps := allPackages(tempEnv)
	pkgs = []*exports.Package{}

	for _, p := range ps {
		//fmt.Printf("a package in ", ...)
		dir, _ := p.Dir()
		r := _repoRoot(dir)
		if skip[r] {
			continue
		}
		_, err := os.Stat(path.Join(o.GOPATH, "src", p.Path))
		/*
			if err != nil {
				fmt.Printf("can't find: %s\n", path.Join(o.GOPATH, "src", p.Path))
			}
		*/
		if err == nil {
			revNew := pkgRevision(o, dir, "")
			revOld := pkgRevision(o, path.Join(o.GOPATH, "src", p.Path), "")
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

func moveCandidatesToGOPATH(o *Options, tempEnv *exports.Environment, pkgs ...*exports.Package) (err error) {
	visited := map[string]bool{}

	for _, pkg := range pkgs {
		dir := path.Join(tempEnv.GOPATH, "src", pkg.Path)
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
		target := _repoRoot(path.Join(o.GOPATH, "src", pkg.Path))
		backup := target + fmt.Sprintf("_backup_of_dep_update_%v", time.Now().UnixNano())
		err = os.Rename(target, backup)
		if err != nil {
			return
			//panic("can't make backup: " + backup)
		}
		err = os.Rename(r, target)
		if err != nil {
			return
			//panic(err.Error())
		}
	}
	return
}

func GetDependancyRevisions(gopath, pkg string) (r map[string]Revision, err error) {
	var data []byte
	data, err = ioutil.ReadFile(path.Join(gopath, "src", pkg, revFileName))
	if err != nil {
		return
	}

	r = map[string]Revision{}
	err = json.Unmarshal(data, &r)
	return
}

func CheckoutDependanciesByRevFile(o *Options, gopath string, pkg string) error {
	revisions, err := GetDependancyRevisions(gopath, pkg)

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
		CheckoutRevision(o, r, rev)
	}
	return nil
}

// return no errors for conflicts, only for severe errors
func _updatePackage(tmpDir string, o *Options, dB *db.DB, pkg string) (conflicts map[string]map[string][3]string, err error) {
	conflicts = map[string]map[string][3]string{}
	err = GoGetPackages(o, tmpDir, pkg)
	if err != nil {
		return
	}
	tempEnv := exports.NewEnv(runtime.GOROOT(), tmpDir)
	err = CheckoutDependanciesByRevFile(o, tempEnv.GOPATH, pkg)

	if err != nil {
		return
	}

	err = CreateDB(tempEnv.GOPATH)
	if err != nil {
		return
	}

	err = CheckIntegrity(o, tempEnv)
	if err != nil {
		return
	}

	candidates := getCandidatesForMovement(o, tempEnv)

	for _, candidate := range candidates {
		if o.Env.PkgExists(candidate.Path) {
			errs := checkConflicts(o, dB, tempEnv, candidate)
			if len(errs) > 0 {
				conflicts[candidate.Path] = errs
			}
		}
	}
	if len(conflicts) == 0 {
		err = moveCandidatesToGOPATH(o, tempEnv, candidates...)
	}
	return
}

func UpdatePackage(o *Options, dB *db.DB, pkg string) error {
	tmpDir := mkdirTempDir(o)
	conflicts, err := _updatePackage(tmpDir, o, dB, pkg)

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

func allPackages(env *exports.Environment) (a []*exports.Package) {
	a = []*exports.Package{}
	// prs := &allPkgParser{map[string]bool{}}
	prs := newAllPkgParser(env)
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
