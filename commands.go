package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

var _ = fmt.Printf

func _store(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	pkgs := packages()

	for _, pkg := range pkgs {
		b := asJson(pkg)
		b = append(b, []byte("\n")...)

		dir := path.Join(GOPATH, "src", pkg.Path)
		err := writeDepFile(dir, b)
		if err != nil {
			panic(err.Error())
		}
	}
	return 0
}

func _lint(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	return 0
}

func _show(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	fmt.Printf("%s\n", asJson(packages()...))
	return 0
}

func _registerPackage(pkgMap map[string]*db.Pkg, pkg *exports.Package) (dbExps []*db.Exp, dbImps []*db.Imp) {
	p := &db.Pkg{}
	p.Package = pkg.Path
	p.Json = asJson(pkg)
	pkgMap[pkg.Path] = p
	dbExps = []*db.Exp{}
	dbImps = []*db.Imp{}

	for im, _ := range pkg.Imports {
		if _, has := pkgMap[im]; has {
			continue
		}
		imPkg := exports.DefaultEnv.Pkg(im)
		pExp, pImp := _registerPackage(pkgMap, imPkg)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	pkgjs := pkg.PackageJSON()

	for k, v := range pkgjs.Exports {
		dbE := &db.Exp{}
		dbE.Package = pkg.Path
		dbE.Name = k
		dbE.Value = v
		dbExps = append(dbExps, dbE)
	}

	for k, v := range pkgjs.Imports {
		dbI := &db.Imp{}
		dbI.Package = pkg.Path
		arr := strings.Split(k, "#")
		dbI.Name = arr[1]
		dbI.Value = v
		dbI.Import = arr[0]
		dbImps = append(dbImps, dbI)
	}
	return
}

func registerPackages(pkgs ...*exports.Package) {
	dbExps := []*db.Exp{}
	dbImps := []*db.Imp{}
	pkgMap := map[string]*db.Pkg{}

	for _, pkg := range pkgs {
		pExp, pImp := _registerPackage(pkgMap, pkg)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	dbPkgs := []*db.Pkg{}

	for _, dbPgk := range pkgMap {
		dbPkgs = append(dbPkgs, dbPgk)
	}

	err := db.InsertPackages(dbPkgs, dbExps, dbImps)
	if err != nil {
		panic(err.Error())
	}

	for _, pk := range dbPkgs {
		fmt.Println("registered: ", pk.Package)
	}
}

func _register(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	_, dbFileerr := os.Stat(DEP)

	err := db.Open(DEP)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables()
	}

	pkgs := packages()
	registerPackages(pkgs...)
	return 0
}

func _get(c *cli.Context) ErrorCode {
	return 0
}

func _install(c *cli.Context) ErrorCode {
	return 0
}

type pkgDiff struct {
	Path    string
	Exports []string
	Imports []string
}

func toJson(i interface{}) []byte {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return b
}

func mapDiff(_old map[string]string, _new map[string]string) (diff []string) {
	diff = []string{}
	var visited = map[string]bool{}

	for k, v := range _old {
		visited[k] = true
		vNew, inNew := _new[k]
		if !inNew {
			diff = append(diff, "---"+k+": "+v)
			continue
		}
		if v != vNew {
			diff = append(diff, "---"+k+": "+v)
			diff = append(diff, "+++"+k+": "+vNew)
		}
	}

	for k, v := range _new {
		if !visited[k] {
			diff = append(diff, "+++"+k+": "+v)
		}
	}
	return
}

func _diff(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	_, dbFileerr := os.Stat(DEP)
	err := db.Open(DEP)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables()
	}

	pkgs := packages()

	res := []*pkgDiff{}

	for _, pk := range pkgs {

		dbpkg, exps, imps, e := db.GetPackage(pk.Path, true, true)
		if e != nil {
			panic("package not registered: " + pk.Path)
		}

		js := asJson(pk)

		// TODO: check the hash instead, escp. check the exports and imports hash
		if string(js) != string(dbpkg.Json) {
			//__diff(a, b)
			pkgjs := pk.PackageJSON()

			var oldExports = map[string]string{}

			for _, dbExp := range exps {
				oldExports[dbExp.Name] = dbExp.Value
			}

			pDiff := &pkgDiff{}
			pDiff.Path = pk.Path
			pDiff.Exports = mapDiff(oldExports, pkgjs.Exports)

			var oldImports = map[string]string{}

			for _, dbImp := range imps {
				oldImports[dbImp.Import+"#"+dbImp.Name] = dbImp.Value
			}
			pDiff.Imports = mapDiff(oldImports, pkgjs.Imports)

			if len(pDiff.Exports) > 0 || len(pDiff.Imports) > 0 {
				res = append(res, pDiff)
			}
		}
	}
	if len(res) > 0 {
		fmt.Printf("%s\n", toJson(res))
	}
	return 0
}

