package main

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	sqlite "github.com/mattn/go-sqlite3"
	"github.com/metakeule/dbwrap"
	// "log"
	// "os"
)

type pqdrv int

var drv = &sqlite.SQLiteDriver{}

// fullfill the driver.Driver interface
func (d pqdrv) Open(name string) (driver.Conn, error) { return drv.Open(name) }

var dbFile = "/home/benny/Entwicklung/gopath/src/github.com/metakeule/dep/db/packages.db"
var db *sql.DB

type pkg struct {
	Package    string
	JsonMd5    string
	Json       string
	ImportsMd5 string
	ExportsMd5 string
	MainMd5    string
	InitMd5    string
}

func UpdatePackage(p *pkg, i []*imp, e []*exp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = db.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec(`
        update 
            packages
        set 
            importsmd5 = ?, 
            exportsmd5 = ?, 
            mainmd5 = ?, 
            initmd5 = ?, 
            json = ?, 
            jsonmd5 = ?
        where
            package = ?
        `,
		p.ImportsMd5, p.ExportsMd5, p.MainMd5, p.InitMd5, p.Json, p.JsonMd5, p.Package)
	if err != nil {
		return
	}

	// TODO: remove all old imports and exports

	err = _insertImports(tx, i...)
	if err != nil {
		return
	}
	err = _insertExports(tx, e...)
	if err != nil {
		return
	}

	tx.Commit()
	return
}

func GetImports(importedPkg string) (imps []*imp, err error) {
	var rows *sql.Rows
	imps = []*imp{}
	rows, err = db.Query("select package, import, name, value from imports where import = ?", importedPkg)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		i := &imp{}
		err = rows.Scan(&i.Package, &i.Import, &i.Name, &i.Value)
		if err != nil {
			return
		}
		imps = append(imps, i)
	}
	return
}

func GetPackage(packagePath string) (p *pkg, err error) {
	var row *sql.Row
	p = &pkg{}
	row = db.QueryRow("select package, importsmd5, exportsmd5, mainmd5, initmd5, jsonmd5, json from packages where package = ? limit 1", packagePath)
	err = row.Scan(&p.Package, &p.ImportsMd5, &p.ExportsMd5, &p.MainMd5, &p.InitMd5, &p.JsonMd5, &p.Json)
	if err != nil {
		return
	}
	return
}

func insertPackages(p ...*pkg) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = db.Begin()
	if err != nil {
		return
	}
	stmt, err := tx.Prepare("insert into packages(package, importsmd5, exportsmd5, mainmd5, initmd5, json, jsonmd5) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	for i := 0; i < len(p); i++ {
		_, err = stmt.Exec(p[i].Package, p[i].ImportsMd5, p[i].ExportsMd5, p[i].MainMd5, p[i].InitMd5, p[i].Json, p[i].JsonMd5)
		if err != nil {
			return
		}
	}
	tx.Commit()
	return
}

type exp struct {
	Package string
	Name    string
	Value   string
}

func _insertExports(tx *sql.Tx, e ...*exp) (err error) {
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("insert into exports(package, name, value) values(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	for i := 0; i < len(e); i++ {
		_, err = stmt.Exec(e[i].Package, e[i].Name, e[i].Value)
		if err != nil {
			return
		}
	}
	return
}

func insertExports(e ...*exp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = db.Begin()
	if err != nil {
		return
	}
	err = _insertExports(tx, e...)
	if err != nil {
		return
	}
	tx.Commit()
	return
}

type imp struct {
	Package string
	Import  string
	Name    string
	Value   string
}

func _insertImports(tx *sql.Tx, im ...*imp) (err error) {
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("insert into imports(package, import, name, value) values(?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	for i := 0; i < len(im); i++ {
		_, err = stmt.Exec(im[i].Package, im[i].Import, im[i].Name, im[i].Value)
		if err != nil {
			return err
		}
	}
	return
}

func insertImports(im ...*imp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = db.Begin()
	if err != nil {
		return
	}
	err = _insertImports(tx, im...)
	if err != nil {
		return
	}
	tx.Commit()
	return
}

var lock = make(chan int, 1)

func cleanupTables() {
	var err error
	sqls := []string{
		`delete from packages`,
		`delete from exports`,
		`delete from imports`,
	}
	for _, sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil {
			panic(fmt.Sprintf("%q: %s\n", err, sql))
			//log.Printf("%q: %s\n", err, sql)
			return
		}
	}
}

