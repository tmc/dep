package dep

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/metakeule/exports"
	_ "github.com/metakeule/goh4"
	"go/build"
	"io"
	"strings"
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

//var PACKAGE *exports.Package

var _ = fmt.Print

func makeBuildContext(gopath string) build.Context {
	ctx := build.Context{}
	ctx.GOARCH = build.Default.GOARCH
	ctx.GOOS = build.Default.GOOS
	ctx.GOROOT = build.Default.GOROOT
	ctx.GOPATH = gopath
	ctx.CgoEnabled = build.Default.CgoEnabled
	ctx.UseAllFiles = build.Default.UseAllFiles
	ctx.Compiler = build.Default.Compiler
	ctx.BuildTags = build.Default.BuildTags
	ctx.ReleaseTags = build.Default.ReleaseTags
	ctx.InstallSuffix = build.Default.InstallSuffix
	ctx.JoinPath = build.Default.JoinPath
	ctx.SplitPathList = build.Default.SplitPathList
	ctx.IsAbsPath = build.Default.IsAbsPath
	ctx.IsDir = build.Default.IsDir
	ctx.HasSubdir = build.Default.HasSubdir
	ctx.ReadDir = build.Default.ReadDir
	ctx.OpenFile = build.Default.OpenFile
	return ctx
}

type allPkgParser struct {
	packages map[string]bool
	context  build.Context
}

func newAllPkgParser(gopath string) *allPkgParser {
	return &allPkgParser{
		packages: map[string]bool{},
		context:  makeBuildContext(gopath),
	}
}

func (ø *allPkgParser) Walker(path string, info os.FileInfo, err error) error {

	if err != nil {
		return err
	}
	if info.IsDir() {
		pkg, err := ø.context.ImportDir(path, build.ImportMode(0))
		if err == nil && pkg != nil {
			ø.packages[pkg.ImportPath] = true
		}
	}
	return nil
}

/* stolen from go1.1/src/cmd/go/main.go */

// envForDir returns a copy of the environment
// suitable for running in the given directory.
// The environment is the current process's environment
// but with an updated $PWD, so that an os.Getwd in the
// child will be faster.
func envForDir(dir string) []string {
	env := os.Environ()
	// Internally we only use rooted paths, so dir is rooted.
	// Even if dir is not rooted, no harm done.
	return mergeEnvLists([]string{"PWD=" + dir}, env)
}

// mergeEnvLists merges the two environment lists such that
// variables with the same name in "in" replace those in "out".
func mergeEnvLists(in, out []string) []string {
NextVar:
	for _, inkv := range in {
		k := strings.SplitAfterN(inkv, "=", 2)[0]
		for i, outkv := range out {
			if strings.HasPrefix(outkv, k) {
				out[i] = inkv
				continue NextVar
			}
		}
		out = append(out, inkv)
	}
	return out
}

var errHTTP = errors.New("no http in bootstrap go command")

func httpGET(url string) ([]byte, error) {
	return nil, errHTTP
}

func httpsOrHTTP(importPath string) (string, io.ReadCloser, error) {
	return "", nil, errHTTP
}

func parseMetaGoImports(r io.Reader) (imports []metaImport) {
	panic("unreachable")
}

/* End Of stolen from go1.1/src/cmd/go/main.go */

func packages(o *Options) (a []*exports.Package) {
	a = []*exports.Package{}
	//prs := &allPkgParser{map[string]bool{}}
	prs := newAllPkgParser(o.GOPATH)

	if !o.Recursive {
		if o.Package.Internal {
			//fmt.Printf("skipping internal package %#v\n", PACKAGE.Path)
			return
		}
		a = append(a, o.Package)
		return
	}

	err := filepath.Walk(o.PackageDir, prs.Walker)
	if err != nil {
		panic(err.Error())
	}

	for fp, _ := range prs.packages {
		//fmt.Println(fp)
		pkg := o.Env.Pkg(fp)
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

func pkgJson(o *Options, path string) (b []byte, internal bool) {
	p := o.Env.Pkg(path)
	internal = p.Internal
	var err error
	b, err = json.MarshalIndent(p, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return
}

func scan(o *Options, dir string) (b []byte, internal bool) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Println(dir)
	b, internal = pkgJson(o, o.Env.PkgPath(dir))
	b = append(b, []byte("\n")...)
	return
}

// for registry files
//var DEP = path.Join(GOPATH, "dep.db")

func DEP(gopath string) string {
	return path.Join(gopath, DEP_DB)
}

var DEP_DB = "dep.db"

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
