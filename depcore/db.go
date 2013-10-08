package depcore

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/go-dep/gdf"
	sqlite "github.com/mattn/go-sqlite3"
	"github.com/metakeule/dbwrap"
)

type pqdrv int

//var DEBUG = false
var VERBOSE bool

type db struct {
	*sql.DB
	File        string
	Opened      bool
	Environment *Environment
}

func (ø *db) Close() error {
	err := ø.DB.Close()
	if err != nil {
		return err
	}
	ø.Opened = false
	return nil
}

func db_open(env *Environment, dbfile string) (ø *db, err error) {
	var d *sql.DB
	d, err = sql.Open("depdb", dbfile)
	ø = &db{d, dbfile, true, env}
	return
}

var drv = &sqlite.SQLiteDriver{}

// fullfill the driver.Driver interface
func (d pqdrv) Open(name string) (driver.Conn, error) {
	//return drv.Open("file:"+name + "?cache=shared&mode=rwc")
	return drv.Open("file:" + name)
}

var lock = make(chan int, 1)

func (ø *db) CleanupTables() {
	var err error
	sqls := []string{
		`delete from packages`,
		`delete from exports`,
		`delete from imports`,
	}
	for _, sql := range sqls {
		_, err = ø.Exec(sql)
		if err != nil {
			panic(fmt.Sprintf("%q: %s\n", err, sql))
			return
		}
	}
}

func (ø *db) NumPackages() (n int) {
	row := ø.QueryRow(`select count(package) from packages`)
	err := row.Scan(&n)
	if err != nil {
		panic(err.Error())
	}
	return
}

