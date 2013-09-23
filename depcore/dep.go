package depcore

import (
	"encoding/json"
	"github.com/metakeule/gdf"
	"go/build"
	//	"io"
	// "go/parser"
	// "go/token"
	// "io/ioutil"
	"os"
	// "path"
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
		//ø.env.Build()
		/*
			base := filepath.Base(path)

			if base == "example" || base == "examples" {
				//	fmt.Printf("pkg base: %s\n", base)
				return filepath.SkipDir
			}
		*/
		//pkg, err := ø.env.Build().ImportDir(path, build.ImportMode(0))

		//pPath := ø.env.PkgPath(path)
		pkg, err := ø.env.Build().ImportDir(path, build.ImportMode(0))
		//pkg, err := ø.env.Build().Import(pPath, ø.env.GOPATH , build.ImportMode(0))
		if err == nil && pkg != nil {
			//	fmt.Printf("found package %s\n", pkg.ImportPath)
			ø.packages[pkg.ImportPath] = true
		}
	}
	return nil
}

// return subpackages of the not internal given package
/*
func SubPackages(pkg *gdf.Package) (subs []*gdf.Package, err error) {
	subs = []*gdf.Package{}
	all := newSubPackages(pkg.Env)
	//dir, _ := pkg.Dir()
	var dir string
	var internal bool
	dir, internal, err = pkg.Env.PkgDir(pkg.Path)
	if err != nil {
		return
	}
	if internal {
		return
	}
	err = filepath.Walk(dir, all.Walker)
	if err != nil {
		return
	}
	for pPath, _ := range all.packages {
		var p *gdf.Package
		p, err = pkg.Env.Pkg(pPath)
		if err != nil {
			return
		}
		subs = append(subs, p)
	}
	return
}
*/

/*
func packages(o *Options) (a []*gdf.Package) {
	a = []*gdf.Package{}
	//prs := &allPkgParser{map[string]bool{}}
	//prs := newAllPkgParser(o.GOPATH)
	prs := newAllPkgParser(o.Env)

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
*/

func asJson(pkgs ...*gdf.Package) (b []byte) {
	var err error
	b, err = json.MarshalIndent(pkgs, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return
}

/*
func readRegisterFile(dir string, internal bool) (*gdf.Package, error) {
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
/*
func writeDepFile(dir string, data []byte) error {
	file := path.Join(dir, "dep.json")
	f, _ := filepath.Abs(file)
	fmt.Printf("writing %s\n", f)
	return ioutil.WriteFile(file, data, 0644)
}
*/
