package packages

import (
	"encoding/json"
	"regexp"
	"strings"
	// "fmt"
	"github.com/metakeule/exports"
	"go/build"
	"os"
	"path"
	"runtime"
)

var (
	goPATH       = os.Getenv("GOPATH")
	goROOT       = runtime.GOROOT()
	cPath        = path.Join(goPATH, "src", "C")
	PackageCache = map[string]*Package{}
	pkgPath      = path.Join(goPATH, "src")
	rePkgPath    = regexp.MustCompile("^" + regexp.QuoteMeta(pkgPath))
	rootPath     = path.Join(goROOT, "src", "pkg")
	reRootPath   = regexp.MustCompile("^" + regexp.QuoteMeta(rootPath))
)

type packageJson struct {
	Path        string
	Exports     map[string]string
	UsedImports map[string]string
}

type Package struct {
	Path            string
	Internal        bool `json:"-"`
	Imports         map[string]bool
	Exports         map[string]exports.Declaration
	ExternalExports map[string]exports.Declaration
}

func (ø *Package) MarshalJSON() ([]byte, error) {
	p := &packageJson{
		Path:        ø.Path,
		Exports:     map[string]string{},
		UsedImports: map[string]string{},
	}

	for name, decl := range ø.Exports {
		p.Exports[name] = decl.String()
	}

	for name, decl := range ø.ExternalExports {
		p.UsedImports[name] = decl.String()
	}
	return json.Marshal(p)
}

func (ø *Package) ParseExternalExports() {
	expr := exports.GetUsedImports(ø.Path)

	for _, k := range expr {
		//fmt.Printf("expr: %#v\n", k)
		a := strings.Split(k, "#")
		ø.ExternalExports[k] = Get(a[0]).Exports[a[1]]
	}

}

func (ø *Package) ParseExports() {
	ø.Exports = exports.GetExports(ø.Path)
}

func PkgPath(dir string) (s string) {
	if rePkgPath.MatchString(dir) {
		return strings.Replace(dir, pkgPath+"/", "", 1)
	}

	if reRootPath.MatchString(dir) {
		return strings.Replace(dir, rootPath+"/", "", 1)
	}
	panic("not a package: " + dir)
	return
}

func (ø *Package) ParseImports() {
	dir := path.Join(goPATH, "src", ø.Path)
	if dir == cPath {
		return
	}
	pkg, err := build.Default.ImportDir(dir, build.AllowBinary)
	if err != nil {
		dir := path.Join(goROOT, "src", "pkg", ø.Path)
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
	if path == cPath {
		panic("can't get C path: " + path)
	}
	if p, ok := PackageCache[path]; ok {
		return p
	}
	ø = &Package{Path: path}
	ø.Imports = map[string]bool{}
	ø.ExternalExports = map[string]exports.Declaration{}
	PackageCache[path] = ø
	ø.ParseImports()
	ø.ParseExports()
	ø.ParseExternalExports()
	return
}
