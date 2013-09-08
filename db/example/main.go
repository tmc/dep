package main

import (
	"fmt"
	"github.com/metakeule/dep/db"
)

var dbFile = "/home/benny/Entwicklung/gopath/src/github.com/metakeule/dep/db/packages.db"

func prefill() {
	var err error
	err = db.InsertPackages(packages...)
	if err != nil {
		panic(err.Error())
	}
	err = db.InsertImports(imports...)
	if err != nil {
		panic(err.Error())
	}
}

func run() {
	//lock <- 1

	// use the new name to connect
	err := db.Open(dbFile)
	//db, err = sql.Open("debug", dbFile)

	//db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err.Error())
	}
	// called before each method call

	// createTables()
	db.CleanupTables()
	prefill()

	var p *db.Pkg
	p, err = db.GetPackage("github.com/metakeule/dep")
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Package: %s\nJson: %s\n", p.Package, p.Json)

	var imps []*db.Imp
	imps, err = db.GetImported("github.com/metakeule/dep/packages")
	if err != nil {
		panic(err.Error())
	}

	for _, im := range imps {
		fmt.Printf("%#v\n", im)

	}

	defer db.Close()

}

var packages = []*db.Pkg{
	{
		Package: "github.com/metakeule/dep",
		Json: []byte(`
{
   "Path": "github.com/metakeule/dep",
   "Exports": {},
   "UsedImports": {
      "github.com/metakeule/dep/packages#Get": "Get(string)(*Package)",
      "github.com/metakeule/dep/packages#PkgPath": "PkgPath(string)(string)",      
   }
}`),
	},
}
var imports = []*db.Imp{
	{
		Import:  "github.com/metakeule/dep/packages",
		Package: "github.com/metakeule/dep",
		Name:    "Get",
		Value:   "Get(string)(*Package)",
	},
	{
		Import:  "github.com/metakeule/dep/packages",
		Package: "github.com/metakeule/dep",
		Name:    "PkgPath",
		Value:   "PkgPath(string)(string)",
	},
}

var (
	_ = fmt.Println
)

func main() {
	run()
}
