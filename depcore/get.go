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
	installed := []*exports.Package{}
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
	"bytes"
	"fmt"
	"github.com/metakeule/exports"
	"os"
	"os/exec"
)

// TODO install it like an update in the safe tentative environment
// and check, if there are problems, only if not, move it to the real place
// make sure the dependancies are checked out in the right version
func (o *Environment) Get(pkg *exports.Package, _args ...string) error {
	args := []string{"get"}
	args = append(args, _args...)
	o.Open()
	defer o.Close()

	if o.PkgExists(pkg.Path) {
		fmt.Printf("package %s is already installed, skipping\n", pkg.Path)
		return nil
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
	o.db.registerPackages(pkg)

	return nil
}
