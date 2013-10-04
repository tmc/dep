package main

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/gdf"
	"io/ioutil"
	"os"
	"strings"
)

func runCmd(cmd string) (err error) {
	switch cmd {
	case "gdf":
		S.Json(pkg)
		os.Exit(0)
	case "lint":
		if Args.Json {
			S.Error("no json format available for dep lint")
		}
		e := env.Lint(pkg)
		if e != nil {
			fmt.Println(e.Error())
			os.Exit(1)
		}
		if Args.Verbose {
			S.Out("Looks ok: %s\n", pkg.Path)
		}
	case "get":
		over := []*gdf.Package{}
		if Args.Override != "" {
			b, e := ioutil.ReadFile(Args.Override)
			if e != nil {
				panic(e.Error())
			}

			e = json.Unmarshal(b, &over)
			if e != nil {
				panic(e.Error())
			}
		}

		conflicts, e := env.Get(Args.PkgPath, over, func(candidates ...*gdf.Package) bool {
			if Args.Yes {
				return true
			}
			c := []string{}

			for _, cand := range candidates {
				c = append(c, cand.Path)
			}

			if S.Ask(
				fmt.Sprintf(
					`The following packages will be updated:
	%s
This is alpha software and may break your existing packages in %s.
Proceed`,
					strings.Join(c, "\n\t"),
					env.GOPATH)) {
				return true
			}
			return false
		})
		if len(conflicts) > 0 {
			if Args.Json {
				S.Json(conflicts)
				os.Exit(1)
			}
			S.PrintConflicts(conflicts)
		}
		if e != nil {
			S.Error(e.Error())
		}
	case "init":
		S.ExitUnless("This will destroy any former registrations.\nProceed")
		conflicts := env.Init()
		if len(conflicts) > 0 {
			if Args.Json {
				S.Json(conflicts)
				os.Exit(1)
			}
			S.PrintConflicts(conflicts)
			os.Exit(1)
		}
		if Args.Verbose {
			S.Out("GOPATH %s successfully initialized with %v packages\n",
				env.GOPATH,
				env.NumPkgsInRegistry(),
			)
		}
	case "check":
		conflicts := env.CheckIntegrity()
		if len(conflicts) > 0 {
			if Args.Json {
				S.Json(conflicts)
				os.Exit(1)
			}
			S.PrintConflicts(conflicts)
			os.Exit(1)
		}
		if Args.Verbose {
			S.Out("GOPATH %s is upright (%v packages)\n", env.GOPATH, env.NumPkgsInRegistry())
		}
	case "track":
		_, err = env.Track(pkg, true)
		if err != nil {
			S.Error(err.Error())
		}
	case "diff":
		diff, e := env.Diff(pkg, true)
		if e != nil {
			S.Error(e.Error())
		}
		if diff != nil {
			if Args.Json {
				S.Json(diff)
				os.Exit(1)
			}
			fmt.Printf("%s has changed since the last registration\n", pkg.Path)
			if len(diff.Exports) > 0 {
				fmt.Printf("\n#### exported symbols ####\n%s\n", strings.Join(diff.Exports, "\n"))
			}
			if len(diff.Imports) > 0 {
				fmt.Printf("\n#### imported symbols ####\n%s\n", strings.Join(diff.Imports, "\n"))
			}
			os.Exit(1)
		}
	case "register":
		S.ExitUnless(fmt.Sprintf("This will update the registration for package %s.\nProceed", pkg.Path))
		err = env.Register(false, pkg)
	case "register-all":
		S.ExitUnless(fmt.Sprintf("This will update the registration for package %s and all packages that it depends on.\nProceed", pkg.Path))
		err = env.Register(true, pkg)
	case "unregister":
		S.ExitUnless(fmt.Sprintf("This will remove the registration for package %s.\nProceed", pkg.Path))
		err = env.UnRegister(Args.PkgPath)
	default:
		fmt.Println(usage)
	}
	return
}
