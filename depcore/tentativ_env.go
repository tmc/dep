package depcore

import (
	"fmt"
	"os"
	"time"
)

// environment during a tentative update
type tentativeEnvironment struct {
	*Environment
	Original *Environment
}

// checks if the updated version of package pkgPath is in conflict with
// packages of the original environment
func (tent *tentativeEnvironment) checkConflicts(pkgPath string) (errs map[string][3]string) {
	if !tent.Original.PkgExists(pkgPath) {
		panic(fmt.Sprintf("package %s is not installed in %s/src", pkgPath, tent.Original.GOPATH))
	}
	return tent.Original.DB.hasConflict(tent.Pkg(pkgPath))
}

// closes db and remove the temporary gopath
func (tent *tentativeEnvironment) Close() {
	tent.Environment.Close()
	os.RemoveAll(tent.GOPATH)
}

// returns the repos that are candidates for the real update
// that will be a movement
func (tempEnv *tentativeEnvironment) getCandidates() (pkgs []string) {
	o := tempEnv.Original
	skip := map[string]bool{}
	ps := tempEnv.allPackages()
	pkgs = []string{}

	for _, p := range ps {
		dir, _ := p.Dir()
		r := _repoRoot(dir)
		if skip[r] {
			continue
		}
		//fmt.Printf("could be a candidate: %s \n", r)

		_, err := os.Stat(o.PkgDir(p.Path))

		if err != nil {
			//	fmt.Printf("can't find: %s\n", o.PkgDir(p.Path))
		}

		if err == nil {
			revNew := tempEnv.getRevision(dir, "")
			revOld := o.getRevision(o.PkgDir(p.Path), "")
			//	fmt.Printf("n: %s o: %s\n", revNew.Rev, revOld.Rev)
			if revNew.Rev == revOld.Rev {
				skip[r] = true
				continue
			}
		}

		//fmt.Printf("add candidate: %s\n", p.Path)
		pkgs = append(pkgs, p.Path)
	}

	return
}

//func (tempEnv *tentativeEnvironment) moveCandidatesToGOPATH(pkgs ...*exports.Package) (err error) {
func (tempEnv *tentativeEnvironment) moveCandidatesToGOPATH(pkgs ...string) (err error) {
	o := tempEnv.Original
	visited := map[string]bool{}

	for _, pkg := range pkgs {
		//revBefore := o.getRevisionGit(o.PkgDir(pkg.Path))
		//fmt.Printf("rev before: %#v\n", revBefore)
		//dir := tempEnv.PkgDir(pkg.Path)
		dir := tempEnv.PkgDir(pkg)
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
		//target := _repoRoot(o.PkgDir(pkg.Path))
		target := _repoRoot(o.PkgDir(pkg))
		backup := target + fmt.Sprintf("_backup_of_dep_update_%v", time.Now().UnixNano())
		err = os.Rename(target, backup)
		if err != nil {
			// fmt.Printf("can't make backup: %s\n", backup)
			return
			//panic("can't make backup: " + backup)
		}
		err = os.Rename(r, target)
		// fmt.Printf("try to move  %#v to %#v\n", r, target)
		if err != nil {
			// fmt.Printf("can't move  %#v to %#v\n", r, target)
			return
			//panic(err.Error())
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
func (tentative *tentativeEnvironment) updatePackage(pkg string) (conflicts map[string]map[string][3]string, err error) {
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

	tentative.mkdb()
	/*
	   err = createDB(tentative.GOPATH)
	   if err != nil {
	       return
	   }
	*/

	conflicts, err = tentative.checkIntegrity()
	if err != nil {
		// fmt.Println("NO INTEGRITY for TENTATIVE")
		return
	}

	candidates := tentative.getCandidates()

	for _, candidate := range candidates {
		//fmt.Printf("candidate: %v\n", candidate.Path)
		//if tentative.Original.PkgExists(candidate.Path) {
		if tentative.Original.PkgExists(candidate) {
			//	fmt.Println("exists")
			//errs := tentative.checkConflicts(candidate.Path)
			errs := tentative.checkConflicts(candidate)
			if len(errs) > 0 {
				//		fmt.Printf("errors: %v\n", len(errs))
				//conflicts[candidate.Path] = errs
				conflicts[candidate] = errs
			}
		}
	}
	if len(conflicts) == 0 {
		err = tentative.moveCandidatesToGOPATH(candidates...)
	}
	return
}
