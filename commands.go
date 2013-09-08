package main

import (
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

func _register(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	//fmt.Println(DEP)
	//return 0
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
		p := &db.Pkg{}
		p.Package = pkg.Path
		p.Json = asJson(pkg)
		//dbPkgs = append(dbPkgs, p)
		pkgMap[pkg.Path] = p

		// TODO make a recursiv function to track the exports
		// and imports from the imports and their imports too
		for im, _ := range pkg.Imports {
			// fmt.Printf(im)
			if _, has := pkgMap[im]; !has {
				continue
			}

			imPkg := exports.Get(im)
			ip := &db.Pkg{}
			ip.Package = imPkg.Path
			ip.Json = asJson(imPkg)
			pkgMap[imPkg.Path] = ip
		}

		for k, v := range pkg.Exports {
			dbE := &db.Exp{}
			dbE.Package = pkg.Path
			dbE.Name = k
			dbE.Value = v.String()
			dbExps = append(dbExps, dbE)
		}

		for k, v := range pkg.ExternalExports {
			dbI := &db.Imp{}
			dbI.Package = pkg.Path
			arr := strings.Split(k, "#")
			dbI.Name = arr[1]
			dbI.Value = v.String()
			dbI.Import = arr[0]
			dbImps = append(dbImps, dbI)
		}
	}

	dbPkgs := []*db.Pkg{}

	for _, dbPgk := range pkgMap {
		dbPkgs = append(dbPkgs, dbPgk)
	}

	// TODO: check first, which packages are allready registered.
	// For the ones that are registered, do an update instead

	db.CleanupTables()

	// TODO: make it all in one large transaction
	err = db.InsertPackages(dbPkgs, dbExps, dbImps)
	if err != nil {
		panic(err.Error())
	}
	all, _ := db.GetAllPackages()

	for _, pk := range all {
		fmt.Printf("%s\n-----\n", pk.Json)
	}

	imps, _ := db.GetAllImports()

	for _, impp := range imps {
		fmt.Printf("Package: %s\nImport: %s\nName: %s\nValue: %s\n-----\n", impp.Package, impp.Import, impp.Name, impp.Value)
	}

	exps, _ := db.GetAllExports()

	for _, expp := range exps {
		fmt.Printf("Package: %s\nName: %s\nValue: %s\n-----\n", expp.Package, expp.Name, expp.Value)
	}
	/*
		dir := fs.Arg(0)
		b, internal := scan(dir)
		err := writeRegisterFile(dir, b, internal)
		if err != nil {
			panic(err.Error())
		}
	*/
	return 0
}

func _get(c *cli.Context) ErrorCode {
	return 0
}

func _install(c *cli.Context) ErrorCode {
	return 0
}

func _diff(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
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
