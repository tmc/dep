package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	// "fmt"
	"github.com/metakeule/dep/depcore"
	"github.com/metakeule/exports"
	"os"
)

var (
	env        *depcore.Environment
	argVerbose bool
	argJson    bool
	pkg        *exports.Package
	pkgPath    string
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

func initV1() {
	if os.Getenv("GOPATH") == "" {
		panic("GOPATH not set")
	}
	if os.Getenv("DEP_TMP") == "" {
		panic("DEP_TMP not set")
	}
	env = depcore.NewEnv(os.Getenv("GOPATH"))
	env.TMPDIR = os.Getenv("DEP_TMP")

	// flag.StringVar(&argDir, "dir", ".", "for the package in the given directory")
	// flag.StringVar(&argPath, "path", "", "for the given package import path")
	flag.BoolVar(&argVerbose, "verbose", false, "print details about the actions taken")
	flag.BoolVar(&argJson, "json", false, "print in readable json format")
	// flag.BoolVar(&argNoInit, "no-init", false, "do no initialization before go getting the package")
}

func init() {
	initV1()
}

func toJson(i interface{}) string {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s", b)
}

var _ = strings.Join

var usage = `dep is a tool for the installation and update of Go packages

It prevents breakage of existing packages in GOPATH with the help
of a tentative installation in a temporary GOPATH and by detecting
breakage of dependancies in the go dependancy format (GDF).

Packages that use relative import paths are not supported and might
break.

For more information, see http://github.com/metakeule/dep

Required environment variables:
       - GOPATH should point to a valid GOPATH
       - DEP_TMP should point to a directory where tentative
         and temporary installations go to

PLEASE BE WARNED:
All actions act within the current GOPATH environment.
As dep is experimental at this point, you might loose all
your packages. No guarantee is made for whatever.

Usage:

         dep [options] command [package] 

If no package is given the package of the current working directory
is chosen.

Options:
    -verbose          Print details about the actions taken.
    -json             Print in machine readable json format

The commands are:

    gdf          Print the package's GDF.

    get          go get -u the given package and its dependancies
                 without breaking installed packages. Returns a list
                 of incompatibilities if there were any.
                 You should check, the integrity of your GOPATH with 'dep check'
                 before running 'dep get', otherwise dependencies might not be
                 checked properly.               
    
    track        track the imported packages with their revisions in 
                 the dep-rev.json file inside the package directory
                 That file will be used to get the exact same revisions
                 when using dep get.
    
    register     Add / update package's GDF inside the registry. 
                 Only needed for packages in the GOPATH that had already
                 been installed with other tools (e.g. go get / go install).
                 Not needed for packages that were installed via dep get.
    
    register-all like register, but also registers any included packages, as
                 they were currently in GOPATH/src
    
    unregister   removes a package from the registry
    
    diff         Show the difference in the GDFs between the given package 
                 and its GDF as it is in the registry.
    
    lint         Check if the given package respects the recommendations
                 for a package maintainer as given by the GDF.
                 Please keep in mind that not all recommendations can be
                 automatically checked.
    
    init         (Re)initialize the registry for the whole GOPATH and
                 check for incompatibilities in exports between the packages 
                 in GOPATH/src. WARNING: this erases the former compatibility
                 informations in the registry and the checksums of the working
                 init functions.
    
    check        checks the integrity of the whole GOPATH while respecting the
                 current registry.
`

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

//var x = depcore.P

func main() {

	flag.Usage = func() {
		fmt.Println(usage)
	}

	flag.Parse()
	if argVerbose && argJson {
		panic("-verbose and -json option are mutually exclusive")
	}
	/*
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("ERROR: %v, aborting\n", r)
				os.Exit(1)
			}
		}()
	*/

	if argVerbose {
		depcore.DEBUG = true
		exports.DEBUG = true
	}

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		//fmt.Println(toJson(pkg))
		fmt.Println(usage)
		return
	}
	cmd := args[0]

	env.Open()
	if len(env.IgnorePkgs) > 0 {
		fmt.Printf("WARNING: ignoring packages in %s\n\n", filepath.Join(env.GOPATH, ".depignore"))
	}
	if cmd != "init" && cmd != "check" {
		//getPackage()

		if len(args) == 2 {
			pkgPath = args[1]
		} else {
			pkgPath = getDefaultPackagePath()
		}
		if cmd != "unregister" {
			getPackage(pkgPath)
		}
	}
	defer env.Close()

	var err error
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
		conflicts, e := env.Get(pkg)
		if len(conflicts) > 0 {
			if argJson {
				fmt.Println(toJson(conflicts))
				os.Exit(1)
			}

			printConflicts(conflicts)
			os.Exit(1)
		}
		if e != nil {
			panic(e.Error())
		}
	case "init":
		//	fmt.Println("before check integrity")
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
		err = env.Register(false, pkg)
	case "register-all":
		err = env.Register(true, pkg)
	case "unregister":
		err = env.UnRegister(pkgPath)
	default:
		//panic("unknown command: " + cmd)
		fmt.Println(usage)
	}
	if err != nil {
		panic(err.Error())
	}

	//fmt.Println(strings.Join(flag.Args(), " "))

	//app.Run(os.Args)
}
