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
	env     *depcore.Environment
	argDir  string
	argPath string
	pkg     *exports.Package
)

func getPackage() {
	if argPath != "" && argDir != "" && argDir != "." {
		panic("you can't pass -path and -dir, just one of them")
	}

	if argPath != "" {
		pkg = env.Pkg(argPath)
		return
	}

	if argDir != "" {
		var err error
		if argDir == "." {
			argDir, err = os.Getwd()
			if err != nil {
				panic("can't get working directory: " + err.Error())
			}
		}

		argDir, err = filepath.Abs(argDir)
		if err != nil {
			panic("can't get absolute path: " + err.Error())
		}

		argPath = env.PkgPath(argDir)
		pkg = env.Pkg(argPath)
	}
	return
}

func init() {
	if os.Getenv("GOPATH") == "" {
		panic("GOPATH not set")
	}
	if os.Getenv("DEP_TMP") == "" {
		panic("DEP_TMP not set")
	}
	env = depcore.NewEnv(os.Getenv("GOPATH"))
	env.TMPDIR = os.Getenv("DEP_TMP")

	flag.StringVar(&argDir, "dir", ".", "for the package in the given directory")
	flag.StringVar(&argPath, "path", "", "for the given package import path")
}

func pkgJsonString() string {
	b, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("%s", b)
}

var _ = strings.Join

func main() {
	/*
		flag.Usage = func() {
			fmt.Println("huho")
		}
	*/
	flag.Parse()
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR: %v, aborting\n", r)
			os.Exit(1)
		}
	}()

	env.Open()
	getPackage()
	defer env.Close()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println(pkgJsonString())
		return
	}
	cmd := args[0]
	switch cmd {
	case "track":
		env.Track(pkg, true)
	default:
		panic("unknown command: " + cmd)
	}

	//fmt.Println(strings.Join(flag.Args(), " "))

	//app.Run(os.Args)
}
