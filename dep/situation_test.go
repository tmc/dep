package dep

import (
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	"os"
	"path"
	"runtime"
	"testing"
)

var opt = &Options{}
var pkgRevAfter = "a85b1db977f9bf5335b0ce99b99935e18c130de6"
var depRevBefore = "4976ff78ef93be4c04774ba18bffdbf2a96c5cc0"
var depRevAfter = "54daa3277ea0a5980b104a80679fffdb1a195d90"

func init() {
	opt.GOPATH = path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "gopath")
	opt.GOROOT = runtime.GOROOT()
	opt.HOME = os.Getenv("HOME")
	opt.Env = exports.NewEnv(opt.GOROOT, opt.GOPATH)
}

func prepare() {
	os.RemoveAll(opt.GOPATH)
	var err error
	err = os.MkdirAll(path.Join(opt.GOPATH, "src"), 0755)
	if err != nil {
		panic(err.Error())
	}
	err = os.MkdirAll(path.Join(opt.GOPATH, "pkg"), 0755)
	if err != nil {
		panic(err.Error())
	}
	err = os.MkdirAll(path.Join(opt.GOPATH, "bin"), 0755)
	if err != nil {
		panic(err.Error())
	}
	//	goget(opt, "github.com/metakeule/deptest_situation/working/compatible_change")
}

func _gogetrevision(o *Options, pkg string, rev string) {
	goGetPackages(o, o.GOPATH, pkg)
	dir := path.Join(opt.GOPATH, "src", pkg)
	r := revision{}
	r.VCM = "git"
	r.Rev = rev
	checkoutRevision(opt, dir, r)
}

func _getWithDeps(o *Options, pkg string, pkgRev string) (err error) {
	_gogetrevision(o, pkg, pkgRev)
	err = checkoutDependanciesByRevFile(o, o.GOPATH, pkg)
	return
}

func TestCompatible(t *testing.T) {
	prepare()
	pkg := "github.com/metakeule/deptest_situation/working/compatible_change"
	pkgRevBefore := "d5cf87993ac713a5aef12e39abf08542f3187587"

	err := _getWithDeps(opt, pkg, pkgRevBefore)
	if err != nil {
		t.Fatal(err.Error())
	}

	// check, if revisions are correct

	if getRevisionGit(opt, path.Join(opt.GOPATH, "src", pkg)) != pkgRevBefore {
		t.Fatalf("revision %#v not checked out for package %#v\n", pkgRevBefore, pkg)
	}

	depPkg := "github.com/metakeule/deptest_mod_a/compatible"

	if getRevisionGit(opt, path.Join(opt.GOPATH, "src", depPkg)) != depRevBefore {
		t.Fatalf("revision %#v not checked out for package %#v\n", depRevBefore, depPkg)
	}

	err = createDB(opt.GOPATH)

	if err != nil {
		t.Fatal(err.Error())
	}

	checkIntegrity(opt, opt.Env)

	dB, err := db.Open(DEP(opt.GOPATH))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer dB.Close()
	// updatePackage(o, dB, pkg)
	err = updatePackage(opt, dB, pkg)
	if err != nil {
		t.Error(err.Error())
	}

	if r := getRevisionGit(opt, path.Join(opt.GOPATH, "src", pkg)); r != pkgRevAfter {
		t.Errorf("revision after update %#v not matching expected: %#v in package %#v\n", r, pkgRevAfter, pkg)
	}

	if r := getRevisionGit(opt, path.Join(opt.GOPATH, "src", depPkg)); r != depRevAfter {
		t.Errorf("revision after update %#v not matching expected: %#v for dependancy package %#v\n", r, depRevAfter, depPkg)
	}

}

// test change where the only a not used part changes in a compatible way
func TestCompatiblePartialChange(t *testing.T) {
	prepare()
	pkg := "github.com/metakeule/deptest_situation/working/partial_change"
	pkgRevBefore := "e1c61521f503ecd370642e95f71284bad2c366a1"

	err := _getWithDeps(opt, pkg, pkgRevBefore)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// check, if revisions are correct

	if getRevisionGit(opt, path.Join(opt.GOPATH, "src", pkg)) != pkgRevBefore {
		t.Fatalf("revision %#v not checked out for package %#v\n", pkgRevBefore, pkg)
	}

	depPkg := "github.com/metakeule/deptest_mod_a/partial_broken"

	if getRevisionGit(opt, path.Join(opt.GOPATH, "src", depPkg)) != depRevBefore {
		t.Fatalf("revision %#v not checked out for package %#v\n", depRevBefore, depPkg)
	}

	err = createDB(opt.GOPATH)

	if err != nil {
		t.Fatal(err.Error())
	}

	checkIntegrity(opt, opt.Env)

	dB, err := db.Open(DEP(opt.GOPATH))
	if err != nil {
		t.Fatal(err.Error())
	}
	defer dB.Close()
	// updatePackage(o, dB, pkg)
	err = updatePackage(opt, dB, pkg)
	if err != nil {
		t.Error(err.Error())
	}
	/*
		if getRevisionGit(opt, path.Join(opt.GOPATH, "src", pkg)) != pkgRevAfter {
			t.Errorf("revision after update %#v not matching package %#v\n", pkgRevAfter, pkg)
		}

		if getRevisionGit(opt, path.Join(opt.GOPATH, "src", depPkg)) != depRevAfter {
			t.Errorf("revision after update %#v not matching for dependancy package %#v\n", depRevAfter, depPkg)
		}
	*/
}

func TestInCompatible(t *testing.T) {
	prepare()
	pkg := "github.com/metakeule/deptest_situation/failing/incompatible_change"
	pkgRevBefore := "d1c7c9a17887cefb8892c8edeafab927da8ce5e8"

	err := _getWithDeps(opt, pkg, pkgRevBefore)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// check, if revisions are correct

	if getRevisionGit(opt, path.Join(opt.GOPATH, "src", pkg)) != pkgRevBefore {
		t.Fatalf("revision %#v not checked out for package %#v\n", pkgRevBefore, pkg)
	}

	depPkg := "github.com/metakeule/deptest_mod_a/broken"

	if getRevisionGit(opt, path.Join(opt.GOPATH, "src", depPkg)) != depRevBefore {
		t.Fatalf("revision %#v not checked out for package %#v\n", depRevBefore, depPkg)
	}

	err = createDB(opt.GOPATH)

	if err != nil {
		t.Fatalf(err.Error())
	}

	checkIntegrity(opt, opt.Env)

	dB, err := db.Open(DEP(opt.GOPATH))
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer dB.Close()
	// updatePackage(o, dB, pkg)
	err = updatePackage(opt, dB, pkg)
	if err == nil {
		t.Errorf("update of a package that depends on a package that broke compatibilities should raise an error but does not")
	}
}
