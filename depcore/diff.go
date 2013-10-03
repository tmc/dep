package depcore

import (
	"fmt"
	"github.com/metakeule/gdf"
)

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

func (o *Environment) Diff(pkg *gdf.Package, includeImportTypeDiffs bool) (diff *pkgDiff, err error) {
	dbpkg, exps, imps, e := o.db.GetPackage(pkg.Path, true, true)
	if e != nil {
		err = fmt.Errorf("package not registered: %s\n", pkg.Path)
		return
	}

	if pkg.JsonMd5() != dbpkg.JsonMd5 {
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
