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

type Pkg struct {
	Package    string
	JsonMd5    string
	Json       []byte
	ImportsMd5 string
	ExportsMd5 string
	MainMd5    string
	InitMd5    string
}

func UpdatePackage(p *Pkg, i []*Imp, e []*Exp) (err error) {
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

	err = _deleteExports(p.Package, tx)
	if err != nil {
		return
	}

	err = _deleteImports(p.Package, tx)
	if err != nil {
		return
	}

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

func GetImported(importedPkg string) (imps []*Imp, err error) {
	var rows *sql.Rows
	imps = []*Imp{}
	rows, err = db.Query("select package, import, name, value from imports where import = ?", importedPkg)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		i := &Imp{}
		err = rows.Scan(&i.Package, &i.Import, &i.Name, &i.Value)
		if err != nil {
			return
		}
		imps = append(imps, i)
	}
	return
}

func GetPackage(packagePath string) (p *Pkg, err error) {
	var row *sql.Row
	p = &Pkg{}
	row = db.QueryRow("select package, importsmd5, exportsmd5, mainmd5, initmd5, jsonmd5, json from packages where package = ? limit 1", packagePath)
	err = row.Scan(&p.Package, &p.ImportsMd5, &p.ExportsMd5, &p.MainMd5, &p.InitMd5, &p.JsonMd5, &p.Json)
	if err != nil {
		return
	}
	return
}

func insertPackages(p ...*Pkg) (err error) {
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

type Exp struct {
	Package string
	Name    string
	Value   string
}

func _insertExports(tx *sql.Tx, e ...*Exp) (err error) {
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

func _deleteExports(pkgPath string, tx *sql.Tx) (err error) {
	_, err = tx.Exec(`delete from exports where package = ?`, pkgPath)
	return
}

func _deleteImports(pkgPath string, tx *sql.Tx) (err error) {
	_, err = tx.Exec(`delete from imports where package = ?`, pkgPath)
	return
}

func _deletePackage(pkgPath string, tx *sql.Tx) (err error) {
	_, err = tx.Exec(`delete from packages where package = ?`, pkgPath)
	return
}

func deletePackage(pkgPath string) (err error) {
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
	err = _deleteExports(pkgPath, tx)
	if err != nil {
		return
	}
	err = _deleteImports(pkgPath, tx)
	if err != nil {
		return
	}

	err = _deletePackage(pkgPath, tx)
	if err != nil {
		return
	}
	tx.Commit()
	return
}

func insertExports(e ...*Exp) (err error) {
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

type Imp struct {
	Package string
	Import  string
	Name    string
	Value   string
}

func _insertImports(tx *sql.Tx, im ...*Imp) (err error) {
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

func insertImports(im ...*Imp) (err error) {
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
            json            blob not null
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
	// createTables()
	cleanupTables()
	prefill()

	var p *Pkg
	p, err = GetPackage("github.com/metakeule/dep")
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Package: %s\nJson: %s\n", p.Package, p.Json)

	var imps []*Imp
	imps, err = GetImported("github.com/metakeule/dep/packages")
	if err != nil {
		panic(err.Error())
	}

	for _, im := range imps {
		fmt.Printf("%#v\n", im)

	}

	defer db.Close()

}

var packages = []*Pkg{
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
var i *Imp
var imports = []*Imp{
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
