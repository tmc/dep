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

func (o *Environment) Track(pkg *exports.Package, recursive bool) (data []byte, err error) {
	revisions := map[string]revision{}
	for im, _ := range pkg.ImportedPackages {
		//o.trackedImportRevisions(pkg.Path)
		iPkg, e := o.Pkg(im)

		if e != nil {
			err = e
			return
		}
		if iPkg.Internal {
			continue
		}
		revisions[im] = o.getRevision(iPkg.Dir, pkg.Path)
		if recursive {
			o.recursiveImportRevisions(revisions, iPkg, pkg.Path)
			continue
		}
	}

	data, err = json.MarshalIndent(revisions, "", "  ")
	if err != nil {
		return
	}

	filename := path.Join(pkg.Dir, revFileName)
	err = ioutil.WriteFile(filename, data, 0644)
	return
}
