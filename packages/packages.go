package packages

import (
	// "fmt"
	"github.com/metakeule/exports"
	"go/build"
	"os"
	"path"
	"runtime"
)

var (
	GOPATH       = os.Getenv("GOPATH")
	GOROOT       = runtime.GOROOT()
	C_Path       = path.Join(GOPATH, "src", "C")
	PackageCache = map[string]*Package{}
)

type Package struct {
	Path            string
	Internal        bool `json:"-"`
	Imports         map[string]bool
	Exports         map[string]interface{}
	ExternalExports map[string]interface{}
}

func (ø *Package) ParseExternalExports() {
	expr := exports.SelectorExpressions(ø.Path)

	for k, _ := range expr {
		//fmt.Printf("expr: %#v\n", k)
		ø.ExternalExports[k[0]+"#"+k[1]] = Get(k[0]).Exports[k[1]]
	}

}

func (ø *Package) ParseExports() {
	ø.Exports = exports.Exports(ø.Path)
}

func (ø *Package) ParseImports() {
	dir := path.Join(GOPATH, "src", ø.Path)
	if dir == C_Path {
		return
	}
	pkg, err := build.Default.ImportDir(dir, build.AllowBinary)
	if err != nil {
		dir := path.Join(GOROOT, "src", "pkg", ø.Path)
		pkg, err = build.Default.ImportDir(dir, build.AllowBinary)
		if err != nil {
			panic(err.Error())
		}
		ø.Internal = true
		return
	}

	for _, imp := range pkg.Imports {
		imPort := Get(imp)
		if err != nil {
			panic("pkg " + ø.Path + " imports " + imp + " with error " + err.Error())
		}
		if !imPort.Internal {
			ø.Imports[imp] = true
		}
	}
	return
}

func Get(path string) (ø *Package) {
	if path == C_Path {
		panic("can't get C path: " + path)
	}
	if p, ok := PackageCache[path]; ok {
		return p
	}
	ø = &Package{Path: path}
	ø.Imports = map[string]bool{}
	ø.ExternalExports = map[string]interface{}{}
	PackageCache[path] = ø
	ø.ParseImports()
	ø.ParseExports()
	ø.ParseExternalExports()
	return
}
