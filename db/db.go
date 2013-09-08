package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	sqlite "github.com/mattn/go-sqlite3"
	"github.com/metakeule/dbwrap"
)

type pqdrv int

var db *sql.DB

func Close() error {
	return db.Close()
}

func Open(dbfile string) (err error) {
	db, err = sql.Open("depdb", dbfile)
	return
}

var drv = &sqlite.SQLiteDriver{}

// fullfill the driver.Driver interface
func (d pqdrv) Open(name string) (driver.Conn, error) { return drv.Open(name) }

var lock = make(chan int, 1)

func CleanupTables() {
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

func CreateTables() {
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

var DBWrapper *dbwrap.Wrapper

func init() {
	DBWrapper = dbwrap.New("depdb", pqdrv(0))
	DBWrapper.BeforeAll = func(conn driver.Conn, event string, data ...interface{}) {
		<-lock
		fmt.Println(data...)
	}

	DBWrapper.AfterAll = func(conn driver.Conn, event string, data ...interface{}) {
		lock <- 1
	}
	lock <- 1
}