func createTables() {
	var err error
	sqls := []string{
		"create table foo (id integer not null primary key, name text)",
		`
        create table packages (
            package         text not null primary key,
            importsmd5      text not null,
            exportsmd5      text not null,
            mainmd5         text not null,
            initmd5         text not null,
            jsonmd5         text not null,
            json            text not null
        )`,
		`
        create table exports (
            package         text not null,
            name            text not null,
            value           text not null,
            PRIMARY KEY (package, name),
            FOREIGN KEY(package) REFERENCES packages(package)
        )`,
		`
        create table imports (
            package         text not null,
            import          text not null,
            name            text not null,
            value           text not null,
            PRIMARY KEY (package, import, name),
            FOREIGN KEY(package) REFERENCES packages(package),
            FOREIGN KEY(import) REFERENCES packages(package)
        )
        `,
	}
	for _, sql := range sqls {
		_, err = db.Exec(sql)
		if err != nil {
			panic(fmt.Sprintf("%q: %s\n", err, sql))
			//log.Printf("%q: %s\n", err, sql)
			return
		}
	}

}

func prefill() {
	var err error
	err = insertPackages(packages...)
	if err != nil {
		panic(err.Error())
	}
	err = insertImports(imports...)
	if err != nil {
		panic(err.Error())
	}
}

func run() {
	lock <- 1
	wrap := dbwrap.New("debug", pqdrv(0))

	// use the new name to connect
	var err error
	db, err = sql.Open("debug", dbFile)

	//db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err.Error())
	}
	// called before each method call
	wrap.BeforeAll = func(conn driver.Conn, event string, data ...interface{}) {
		<-lock
		//vals := []interface{}{"before: ", event}
		//vals = append(vals, data...)
		//fmt.Println(vals...)
	}

	wrap.AfterAll = func(conn driver.Conn, event string, data ...interface{}) {
		lock <- 1
		//vals := []interface{}{"after: ", event}
		//vals = append(vals, data...)
		//fmt.Println(vals...)
	}

	cleanupTables()
	prefill()

	var p *pkg
	p, err = GetPackage("github.com/metakeule/dep")
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%#v\n", p)

	var imps []*imp
	imps, err = GetImports("github.com/metakeule/dep/packages")
	if err != nil {
		panic(err.Error())
	}

	for _, im := range imps {
		fmt.Printf("%#v\n", im)

	}

	defer db.Close()

}

/*
    {
   "Path": "github.com/metakeule/dep",
   "Exports": {},
   "UsedImports": {
      "encoding/json#MarshalIndent": "MarshalIndent(interface{},string,string)([]byte,error)",
      "flag#Args": "Args()([]string)",
      "flag#ContinueOnError": "ContinueOnError ErrorHandling",
      "flag#FlagSet": "type FlagSet struct {\n  Usage ()()\n}",
      "flag#NewFlagSet": "NewFlagSet(string,ErrorHandling)(*FlagSet)",
      "flag#Parse": "Parse()()",
      "flag#Usage": "Usage ()()",
      "fmt#Print": "Print(...interface{})(int,error)",
      "fmt#Printf": "Printf(string,...interface{})(int,error)",
      "fmt#Println": "Println(...interface{})(int,error)",
      "github.com/metakeule/dep/packages#Get": "Get(string)(*Package)",
      "github.com/metakeule/dep/packages#PkgPath": "PkgPath(string)(string)",
      "io/ioutil#WriteFile": "WriteFile(string,[]byte,os#FileMode)(error)",
      "os#Exit": "Exit(int)()",
      "path#Join": "Join(...string)(string)",
      "path/filepath#Abs": "Abs(string)(string,error)"
   }
}

*/
var p = &pkg{}

var packages = []*pkg{
	{
		Package: "github.com/metakeule/dep",
		Json: `
{
   "Path": "github.com/metakeule/dep",
   "Exports": {},
   "UsedImports": {
      "github.com/metakeule/dep/packages#Get": "Get(string)(*Package)",
      "github.com/metakeule/dep/packages#PkgPath": "PkgPath(string)(string)",      
   }
}`,
	},
}
var i *imp
var imports = []*imp{
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

type ()

var (
	_ = fmt.Println
)

func init() {

}

func main() {
	run()
}
