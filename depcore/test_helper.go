package depcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "github.com/metakeule/dep/db"
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

func _gogetrevision(env *Environment, pkg string, rev string) {
	err := env.getPackage(pkg)
	if err != nil {
		panic(err.Error())
	}
	dir := env.PkgDir(pkg)
	r := revision{}
	r.VCM = "git"
	r.Rev = rev
	env.checkoutImport(dir, r)
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

func _getWithDeps(env *Environment, pkg string, pkgRev string) (err error) {
	_gogetrevision(env, pkg, pkgRev)
	err = env.checkoutTrackedImports(pkg)
	return
}

func NewEnv(gopath string) (ø *Environment) {
	ø = &Environment{}
	ø.Environment = exports.NewEnv(runtime.GOROOT(), gopath)
	ø.TMPDIR = os.Getenv("DEP_TMP")
	ø.mkdb()
	return
}

func newTestEnv() *Environment {
	return NewEnv(path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "gopath"))
}

func TestCleanup() {
	/*
		opt := &Options{}
		gopath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "gopath")
		opt.GOPATH = gopath
		opt.GOROOT = runtime.GOROOT()
		opt.TMPDIR = os.Getenv("DEP_TMP")
		opt.Env = exports.NewEnv(opt.GOROOT, opt.GOPATH)
	*/
	env := newTestEnv()
	prepare(env.GOPATH)
}

// go get the given revision of the given package
// with its dependancies
func TestGet(pkg, rev string) error {
	env := newTestEnv()
	/*
		opt := &Options{}
		gopath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "gopath")
		opt.GOPATH = gopath
		opt.GOROOT = runtime.GOROOT()
		opt.TMPDIR = os.Getenv("DEP_TMP")
		opt.Env = exports.NewEnv(opt.GOROOT, opt.GOPATH)
		//prepare(gopath)
	*/
	return _getWithDeps(env, pkg, rev)
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
func TestUpdate(pkg, rev string) error {
	env := newTestEnv()
	// opt := &Options{}

	// gopath := path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "gopath")
	//prepare(gopath)

	/*
	   defer func() {
	       os.RemoveAll(gopath)
	   }()
	*/
	/*
		opt.GOPATH = gopath
		opt.GOROOT = runtime.GOROOT()
		opt.TMPDIR = os.Getenv("DEP_TMP")
		opt.Env = exports.NewEnv(opt.GOROOT, opt.GOPATH)
	*/
	err := env.getPackage(pkg)
	if err != nil {
		panic(err.Error())
	}
	dir := path.Join(env.GOPATH, "src", pkg)
	master := getmasterRevision(pkg, dir)
	r := revision{}
	r.VCM = "git"
	r.Rev = rev
	env.checkoutImport(dir, r)
	err = env.checkoutTrackedImports(pkg)
	if err != nil {
		panic(err.Error())
	}

	// check, if revisions are correct
	if env.getRevisionGit(path.Join(env.GOPATH, "src", pkg)) != rev {
		panic(fmt.Sprintf("revision %#v not checked out for package %#v\n", rev, pkg))
	}

	depsBefore, eb := env.trackedImportRevisions(pkg)
	if eb != nil {
		panic(eb.Error())
	}

	for d, drev := range depsBefore {
		if r := env.getRevisionGit(path.Join(env.GOPATH, "src", d)); r != drev.Rev {
			panic(fmt.Sprintf("revision before update %#v not checked out, expected: %#v for dependancy package %#v\n", r, drev.Rev, d))
		}
	}

	//err = createDB(env.GOPATH)
	/*
		if err != nil {
			panic(err.Error())
		}
	*/
	conflicts, e := env.checkIntegrity()
	if e != nil {
		data, _ := json.MarshalIndent(conflicts, "", "  ")
		fmt.Printf("%s\n", data)
		panic(e.Error())
	}

	//env.mkdb()

	//fmt.Println("UPDATING...")
	// updatePackage(o, dB, pkg)
	err = env.DB.updatePackage(pkg)
	if err != nil {
		fmt.Printf("normal error in updating package\n")
		return err
	}

	if r := env.getRevisionGit(env.PkgDir(pkg)); r != master {
		return fmt.Errorf("revision after update %#v not matching master: %#v in package %#v\n", r, master, pkg)
	}
	depsAfter, e := env.trackedImportRevisions(pkg)
	if e != nil {
		panic(e.Error())
	}

	for d, drev := range depsAfter {
		if r := env.getRevisionGit(path.Join(env.GOPATH, "src", d)); r != drev.Rev {
			return fmt.Errorf("revision after update %#v not matching expected: %#v for dependancy package %#v\n", r, drev.Rev, d)
		}
	}

	return nil
}
