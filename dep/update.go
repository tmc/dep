package dep

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"os"
)

func _update_pkg(c *cli.Context, o *Options, dB *db.DB, pkg string, keepTemp bool) (conflicts map[string]map[string][3]string, err error) {
	// parseGlobalFlags(c)
	tmpDir := mkdirTempDir(o)

	defer func() {
		if !keepTemp {
			err := os.RemoveAll(tmpDir)
			if err != nil {
				panic(err.Error())
			}
		}
	}()

	return _updatePackage(tmpDir, o, dB, pkg)
}

// TODO: checkout certain revisions if there is a dep.rev file
func Update(c *cli.Context, o *Options) ErrorCode {
	//_, dbFileerr := os.Stat(DEP(o.GOPATH))
	dB, err := db.Open(DEP(o.GOPATH))
	if err != nil {
		err = CreateDB(o.GOPATH)
		if err != nil {
			panic(err.Error())
		}
		dB, err = db.Open(DEP(o.GOPATH))
		if err != nil {
			panic(err.Error())
		}
		//panic(err.Error())
	}
	defer dB.Close()

	err = CheckIntegrity(o, o.Env)
	if err != nil {
		panic(err.Error())
	}

	pkgs := packages(o)

	// TODO: make no double updates for packages, that had already been updated
	for _, pkg := range pkgs {
		conflicts, pkgErr := _update_pkg(c, o, dB, pkg.Path, c.Bool("keep-temp-gopath"))
		if pkgErr != nil {
			panic(pkgErr.Error())
		}
		if len(conflicts) > 0 {
			b, e := json.MarshalIndent(conflicts, "", "  ")
			if e != nil {
				panic(e.Error())
			}
			fmt.Printf("%s\n", b)
			return UpdateConflict
		}
	}

	return 0
}
