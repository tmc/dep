package depcore

/*
import (
	"bytes"
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	"os"
	"os/exec"
)

func CLIGet(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	pkgs := packages(o)
	args := []string{"get"}
	args = append(args, c.Args()...)
	installed := []*gdf.Package{}
	dB, err := db.Open(dEP(o.GOPATH))

	if err != nil {
		db.CreateTables(dB)
	}

	defer func() {
		registerPackages(o.Env, dB, installed...)
		dB.Close()
	}()

	for _, pkg := range pkgs {
		if o.Env.PkgExists(pkg.Path) {
			fmt.Printf("package %s is already installed, skipping\n", pkg.Path)
			continue
		}

		cmd := exec.Command("go", append(args, pkg.Path)...)
		cmd.Env = []string{
			fmt.Sprintf(`GOPATH=%s`, o.GOPATH),
			fmt.Sprintf(`GOROOT=%s`, o.GOROOT),
			fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		cmd.Stdout = &stdout
		err := cmd.Run()
		if err != nil {
			panic(stdout.String() + "\n" + stderr.String())
		}
		installed = append(installed, pkg)
	}

	return 0
}
*/

import (
	"fmt"
	"github.com/metakeule/gdf"
	"os"
	"path/filepath"
	"time"
)

func (o *Environment) Get(pkg *gdf.Package, confirmation func(candidates ...*gdf.Package) bool) (conflicts map[string]map[string][3]string, err error) {
	o.Open()
	defer o.Close()

	t := o.newTentative()
	var er error
	conflicts, er = t.updatePackage(pkg.Path, confirmation)

	if er != nil {
		dir := filepath.Dir(t.GOPATH)
		new_path := filepath.Join(dir, fmt.Sprintf("gopath_%v", time.Now().UnixNano()))
		os.Rename(t.GOPATH, new_path)
		err = fmt.Errorf(er.Error()+"\ncheck or remove the temporary gopath at %s\n", new_path)
	}

	return
}
