package depcore

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
)

type DB struct {
	*db.DB
	Environment *Environment
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
func (dB *DB) hasConflict(p *exports.Package) (errors map[string][3]string) {
	pkg := p
	imp, err := db.GetImported(dB.DB, pkg.Path)
	if err != nil {
		panic(err.Error())
	}
	errors = map[string][3]string{}
	for _, im := range imp {
		key := fmt.Sprintf("%s: %s", im.Package, im.Name)
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

func (dB *DB) registerPackages(pkgs ...*exports.Package) {
	dbExps := []*db.Exp{}
	dbImps := []*db.Imp{}
	pkgMap := map[string]*db.Pkg{}

	for _, pkg := range pkgs {
		pExp, pImp := dB.Environment.packageToDBFormat(pkgMap, pkg)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	dbPkgs := []*db.Pkg{}

	for _, dbPgk := range pkgMap {
		dbPkgs = append(dbPkgs, dbPgk)
	}

	err := db.InsertPackages(dB.DB, dbPkgs, dbExps, dbImps)
	if err != nil {
		panic(err.Error())
	}
}

func (dB *DB) updatePackage(pkg string) error {
	tentative := dB.Environment.NewTentative()
	conflicts, err := tentative.updatePackage(pkg)

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
