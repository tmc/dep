package depcore

/*
import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/exports"
	"io/ioutil"
	"path"
)

func CLIRevisions(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	file := c.String("file")
	stdout := c.Bool("stdout")
	inclIndirect := c.Bool("include-indirect")
	pkgs := packages(o)
	allrevisions := map[string]revision{}

	for _, pkg := range pkgs {
		revisions := map[string]revision{}
		for im, _ := range pkg.ImportedPackages {
			revisions[im] = pkgRevision(o, path.Join(o.Env.GOPATH, "src", im), pkg.Path)
			if inclIndirect {
				indirectRev(o, revisions, o.Env.Pkg(im), pkg.Path)
				continue
			}
		}

		if stdout {
			for k, v := range revisions {
				if _, exists := allrevisions[k]; !exists {
					allrevisions[k] = v
				}
			}
			continue
		}
		data, err := json.MarshalIndent(revisions, "", "  ")
		if err != nil {
			panic(err.Error())
		}

		dir, _ := pkg.Dir()
		filename := path.Join(dir, file)
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			panic(err.Error())
		}
	}
	if stdout {
		data, err := json.MarshalIndent(allrevisions, "", "  ")
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%s\n", data)
	}
	return 0
}
*/

import (
	"encoding/json"
	"github.com/metakeule/exports"
	"io/ioutil"
	"path"
)

func (o *Environment) CLIRevisions(pkg *exports.Package, includeIndirect bool) (data []byte, err error) {
	revisions := map[string]revision{}
	for im, _ := range pkg.ImportedPackages {
		o.trackedImportRevisions(pkg)
		revisions[im] = o.getRevision(o.PkgDir(im), pkg.Path)
		if includeIndirect {

			indirectRev(o, revisions, o.Pkg(im), pkg.Path)
			continue
		}
	}

	data, err = json.MarshalIndent(revisions, "", "  ")
	if err != nil {
		return
	}

	dir, _ := pkg.Dir()
	filename := path.Join(dir, file)
	err = ioutil.WriteFile(filename, data, 0644)
	return
}
