package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-dep/dep/depcore"
	"github.com/go-dep/gdf"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
	case "init-functions":
		if Args.Json {
			S.Json(pkg.RawInits)
			os.Exit(0)
		}

		if len(pkg.RawInits) == 0 {
			fmt.Println("No init functions.")
			os.Exit(0)
		}

		for file, initfn := range pkg.RawInits {
			fmt.Printf("\n\nFile: %s\n%s\n", file, initfn)
		}
		os.Exit(0)
	case "get":
		if !Args.SkipCheck {
			fmt.Println("running dep check, please wait...")
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
				fmt.Sprintf("GOPATH %s is ok (%v packages)\n",
					env.GOPATH, env.NumPkgsInRegistry())
			}
		}
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

		conflicts, changed, e := env.Get(Args.PkgPath, over, func(candidates ...*gdf.Package) bool {
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
		if Args.Json {
			S.Json(changed)
			os.Exit(0)
		}
		S.PrintChanged(changed)
		os.Exit(0)
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
			S.Out("GOPATH %s is ok (%v packages)\n",
				env.GOPATH, env.NumPkgsInRegistry())
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
				fmt.Printf("\n#### exported symbols ####\n%s\n",
					strings.Join(diff.Exports, "\n"))
			}
			if len(diff.Imports) > 0 {
				fmt.Printf("\n#### imported symbols ####\n%s\n",
					strings.Join(diff.Imports, "\n"))
			}
			os.Exit(1)
		}
	case "register":
		S.ExitUnless(
			fmt.Sprintf(
				"This will update the registration for package %s.\nProceed",
				pkg.Path))
		err = env.Register(false, pkg)
	case "imports":
		imported := []string{}

		for im, _ := range pkg.ImportedPackages {
			imported = append(imported, im)
		}

		if Args.Json {
			S.Json(imported)
			os.Exit(0)
		}

		S.Out("packages imported by %s:\n\n\t%s\n",
			pkg.Path,
			strings.Join(imported, "\n\t"))

	case "register-included":
		S.ExitUnless(
			fmt.Sprintf(
				"This will update the registration for package %s and all packages that it depends on.\nProceed",
				pkg.Path))
		err = env.Register(true, pkg)
	case "unregister":
		S.ExitUnless(
			fmt.Sprintf(
				"This will remove the registration for package %s.\nProceed",
				Args.PkgPath))
		err = env.UnRegister(Args.PkgPath)

		if err != nil {
			S.Error(err.Error())
		}
		os.Exit(0)

	case "dump":
		pkgs, err := env.Dump()

		if err != nil {
			S.Error(err.Error())
		}
		S.Json(pkgs)
		os.Exit(0)

	case "backups-cleanup":
		cmd := exec.Command(
			"find",
			filepath.Join(env.GOPATH, "src"),
			"-type", "d",
			"-name", "*"+depcore.BackupPostFix,
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		errr := cmd.Run()
		if errr != nil {
			S.Error(errr.Error())
		}
		files, er := S.ReadLines(&out)
		if er != nil {
			S.Error(er.Error())
		}

		if len(files) == 0 {
			S.Out("No backups found.")
			os.Exit(0)
		}
		pkgPaths := make([]string, len(files))
		srcPath := filepath.Join(env.GOPATH, "src") + "/"
		for i, m := range files {
			pkgPaths[i] = strings.Replace(m, srcPath, "", 1)
		}

		question := fmt.Sprintf("the following backups will be deleted: \n\n\t%s\n\nDo you want to continue", strings.Join(pkgPaths, "\n\t"))
		S.ExitUnless(question)

		for _, m := range files {
			os.RemoveAll(m)
		}

	case "gopath-cleanup":
		gopaths := filepath.Join(env.TMPDIR, depcore.TempGOPATHPreFix+"*")
		matches, err := filepath.Glob(gopaths)
		if err != nil {
			S.Error(err.Error())
		}

		if len(matches) == 0 {
			S.Out("No temporary gopaths found.")
			os.Exit(0)
		}

		question := fmt.Sprintf("the following temporary gopaths will be deleted: \n\n\t%s\n\nDo you want to continue", strings.Join(matches, "\n\t"))
		S.ExitUnless(question)
		for _, m := range matches {
			os.RemoveAll(m)
		}
	default:
		fmt.Println(usage)
	}
	return
}
