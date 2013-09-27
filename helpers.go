package main

import (
	"encoding/json"
	"fmt"
	"github.com/metakeule/gdf"
	"os"
	"path/filepath"
	"strings"
)

func getDefaultPackagePath() string {
	dir, err := os.Getwd()
	if err != nil {
		panic("can't get working directory: " + err.Error())
	}
	dir, err = filepath.Abs(dir)
	return env.PkgPath(dir)
}

func getPackage(pkgPath string) {
	var err error
	pkg, err = env.Pkg(pkgPath)
	if err != nil {
		panic(err.Error())
	}
	return
}

func toJson(i interface{}) string {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s", b)
}

func ask(question string) bool {
	fmt.Println(question + " (y/n)?")
	answer := ""
	_, err := fmt.Scanln(&answer)
	if err != nil {
		panic(err.Error())
	}
	answer = strings.ToLower(answer)
	answer = strings.TrimSpace(answer)
	if answer == "y" || answer == "yes" {
		return true
	}
	return false
}

func exitUnless(question string) {
	if !ask(question) {
		fmt.Println("aborted")
		os.Exit(1)
	}
}

func printConflicts(conflicts map[string]map[string][3]string) {
	if len(conflicts) > 0 {
		fmt.Println("#####     ERROR     #####")
	}
	for k, v := range conflicts {
		if k == "#dep-registry-orphan#" {
			for kk, _ := range v {
				fmt.Printf("orphaned package in registry: %s\n", kk)
			}
			continue
		}
		for kk, vv := range v {
			if kk == "#dep-registry-inconsistency#" {
				switch vv[0] {
				case "missing":
					fmt.Printf("missing package in registry: %s\n", k)
				case "imports":
					fmt.Printf("the imports of %s have changed:\n\n%s\n", k, vv[1])
				case "exports":
					fmt.Printf("the exports of %s have changed:\n\n%s\n", k, vv[1])
				default:
					panic("unknown #dep-registry-inconsistency# key: " + vv[0])
				}
				continue
			}

			switch vv[0] {
			case "error":
				fmt.Printf("can't get packages that are importing symbols from %s, error: %s\n", kk, vv[1])
			case "changed":
				imA := strings.Split(kk, ":")
				imP, imName := imA[0], imA[1]
				fmt.Printf("\n\npackage %#v\nimports symbol %#v\nof %#v\n\n   but that changed from\n\n%s\n\n     to\n\n%s\n", imP, imName, k, vv[1], vv[2])
			case "removed":
				imA := strings.Split(kk, ":")
				imP, imName := imA[0], imA[1]
				fmt.Printf("\n\npackage %#v\nimports symbol %#v\nof %#v\n\n   that now does no longer exists, but was\n\n%s\n", imP, imName, k, vv[1])
			default:
				panic("unknown conflict type " + vv[0])
			}

		}
	}
}

func runCmd(cmd string) (err error) {
	switch cmd {
	case "gdf":
		fmt.Println(toJson(pkg))
	case "lint":
		if argJson {
			panic("no json format available for dep lint")
		}
		e := env.Lint(pkg)
		if e != nil {
			fmt.Println(e.Error())
			os.Exit(1)
		}
		fmt.Printf("Looks ok: %s\n", pkg.Path)
	case "get":
		conflicts, e := env.Get(pkgPath, func(candidates ...*gdf.Package) bool {
			if argYes {
				return true
			}
			c := []string{}

			for _, cand := range candidates {
				c = append(c, cand.Path)
			}

			if ask(
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
			if argJson {
				fmt.Println(toJson(conflicts))
				os.Exit(1)
			}

			printConflicts(conflicts)
		}
		if e != nil {
			panic(e.Error())
		}
	case "init":
		exitUnless("This will destroy any former registrations.\nProceed")
		conflicts := env.Init()
		if len(conflicts) > 0 {
			if argJson {
				fmt.Println(toJson(conflicts))
				os.Exit(1)
			}
			printConflicts(conflicts)
			os.Exit(1)
		}
		num := env.NumPkgsInRegistry()
		fmt.Printf("GOPATH %s successfully initialized with %v packages\n", env.GOPATH, num)
	case "check":
		conflicts := env.CheckIntegrity()
		if len(conflicts) > 0 {
			if argJson {
				fmt.Println(toJson(conflicts))
				os.Exit(1)
			}
			printConflicts(conflicts)
			os.Exit(1)
		}
		num := env.NumPkgsInRegistry()
		fmt.Printf("GOPATH %s is upright (%v packages)\n", env.GOPATH, num)
	case "track":
		_, err = env.Track(pkg, true)
	case "diff":
		diff, e := env.Diff(pkg, true)
		if e != nil {
			panic(e.Error())
		}
		if diff != nil {
			if argJson {
				fmt.Println(toJson(diff))
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
		exitUnless(fmt.Sprintf("This will update the registration for package %s.\nProceed", pkg.Path))
		err = env.Register(false, pkg)
	case "register-all":
		exitUnless(fmt.Sprintf("This will update the registration for package %s and all packages that it depends on.\nProceed", pkg.Path))
		err = env.Register(true, pkg)
	case "unregister":
		exitUnless(fmt.Sprintf("This will remove the registration for package %s.\nProceed", pkg.Path))
		err = env.UnRegister(pkgPath)
	default:
		fmt.Println(usage)
	}
	return
}
