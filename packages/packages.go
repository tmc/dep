package packages

import (
	"github.com/metakeule/exports"
	"go/build"
	"os"
	"path"
	"runtime"
)

var (
	GOPATH   = os.Getenv("GOPATH")
	GOROOT   = runtime.GOROOT()
	PkgCache = map[string]*Pkg{}
)

type Pkg struct {
	Path     string
	Internal bool `json:"-"`
	Imports  map[string]bool
	Exports  map[string]interface{}
}

func (ø *Pkg) ParseExports() {
	ø.Exports = exports.Exports(ø.Path)
}

func (ø *Pkg) ParseImports() {
	dir := path.Join(GOPATH, "src", ø.Path)
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

func Get(path string) (ø *Pkg) {
	if p, ok := PkgCache[path]; ok {
		return p
	}
	ø = &Pkg{Path: path}
	ø.Imports = map[string]bool{}
	PkgCache[path] = ø
	ø.ParseImports()
	ø.ParseExports()
	return
}
