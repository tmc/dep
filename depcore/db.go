package depcore

import (
	"encoding/json"
	"fmt"

	// "github.com/metakeule/dep/db"
	"database/sql"
	"database/sql/driver"
	sqlite "github.com/mattn/go-sqlite3"
	"github.com/metakeule/dbwrap"
	"github.com/metakeule/gdf"
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
			//log.Printf("%q: %s\n", err, sql)
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
	// fmt.Printf("CREATE TABLES FOR %s\n", db.File)
	var err error
	sqls := []string{
		`
        create table packages (
            package         text not null primary key,
            importsmd5      text not null,
            exportsmd5      text not null,
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
		_, err = ø.Exec(sql)
		if err != nil {
			panic(fmt.Sprintf("%q: %s\n", err, sql))
			//log.Printf("%q: %s\n", err, sql)
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

/*
   TODO
   - check for all packages within the repoDir
   - check for package and all their dependancies, if
       - their dependant packages would be fine with the new exports
   - if all is fine, move the missing src entries to the real GOPATH and install the package
   - if there are some conflicts, show them
*/

// TODO: add verbose flag for verbose output
// ignore the given dependencies
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
		//fmt.Printf("package %s \n\timports %s of %s, \n\tbut that has \n\t%s\n\n", im.Package, im.Name, pkg.Path, strings.Join(mapkeys(pkg.Exports), "\n\t"))
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

func (dB *db) updatePackage(pkg string, confirmation func(candidates ...*gdf.Package) bool) error {
	tentative := dB.Environment.newTentative()
	conflicts, err := tentative.updatePackage(pkg, confirmation)

	if err != nil {
		return err
	}

	if len(conflicts) > 0 {
		b, e := json.MarshalIndent(conflicts, "", "  ")
		if e != nil {
			panic(e.Error())
		}
		fmt.Printf("%s\n", b)
		return fmt.Errorf("update conflict")
	}
	return nil
}
