package dep

import (
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"os"
)

func Register(c *cli.Context, o *Options) ErrorCode {
	//parseGlobalFlags(c)
	_, dbFileerr := os.Stat(DEP(o.GOPATH))

	dB, err := db.Open(DEP(o.GOPATH))
	if err != nil {
		panic(err.Error())
	}
	defer dB.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables(dB)
	}

	pkgs := packages(o)
	registerPackages(o, dB, pkgs...)
	return 0
}
