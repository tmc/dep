package depcore

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type testEnv struct{ inner *Environment }

func NewTestEnv() *testEnv {
	t := &testEnv{
		NewEnv(
			path.Join(
				os.Getenv("GOPATH"),
				"src",
				"github.com",
				"metakeule",
				"dep",
				"gopath"))}
	t.prepare()
	return t
}

func (env *testEnv) prepare() {
	os.RemoveAll(env.inner.GOPATH)
	os.MkdirAll(env.inner.GOPATH, 0755)
	env.inner.Open()
}

func (env *testEnv) Get(pkg, rev string) error {
	return env.inner.getWithImports(pkg, rev)
}

func (ev *testEnv) Update(pkg, rev string) error {
	defer ev.inner.Close()
	env := ev.inner
	err := env.getPackage(pkg)
	if err != nil {
		panic(err.Error())
	}
	dir := env.PkgDir(pkg)
	master := getmasterRevision(pkg, dir)
	env.checkoutImport(dir, revision{VCM: "git", Rev: rev})
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

	conflicts, e := env.checkIntegrity()
	if e != nil {
		data, _ := json.MarshalIndent(conflicts, "", "  ")
		fmt.Printf("%s\n", data)
		panic(e.Error())
	}

	err = env.DB.updatePackage(pkg)
	if err != nil {
		//		fmt.Printf("normal error in updating package\n")
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
