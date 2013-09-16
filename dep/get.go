package dep

import (
	"bytes"
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	"os"
	"os/exec"
)

func Get(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	pkgs := packages(o)
	args := []string{"get"}
	args = append(args, c.Args()...)
	installed := []*exports.Package{}
	dB, err := db.Open(DEP(o.GOPATH))

	if err != nil {
		db.CreateTables(dB)
	}

	defer func() {
		registerPackages(o, dB, installed...)
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
