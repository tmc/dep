package dep

import (
	"bytes"
	_ "code.google.com/p/go.exp/inotify"
	"encoding/json"
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	_ "launchpad.net/goamz/aws"
	"os"
	"os/exec"
	"path"
	"runtime"
)

// TODO: checkout certain revisions if there is a dep.rev file
func Update(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	tmpDir := mkdirTempDir(o)

	defer func() {
		if !c.Bool("keep-temp-gopath") {
			err := os.RemoveAll(tmpDir)
			if err != nil {
				panic(err.Error())
			}
		}
	}()

	pkgs := packages(o)
	tempEnv := exports.NewEnv(runtime.GOROOT(), tmpDir)

	_, dbFileerr := os.Stat(DEP(o))
	dB, err := db.Open(DEP(o))
	if err != nil {
		panic(err.Error())
	}
	defer dB.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables(dB)
	}

	conflicts := map[string]map[string][3]string{}

	// TODO: check if the package and its dependancies are installed in the
	// default path, if so, check, they are registered / updated in the database.
	// if not, register /update them
	// TODO make a db connection to get the conflicting
	// packages.
	// it might be necessary to make an update of the db infos first
	visited := map[string]bool{}

	for _, pkg := range pkgs {
		if !visited[pkg.Path] {
			visited[pkg.Path] = true
			args := []string{"get", "-u", pkg.Path}
			//args = append(args, c.Args()...)

			cmd := exec.Command("go", args...)
			cmd.Env = []string{
				fmt.Sprintf(`GOPATH=%s`, tmpDir),
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
			// make flag with --skip-gotest
			tempPkgPath := path.Join(tempEnv.GOPATH, "src", pkg.Path)
			// go test should fail, if the newly installed packages are not compatible
			res, e := runGoTest(o, tmpDir, tempPkgPath)
			if e != nil {
				panic("Error while running 'go test " + tempPkgPath + "'\n" + string(res) + "\n" + e.Error())
			}

			errs := checkConflicts(o, dB, tempEnv, pkg)
			if len(errs) > 0 {
				conflicts[pkg.Path] = errs
			}
		}

		for im, _ := range pkg.Imports {
			if !visited[pkg.Path] {
				imPkg := o.Env.Pkg(im)
				tempPkgPath := path.Join(tempEnv.GOPATH, "src", im)
				// go test should fail, if the newly installed packages are not compatible
				res, e := runGoTest(o, tmpDir, tempPkgPath)
				if e != nil {
					panic("Error while running 'go test " + tempPkgPath + "'\n" + string(res) + "\n" + e.Error())
				}

				errs := checkConflicts(o, dB, tempEnv, imPkg)
				if len(errs) > 0 {
					conflicts[pkg.Path] = errs
				}
			}
		}
	}

	if len(conflicts) > 0 {
		b, e := json.MarshalIndent(conflicts, "", "  ")
		if e != nil {
			panic(e.Error())
		}
		fmt.Printf("%s\n", b)
		return UpdateConflict
	}

	// if we got here, everything is fine and we may do our update

	// TODO: it might be better to simply move the installed packages
	// instead of go getting them again, because they might
	// have changed in the meantime
	for _, pkg := range pkgs {
		// update all dependant packages as well
		args := []string{"get", "-u", pkg.Path}
		//args = append(args, c.Args()...)

		cmd := exec.Command("go", args...)
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
	}
	return 0
}
