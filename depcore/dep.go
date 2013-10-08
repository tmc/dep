package depcore

import (
	"encoding/json"
	"github.com/go-dep/gdf"
	"go/build"
	"os"
	"path/filepath"
)

type subPackages struct {
	packages map[string]bool
	env      environmental
}

type environmental interface {
	shouldIgnorePkg(string) bool
	Build() *build.Context
	PkgPath(string) string
}

func newSubPackages(env environmental) *subPackages {
	return &subPackages{
		packages: map[string]bool{},
		env:      env,
	}
}

func (ø *subPackages) Walker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		pPath := ø.env.PkgPath(path)
		if ø.env.shouldIgnorePkg(pPath) {
			return filepath.SkipDir
		}
		pkg, err := ø.env.Build().ImportDir(path, build.ImportMode(0))
		if err == nil && pkg != nil {
			ø.packages[pkg.ImportPath] = true
		}
	}
	return nil
}

func niceJson(pkgs ...*gdf.Package) (b []byte) {
	var err error
	b, err = json.MarshalIndent(pkgs, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return
}
