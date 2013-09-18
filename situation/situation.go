package situation

import (
	"bytes"
	"fmt"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/dep/dep"
	"github.com/metakeule/exports"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

func prepare(gopath string) {
	os.RemoveAll(gopath)
	err := os.MkdirAll(path.Join(gopath, "src"), 0755)
	if err != nil {
		panic(err.Error())
	}
	err = os.MkdirAll(path.Join(gopath, "pkg"), 0755)
	if err != nil {
		panic(err.Error())
	}
	err = os.MkdirAll(path.Join(gopath, "bin"), 0755)
	if err != nil {
		panic(err.Error())
	}
}

func _gogetrevision(o *dep.Options, pkg string, rev string) {
	dep.GoGetPackages(o, o.GOPATH, pkg)
	dir := path.Join(o.GOPATH, "src", pkg)
	r := dep.Revision{}
	r.VCM = "git"
	r.Rev = rev
	dep.CheckoutRevision(o, dir, r)
}

func getmasterRevision(pkg string, dir string) string {
	cmd := exec.Command("git", "rev-parse", "master")
	cmd.Env = []string{
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}
	cmd.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		panic(stdout.String() + "\n" + stderr.String())
	}
	return strings.Trim(stdout.String(), "\n\r")
}

func _getWithDeps(o *dep.Options, pkg string, pkgRev string) (err error) {
	_gogetrevision(o, pkg, pkgRev)
	err = dep.CheckoutDependanciesByRevFile(o, o.GOPATH, pkg)
	return
}

func CleanEnv() {
	opt := &dep.Options{}
	gopath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "situation", "gopath")
	opt.GOPATH = gopath
	opt.GOROOT = runtime.GOROOT()
	opt.HOME = os.Getenv("HOME")
	opt.Env = exports.NewEnv(opt.GOROOT, opt.GOPATH)
	prepare(gopath)
}

// go get the given revision of the given package
// with its dependancies
func GetPackage(pkg, rev string) error {
	opt := &dep.Options{}
	gopath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "situation", "gopath")
	opt.GOPATH = gopath
	opt.GOROOT = runtime.GOROOT()
	opt.HOME = os.Getenv("HOME")
	opt.Env = exports.NewEnv(opt.GOROOT, opt.GOPATH)
	//prepare(gopath)
	return _getWithDeps(opt, pkg, rev)
}

/*
   panic for the following
   - goget error
   - switch revision error
   - create db error
   - HOME is not set

   return different errors for
   - integrity error
   - update error
*/

// rev is the revision of the package before the update
// the package will be updated to master
// the depencies should be tracked for the initial rev of pkg
// and for the master with 'dep revisions'. Before the update, the revisions of the
// dependancies are checkout as defined in the dep-rev.json of the initial rev of pkg
// After the update, it will be checked, if
// the target revisions as defined in the dep-rev.json of the master of pkg have been updated
func Update(pkg, rev string) error {
	opt := &dep.Options{}
	gopath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "situation", "gopath")
	//prepare(gopath)

	/*
		defer func() {
			os.RemoveAll(gopath)
		}()
	*/

	opt.GOPATH = gopath
	opt.GOROOT = runtime.GOROOT()
	opt.HOME = os.Getenv("HOME")
	opt.Env = exports.NewEnv(opt.GOROOT, opt.GOPATH)

	dep.GoGetPackages(opt, opt.GOPATH, pkg)
	dir := path.Join(opt.GOPATH, "src", pkg)
	master := getmasterRevision(pkg, dir)
	r := dep.Revision{}
	r.VCM = "git"
	r.Rev = rev
	dep.CheckoutRevision(opt, dir, r)
	err := dep.CheckoutDependanciesByRevFile(opt, opt.GOPATH, pkg)
	if err != nil {
		panic(err.Error())
	}

	// check, if revisions are correct
	if dep.GetRevisionGit(opt, path.Join(opt.GOPATH, "src", pkg)) != rev {
		panic(fmt.Sprintf("revision %#v not checked out for package %#v\n", rev, pkg))
	}

	depsBefore, eb := dep.GetDependancyRevisions(opt.GOPATH, pkg)
	if eb != nil {
		panic(eb.Error())
	}

	for d, drev := range depsBefore {
		if r := dep.GetRevisionGit(opt, path.Join(opt.GOPATH, "src", d)); r != drev.Rev {
			panic(fmt.Sprintf("revision before update %#v not checked out, expected: %#v for dependancy package %#v\n", r, drev.Rev, d))
		}
	}

	err = dep.CreateDB(opt.GOPATH)

	if err != nil {
		panic(err.Error())
	}

	err = dep.CheckIntegrity(opt, opt.Env)
	if err != nil {
		panic(err.Error())
	}

	var dB *db.DB

	dB, err = db.Open(dep.DEP(opt.GOPATH))
	if err != nil {
		panic(err.Error())
	}
	defer dB.Close()
	//fmt.Println("UPDATING...")
	// updatePackage(o, dB, pkg)
	err = dep.UpdatePackage(opt, dB, pkg)
	if err != nil {
		//	fmt.Printf("normal error in updating package\n")
		return err
	}

	if r := dep.GetRevisionGit(opt, path.Join(opt.GOPATH, "src", pkg)); r != master {
		return fmt.Errorf("revision after update %#v not matching master: %#v in package %#v\n", r, master, pkg)
	}
	depsAfter, e := dep.GetDependancyRevisions(opt.GOPATH, pkg)
	if e != nil {
		panic(e.Error())
	}

	for d, drev := range depsAfter {
		if r := dep.GetRevisionGit(opt, path.Join(opt.GOPATH, "src", d)); r != drev.Rev {
			return fmt.Errorf("revision after update %#v not matching expected: %#v for dependancy package %#v\n", r, drev.Rev, d)
		}
	}

	return nil
}
