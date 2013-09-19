package depcore

import (
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

func mapDiff(_old map[string]string, _new map[string]string) (diff []string) {
	diff = []string{}
	var visited = map[string]bool{}

	for k, v := range _old {
		visited[k] = true
		vNew, inNew := _new[k]
		if !inNew {
			diff = append(diff, "---"+k+": "+v)
			continue
		}
		if v != vNew {
			diff = append(diff, "---"+k+": "+v)
			diff = append(diff, "+++"+k+": "+vNew)
		}
	}

	for k, v := range _new {
		if !visited[k] {
			diff = append(diff, "+++"+k+": "+v)
		}
	}
	return
}

func (o *Environment) CLIDiff(pkg *exports.Package) (diff *pkgDiff, err ErrorCode) {
	o.Open()
	defer o.Close()

	dbpkg, exps, imps, e := o.DB.GetPackage(pkg.Path, true, true)
	if e != nil {
		panic("package not registered: " + pkg.Path)
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
		pDiff.Exports = mapDiff(oldExports, pkgjs.Exports)

		var oldImports = map[string]string{}

		for _, dbImp := range imps {
			oldImports[dbImp.Import+"#"+dbImp.Name] = dbImp.Value
		}
		pDiff.Imports = mapDiff(oldImports, pkgjs.Imports)

		if len(pDiff.Exports) > 0 || len(pDiff.Imports) > 0 {
			return pDiff, 0
		}
	}
	return nil, 0
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
