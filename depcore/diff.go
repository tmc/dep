package depcore

import (
	"fmt"
	"github.com/metakeule/exports"
)

/*
import (
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"os"
)
*/

type pkgDiff struct {
	Path    string
	Exports []string
	Imports []string
}

func mapDiff(_old map[string]string, _new map[string]string, includeVals bool) (diff []string) {
	diff = []string{}
	var visited = map[string]bool{}

	for k, v := range _old {
		visited[k] = true
		vNew, inNew := _new[k]
		if !inNew {
			if includeVals {
				diff = append(diff, "---"+k+": "+v)
			} else {
				diff = append(diff, "---"+k)
			}
			continue
		}
		if includeVals {
			if v != vNew {
				diff = append(diff, "---"+k+": "+v)
				diff = append(diff, "+++"+k+": "+vNew)
			}
		}
	}

	for k, v := range _new {
		if !visited[k] {
			if includeVals {
				diff = append(diff, "+++"+k+": "+v)
			} else {
				diff = append(diff, "+++"+k)
			}
		}
	}
	return
}

func (o *Environment) Diff(pkg *exports.Package, includeImportTypeDiffs bool) (diff *pkgDiff, err error) {
	dbpkg, exps, imps, e := o.db.GetPackage(pkg.Path, true, true)
	if e != nil {
		err = fmt.Errorf("package not registered: %s\n", pkg.Path)
		return
	}

	js := asJson(pkg)

	// TODO: check the hash instead, escp. check the exports and imports hash
	if exports.Hash(string(js)) != exports.Hash(string(dbpkg.Json)) {
		//__diff(a, b)
		pkgjs := pkg

		var oldExports = map[string]string{}

		for _, dbExp := range exps {
			oldExports[dbExp.Name] = dbExp.Value
		}

		pDiff := &pkgDiff{}
		pDiff.Path = pkg.Path
		pDiff.Exports = mapDiff(oldExports, pkgjs.Exports, true)

		var oldImports = map[string]string{}

		for _, dbImp := range imps {
			oldImports[dbImp.Import+"#"+dbImp.Name] = dbImp.Value
		}
		pDiff.Imports = mapDiff(oldImports, pkgjs.Imports, includeImportTypeDiffs)

		if len(pDiff.Exports) > 0 || len(pDiff.Imports) > 0 {
			return pDiff, nil
		}
	}
	return nil, nil
}

/*
func CLIDiff(c *cli.Context, o *Environment) ErrorCode {
	// parseGlobalFlags(c)
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

	res := []*pkgDiff{}

	for _, pk := range pkgs {

		dbpkg, exps, imps, e := db.GetPackage(dB, pk.Path, true, true)
		if e != nil {
			panic("package not registered: " + pk.Path)
		}

		js := asJson(pk)

		// TODO: check the hash instead, escp. check the exports and imports hash
		if string(js) != string(dbpkg.Json) {
			//__diff(a, b)
			pkgjs := pk

			var oldExports = map[string]string{}

			for _, dbExp := range exps {
				oldExports[dbExp.Name] = dbExp.Value
			}

			pDiff := &pkgDiff{}
			pDiff.Path = pk.Path
			pDiff.Exports = mapDiff(oldExports, pkgjs.Exports)

			var oldImports = map[string]string{}

			for _, dbImp := range imps {
				oldImports[dbImp.Import+"#"+dbImp.Name] = dbImp.Value
			}
			pDiff.Imports = mapDiff(oldImports, pkgjs.Imports)

			if len(pDiff.Exports) > 0 || len(pDiff.Imports) > 0 {
				res = append(res, pDiff)
			}
		}
	}
	if len(res) > 0 {
		fmt.Printf("%s\n", toJson(res))
	}
	return 0
}
*/