func (ø *db) CreateTables() {
	var err error
	sqls := []string{
		`
        create table packages (
            package         text not null primary key,
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
		_, err = ø.Exec(sql)
		if err != nil {
			panic(fmt.Sprintf("%q: %s\n", err, sql))
			return
		}
	}

}

var dBWrapper *dbwrap.Wrapper

func initDB() {
	dBWrapper = dbwrap.New("depdb", pqdrv(0))
	dBWrapper.BeforeAll = func(conn driver.Conn, event string, data ...interface{}) {
		<-lock
		/*
			    	if DEBUG {
						fmt.Println(data...)
					}
		*/
	}

	dBWrapper.AfterAll = func(conn driver.Conn, event string, data ...interface{}) {
		lock <- 1
	}
	lock <- 1
}

func mapkeys(m map[string]string) []string {
	res := []string{}
	for k, _ := range m {
		res = append(res, k)
	}
	return res
}

func (dB *db) hasConflict(pkg *gdf.Package, ignoring map[string]bool) (errors map[string][3]string) {
	errors = map[string][3]string{}
	/*
		if p == nil {
			errors[pkg.Path] = [3]string{"missing", pkg.Path, ""}
			return
		}
	*/
	imp, err := dB.GetImported(pkg.Path)
	if err != nil {
		errors[pkg.Path] = [3]string{"error", err.Error(), ""}
		return
	}
	for _, im := range imp {
		if ignoring[im.Package] {
			continue
		}
		key := fmt.Sprintf("%s:%s", im.Package, im.Name)
		if val, exists := pkg.Exports[im.Name]; exists {
			if val != im.Value {
				errors[key] = [3]string{"changed", im.Value, val}
			}
			continue
		}
		errors[key] = [3]string{"removed", im.Value, ""}
	}
	return
}

func (dB *db) registerPackages(includeImported bool, pkgs ...*gdf.Package) {
	dbExps := []*exp{}
	dbImps := []*imp{}
	pkgMap := map[string]*dbPkg{}

	for _, pkg := range pkgs {
		/*
			if DEBUG {
				fmt.Printf("register package %s\n", pkg.Path)
			}
		*/
		pExp, pImp := dB.Environment.packageToDBFormat(pkgMap, pkg, includeImported)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	dbPkgs := []*dbPkg{}

	for _, dbPgk := range pkgMap {
		dbPkgs = append(dbPkgs, dbPgk)
	}

	err := dB.InsertPackages(dbPkgs, dbExps, dbImps)
	if err != nil {
		panic(err.Error())
	}
}

func (dB *db) updatePackage(pkg string, confirmation func(candidates ...*gdf.Package) bool) (changed map[string][2]string, err error) {
	tentative := dB.Environment.newTentative()
	conflicts, changed, e := tentative.updatePackage(pkg, nil, confirmation)

	if e != nil {
		err = e
		return
	}

	if len(conflicts) > 0 {
		b, e := json.MarshalIndent(conflicts, "", "  ")
		if e != nil {
			panic(e.Error())
		}
		fmt.Printf("%s\n", b)
		err = fmt.Errorf("update conflict")
		return
	}
	return
}

// remove packages, that are in the db but that does not exist anymore
// in the gopath
func (dB *db) removeOrphanedPackages() (candidates map[string]bool, err error) {
	var all []*dbPkg
	all, err = dB.GetAllPackages()

	if err != nil {
		return
	}

	// candidates that should be deleted
	candidates = map[string]bool{}

	for _, p := range all {
		if !dB.Environment.PkgExists(p.Package) {
			candidates[p.Package] = true
		}
	}

	// check for all candidates, if there are still existing packages that depends on them

	// blocking packages: package <key> is imported by package <value> and therefore blocked
	blockingPkgs := map[string]string{}

	for c, _ := range candidates {
		var imps []*imp
		imps, err = dB.GetImported(c)
		if err != nil {
			return
		}
		for _, im := range imps {
			if !candidates[im.Package] {
				blockingPkgs[c] = im.Package
			}
		}
	}

	if len(blockingPkgs) > 0 {
		msg := bytes.NewBufferString("There are packages that still exist and import orphaned packages and therefore block their removal from the registry:\n")
		for a, b := range blockingPkgs {
			msg.WriteString(fmt.Sprintf("\n\t%s is blocked by %s", a, b))
		}
		err = fmt.Errorf(msg.String() + "\n")
		return
	}

	for c, _ := range candidates {
		err = dB.DeletePackage(c)
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

func (ø *db) DeletePackage(pkgPath string) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = ø.Begin()
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
	fmt.Printf("deleted: %s\n", pkgPath)
	return
}

// the the exports of the given package that were imported
// by other packages
func (ø *db) GetImported(importedPkg string) (imps []*imp, err error) {
	var rows *sql.Rows
	imps = []*imp{}
	rows, err = ø.Query("select package, import, name, value from imports where import = ?", importedPkg)
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

func (ø *db) GetPackage(packagePath string, withExports bool, withImports bool) (p *dbPkg, exps []*exp, imps []*imp, err error) {
	var row *sql.Row
	p = &dbPkg{}
	row = ø.QueryRow("select package,  jsonmd5, json from packages where package = ? limit 1", packagePath)
	err = row.Scan(&p.Package, &p.JsonMd5, &p.Json)
	if err != nil {
		return
	}
	if withExports {
		var rows *sql.Rows
		exps = []*exp{}
		rows, err = ø.Query("select package, name, value from exports where package = ?", packagePath)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			e := &exp{}
			err = rows.Scan(&e.Package, &e.Name, &e.Value)
			if err != nil {
				return
			}
			exps = append(exps, e)
		}
	}

	if withImports {
		var rows *sql.Rows
		imps = []*imp{}
		rows, err = ø.Query("select package, import, name, value from imports where package = ?", packagePath)
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
	}
	return
}

func (ø *db) GetAllPackages() (ps []*dbPkg, err error) {
	var rows *sql.Rows
	ps = []*dbPkg{}
	rows, err = ø.Query("select package, jsonmd5, json from packages order by package asc")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		p := &dbPkg{}
		err = rows.Scan(&p.Package, &p.JsonMd5, &p.Json)
		if err != nil {
			return
		}
		ps = append(ps, p)
	}
	return
}

func (ø *db) GetAllImports() (is []*imp, err error) {
	var rows *sql.Rows
	is = []*imp{}
	rows, err = ø.Query("select package, import, name, value from imports")
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
		is = append(is, i)
	}
	return
}

func (ø *db) GetAllExports() (es []*exp, err error) {
	var rows *sql.Rows
	es = []*exp{}
	rows, err = ø.Query("select package, name, value from exports")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		e := &exp{}
		err = rows.Scan(&e.Package, &e.Name, &e.Value)
		if err != nil {
			return
		}
		es = append(es, e)
	}
	return
}

func (ø *db) InsertPackages(p []*dbPkg, e []*exp, im []*imp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = ø.Begin()
	if err != nil {
		return
	}
	stmt, err := tx.Prepare("insert or replace into packages(package, json, jsonmd5) values(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	for i := 0; i < len(p); i++ {
		_, err = stmt.Exec(p[i].Package, p[i].Json, p[i].JsonMd5)
		if err != nil {
			return
		}
		err = _deleteImports(p[i].Package, tx)
		if err != nil {
			return
		}
		err = _deleteExports(p[i].Package, tx)
		if err != nil {
			return
		}
	}

	err = _insertExports(tx, e...)
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

func _insertExports(tx *sql.Tx, e ...*exp) (err error) {
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("insert or replace into exports(package, name, value) values(?, ?, ?)")
	if err != nil {
		fmt.Println("Error in Exports ", err.Error())
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

func (ø *db) InsertExports(e ...*exp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = ø.Begin()
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

func _insertImports(tx *sql.Tx, im ...*imp) (err error) {
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("insert or replace into imports(package, import, name, value) values(?, ?, ?, ?)")
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

func (ø *db) InsertImports(im ...*imp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = ø.Begin()
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

type dbPkg struct {
	Package string
	JsonMd5 string
	Json    []byte
}

type exp struct {
	Package string
	Name    string
	Value   string
}

type imp struct {
	Package string
	Import  string
	Name    string
	Value   string
}

func (ø *db) UpdatePackage(p *dbPkg, i []*imp, e []*exp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = ø.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec(`
        update 
            packages
        set 
            json = ?, 
            jsonmd5 = ?
        where
            package = ?
        `,
		p.Json, p.JsonMd5, p.Package)
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