func mkdirTempDir() (tmpGoPath string) {
	depPath := path.Join(HOME, ".dep")
	fl, err := os.Stat(depPath)
	if err != nil {
		err = os.MkdirAll(depPath, 0755)
		if err != nil {
			panic(err.Error())
		}
	}
	if !fl.IsDir() {
		panic(depPath + " is a file. but should be a directory")
	}

	tmpGoPath, err = ioutil.TempDir(depPath, "gopath_")
	if err != nil {
		panic(err.Error())
	}
	err = os.Mkdir(path.Join(tmpGoPath, "src"), 0755)
	if err != nil {
		panic(err.Error())
	}
	err = os.Mkdir(path.Join(tmpGoPath, "bin"), 0755)
	if err != nil {
		panic(err.Error())
	}
	err = os.Mkdir(path.Join(tmpGoPath, "pkg"), 0755)
	if err != nil {
		panic(err.Error())
	}
	return
}

/*
	TODO

	- make a tempdir in .deb with subdirs pkg, src and bin
	- set GOPATH to the tempdir and go get the needed package
	- check for package and all their dependancies, if
		- their dependant packages would be fine with the new exports
	- if all is fine, move the missing src entries to the real GOPATH and install the package
	- if there are some conflicts, show them
	- remove -r the tempdir
*/

func hasConflict(p *exports.Package) (errors map[string]error) {
	pkg := p.PackageJSON()
	imp, err := db.GetImported(pkg.Path)
	if err != nil {
		panic(err.Error())
	}
	errors = map[string]error{}

	for _, im := range imp {
		if val, exists := pkg.Exports[im.Name]; exists {
			if val != im.Value {
				errors[im.Name] = fmt.Errorf("%s will change as required from %s (was %s, would be %s", im.Name, im.Package, im.Value, val)
			}
		}
		errors[im.Name] = fmt.Errorf("%s will be missing, required by %s", im.Name, im.Package)
	}
	return
}

func _update(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	tmpDir := mkdirTempDir()

	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			panic(err.Error())
		}
	}()

	pkgs := packages()
	env := exports.NewEnv(runtime.GOROOT(), tmpDir)

	_, dbFileerr := os.Stat(DEP)
	err := db.Open(DEP)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables()
	}

	conflicts := map[string]map[string]error{}
	// TODO: check if the package and its dependancies are installed in the
	// default path, if so, check, they are registered / updated in the database.
	// if not, register /update them
	// TODO make a db connection to get the conflicting
	// packages.
	// it might be necessary to make an update of the db infos first
	for _, pkg := range pkgs {
		args := []string{"get", pkg.Path}
		//args = append(args, c.Args()...)

		cmd := exec.Command("go", args...)
		cmd.Env = []string{
			fmt.Sprintf(`GOPATH='%s'`, tmpDir),
			fmt.Sprintf(`GOROOT='%s'`, GOROOT),
			fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
		}
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout
		err := cmd.Run()
		if err != nil {
			panic(stdout.String() + "\n" + stderr.String())
		}

		tempPkg := env.Pkg(pkg.Path)

		/*
			TODO: check for package and each dependancy,
			if there are depending packages in the default environment
			that do conflict with the exports of the package
		*/
		oldPkg := exports.DefaultEnv.Pkg(pkg.Path)
		if oldPkg == nil {
			panic(fmt.Sprintf("%s is not installed", pkg.Path))
		}
		registerPackages(oldPkg)

		// TODO update entries in DB

		errs := hasConflict(tempPkg)

		if len(errs) > 0 {
			conflicts[pkg.Path] = errs
		}
	}

	if len(conflicts) > 0 {
		b, e := json.Marshal(conflicts)
		if e != nil {
			panic(e.Error())
		}
		fmt.Printf("%s\n", b)
		return UpdateConflict
	}

	// TODO: now check for the package the dependancies
	// therefor we need a second instance (environment) of the exports
	// package, but for another gopath

	return 0
}

func _revisions(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	return 0
}

func _checkout(c *cli.Context) ErrorCode {
	return 0
}

/*
func _fix(c *cli.Context) ErrorCode {
	return 0
}
*/
