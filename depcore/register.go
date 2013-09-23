package depcore

import (
	// "fmt"
	// "github.com/metakeule/cli"
	// "github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	// "os"
)

/*
func CLIRegister(c *cli.Context, o *Options) ErrorCode {
	//parseGlobalFlags(c)
	_, dbFileerr := os.Stat(dEP(o.GOPATH))

	dB, err := db.Open(dEP(o.GOPATH))
	if err != nil {
		panic(err.Error())
	}
	defer dB.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables(dB)
	}

	pkgs := packages(o)
	registerPackages(o.Env, dB, pkgs...)
	return 0
}
*/

/*
func getDB(gopath string) *db.DB {
	_, dbFileerr := os.Stat(dEP(gopath))
	dB, err := db.Open(dEP(gopath))
	if err != nil {
		panic(err.Error())
	}
	if dbFileerr != nil {
		// fmt.Println(dbFileerr)
		db.CreateTables(dB)
	}
	return dB
}
*/

func (env *Environment) Register(includeImported bool, pkg *exports.Package) error {
	// dB := getDB(env.GOPATH)
	env.mkdb()
	env.db.registerPackages(includeImported, pkg)
	return nil
}

func (env *Environment) UnRegister(pkgPath string) error {
	// dB := getDB(env.GOPATH)
	env.mkdb()
	return env.db.DeletePackage(pkgPath)
}
