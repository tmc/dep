package main

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	"os"
	"path"
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
		imPkg := exports.Get(im)
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

func _register(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	// db.DEBUG = true
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

	err = db.InsertPackages(dbPkgs, dbExps, dbImps)
	if err != nil {
		panic(err.Error())
	}

	for _, pk := range dbPkgs {
		fmt.Println("inserted/replaced: ", pk.Package)
	}
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
			panic(e.Error())
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

func _update(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
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
