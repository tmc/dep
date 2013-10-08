package depcore

/*
import (
	"bytes"
	"fmt"
	// "github.com/go-dep/cli"
	// "github.com/go-dep/dep/db"
	"github.com/go-dep/exports"
	"os"
	"os/exec"
)
*/

/*

func CLIInstall(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	pkgs := packages(o)
	args := []string{"install"}
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

// TODO install it like an update in the safe tentative environment
// and check, if there are problems, only if not, move it to the real place
// make sure the dependancies are checked out in the right version
/*
func (o *Environment) Install(pkg *exports.Package, _args ...string) error {
	args := []string{"install"}
	args = append(args, _args...)
	o.Open()
	defer o.Close()

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
*/
