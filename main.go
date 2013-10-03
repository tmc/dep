package main

import (
	"flag"
	"fmt"
	"github.com/metakeule/dep/depcore"
	"github.com/metakeule/gdf"
	"os"
	"path/filepath"
	"strings"
)

var (
	env        *depcore.Environment
	argVerbose bool
	argJson    bool
	argYes     bool
	argNoWarn  bool
	argPanic   bool
	pkg        *gdf.Package
	pkgPath    string
)

func initV1() {
	if os.Getenv("GOPATH") == "" {
		panic("GOPATH not set")
	}
	if os.Getenv("DEP_TMP") == "" {
		panic("DEP_TMP not set")
	}
	env = depcore.NewEnv(strings.Split(os.Getenv("GOPATH"), ";")[0])
	env.TMPDIR = os.Getenv("DEP_TMP")

	flag.BoolVar(&argVerbose, "verbose", false, "print details about the actions taken")
	flag.BoolVar(&argJson, "json", false, "print in readable json format")
	flag.BoolVar(&argYes, "y", false, "answer all questions with 'yes'")
	flag.BoolVar(&argNoWarn, "no-warn", false, "suppress warnings")
	flag.BoolVar(&argPanic, "panic", false, "panic on errors")
}

func main() {
	flag.Usage = func() {
		fmt.Println(usage)
	}

	flag.Parse()
	if argVerbose && argJson {
		panic("-verbose and -json option are mutually exclusive")
	}

	if !argPanic {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("ERROR: %v, aborting\n", r)
				os.Exit(1)
			}
		}()
	}

	if argVerbose {
		depcore.VERBOSE = true
		gdf.VERBOSE = true
	}

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		fmt.Println(usage)
		return
	}
	cmd := args[0]

	env.Open()
	if len(env.IgnorePkgs) > 0 && !argNoWarn {
		fmt.Printf("WARNING: ignoring packages in %s\n\n", filepath.Join(env.GOPATH, ".depignore"))
	}

	if cmd != "init" && cmd != "check" {
		if len(args) == 2 {
			pkgPath = args[1]
		} else {
			pkgPath = getDefaultPackagePath()
		}

		if cmd != "unregister" && cmd != "get" {
			getPackage(pkgPath)
		}

	}
	defer env.Close()

	if err := runCmd(cmd); err != nil {
		panic(err.Error())
	}
}
