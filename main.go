package main

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/exports"
	"go/build"
	//	"io"
	// "go/parser"
	// "go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

/*
Default
*/

var _ = fmt.Print

type allPkgParser struct {
	packages map[string]bool
}

func (ø *allPkgParser) Walker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		pkg, err := build.Default.ImportDir(path, build.ImportMode(0))
		if err == nil && pkg != nil {
			ø.packages[pkg.ImportPath] = true
		}
	}
	return nil
}

func packages() (a []*exports.Package) {
	a = []*exports.Package{}
	prs := &allPkgParser{map[string]bool{}}

	if !RECURSIVE {
		if PACKAGE.Internal {
			//fmt.Printf("skipping internal package %#v\n", PACKAGE.Path)
			return
		}
		a = append(a, PACKAGE)
		return
	}

	err := filepath.Walk(PACKAGE_DIR, prs.Walker)
	if err != nil {
		panic(err.Error())
	}

	for fp, _ := range prs.packages {
		//fmt.Println(fp)
		pkg := exports.DefaultEnv.Pkg(fp)
		if pkg.Internal {
			//fmt.Printf("skipping internal package %#v\n", pkg.Path)
			continue
		}
		a = append(a, pkg)
	}
	return
}

func asJson(pkgs ...*exports.Package) (b []byte) {
	var err error
	b, err = json.MarshalIndent(pkgs, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	// fmt.Printf("%s\n", b)
	return
}

func pkgJson(path string) (b []byte, internal bool) {
	p := exports.DefaultEnv.Pkg(path)
	internal = p.Internal
	var err error
	b, err = json.MarshalIndent(p, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return
}

func scan(dir string) (b []byte, internal bool) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Println(dir)
	b, internal = pkgJson(exports.DefaultEnv.PkgPath(dir))
	b = append(b, []byte("\n")...)
	return
}

// for registry files
var DEP = path.Join(GOPATH, "dep.db")

/*
func readRegisterFile(dir string, internal bool) (*exports.Package, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	trimmedPath := exports.PkgPath(dir)
	// homedir := os.Getenv("HOME")
	registerPath := path.Join(depRegistry, trimmedPath)

	if internal {
		registerPath = path.Join(depRegistryRoot, trimmedPath)
	}

	registerPath, _ = filepath.Abs(registerPath)
	file := path.Join(registerPath, "dep.json")
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	p := &exports.Package{}
	err = json.Unmarshal(b, p)
	return p, err
}

func writeRegisterFile(dir string, data []byte, internal bool) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	trimmedPath := exports.PkgPath(dir)
	// homedir := os.Getenv("HOME")
	registerPath := path.Join(depRegistry, trimmedPath)

	if internal {
		registerPath = path.Join(depRegistryRoot, trimmedPath)
	}

	registerPath, _ = filepath.Abs(registerPath)
	// fmt.Println(registerPath)
	err = os.MkdirAll(registerPath, 0755)
	if err != nil {
		return err
	}
	file := path.Join(registerPath, "dep.json")
	fmt.Printf("writing %s\n", file)
	err = ioutil.WriteFile(file, data, 0644)
	if err != nil {
		return err
	}

	chk := exports.Md5(string(data))
	file = path.Join(registerPath, "dep.md5")
	fmt.Printf("writing %s\n", file)
	return ioutil.WriteFile(file, []byte(chk), 0644)
}
*/
func writeDepFile(dir string, data []byte) error {
	file := path.Join(dir, "dep.json")
	f, _ := filepath.Abs(file)
	fmt.Printf("writing %s\n", f)
	return ioutil.WriteFile(file, data, 0644)
}

func init() {
	if os.Getenv("HOME") == "" {
		panic("HOME environment variable not set")
	}

	if os.Getenv("GOPATH") == "" {
		panic("GOPATH environment variable not set")
	}

	if os.Getenv("GOROOT") == "" {
		panic("GOROOT environment variable not set")
	}
	//Init.Args()
}
