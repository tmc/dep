package depcore

import (
	"encoding/json"
	"github.com/metakeule/exports"
	_ "github.com/metakeule/goh4"
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
	env      *exports.Environment
}

func newSubPackages(env *exports.Environment) *subPackages {
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
		pkg, err := ø.env.Build().ImportDir(path, build.ImportMode(0))
		if err == nil && pkg != nil {
			ø.packages[pkg.ImportPath] = true
		}
	}
	return nil
}

// return subpackages of the not internal given package
func SubPackages(pkg *exports.Package) (subs []*exports.Package, err error) {
	subs = []*exports.Package{}
	all := newSubPackages(pkg.Env)
	dir, _ := pkg.Dir()
	err = filepath.Walk(dir, all.Walker)
	if err != nil {
		return
	}
	for pPath, _ := range all.packages {
		subs = append(subs, pkg.Env.Pkg(pPath))
	}
	return
}

/*
func packages(o *Options) (a []*exports.Package) {
	a = []*exports.Package{}
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

func asJson(pkgs ...*exports.Package) (b []byte) {
	var err error
	b, err = json.MarshalIndent(pkgs, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return
}

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
/*
func writeDepFile(dir string, data []byte) error {
	file := path.Join(dir, "dep.json")
	f, _ := filepath.Abs(file)
	fmt.Printf("writing %s\n", f)
	return ioutil.WriteFile(file, data, 0644)
}
*/
