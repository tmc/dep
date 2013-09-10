package main

import (
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/exports"
	"os"
	"path"
	"path/filepath"
)

var app = cli.NewApp()

func action(fn func(c *cli.Context) ErrorCode) func(c *cli.Context) {
	return func(c *cli.Context) {
		// panics are handled in main
		errCode := fn(c)
		if errCode > 0 {
			exit(errCode)
		}
	}
}

var globalFlags = []cli.Flag{
	//cli.BoolFlag{"a", "for all packages in current GOPATH"},
	cli.StringFlag{"p", "", "for the given package import path"},
	cli.StringFlag{"d", ".", "for the package in the given directory"},
	cli.BoolFlag{"r", "also for all packages that are subdirectories of the package"},
}

var furtherInformation = "For more information, see 'dep intro'"

var PACKAGE *exports.Package

func parseGlobalFlags(c *cli.Context) {
	if c.Bool("r") {
		RECURSIVE = true
	}

	if c.String("p") != "" && c.Bool("a") {
		panic("you need to pass either -p or -a, not both")
	}

	if c.String("p") != "" && (c.String("d") != "" && c.String("d") != ".") {
		panic("you can't pass -p and -d, just one of them")
	}

	if c.Bool("a") && (c.String("d") != "" && c.String("d") != ".") {
		panic("you can't pass -p and -d, just one of them")
	}

	/*
		// disabled for now, since a single invalid package would prevent
		// everything from working
		if c.Bool("a") {
			//fmt.Println("IS ALL")
			ALL = true
			return
		}
	*/
	if p := c.String("p"); p != "" {
		PACKAGE_PATH = p
		PACKAGE = exports.DefaultEnv.Pkg(PACKAGE_PATH)
		PACKAGE_DIR = path.Join(GOPATH, "src", PACKAGE_PATH)
		return
	}

	if d := c.String("d"); d != "" {
		var err error
		if d == "." {
			d, err = os.Getwd()
			if err != nil {
				panic("can't get working directory: " + err.Error())
			}
		}

		d, err = filepath.Abs(d)
		if err != nil {
			panic("can't get absolute path: " + err.Error())
		}

		PACKAGE_DIR = d
		PACKAGE_PATH = exports.DefaultEnv.PkgPath(PACKAGE_DIR)
		PACKAGE = exports.DefaultEnv.Pkg(PACKAGE_PATH)
	}

}

func init() {
	app.Name = "dep"
	app.Usage = "manage go package dependancies via imports and exports"
	app.Version = "0.0.1"
	app.Action = func(c *cli.Context) {
		if c.Args()[0] == "intro" {
			fmt.Println(`here is an introduction into dep`)
		}
	}
	app.Commands = []cli.Command{
		{
			Name:        "store",
			Usage:       "writes dependancies to $GOPATH/src/[package]/dep.json",
			Description: furtherInformation,
			Action:      action(_store),
			Flags:       globalFlags,
		},
		{
			Name:        "lint",
			Usage:       "shows violations of the best practices $GOPATH/src/[package]",
			Description: furtherInformation,
			Flags:       globalFlags,
			Action:      action(_lint),
		},
		{
			Name:        "show",
			Usage:       "shows the dependancies of $GOPATH/src/[package]",
			Description: furtherInformation,
			Flags:       globalFlags,
			Action:      action(_show),
		},
		{
			Name:        "register",
			Usage:       "registers the dependancies in $GOPATH/dep.db",
			Description: furtherInformation,
			Flags:       globalFlags,
			Action:      action(_register),
		},
		{
			Name:        "get",
			Usage:       "get a go package and register it",
			Description: "passes all arguments to 'go get', see 'go help get'",
			Action:      action(_get),
		},
		{
			Name:        "install",
			Usage:       "installs a go package and register it",
			Description: "passes all arguments to 'go install', see 'go help get'",
			Action:      action(_install),
		},
		{
			Name:        "diff",
			Usage:       "shows the differences of current dependacies to the registered dependancies in $GOPATH/dep.db",
			Description: furtherInformation,
			Flags: append(globalFlags,
				cli.BoolFlag{"local", "compare with the local dependancies in $GOPATH/src/[package]/dep.json"},
			),
			Action: action(_diff),
		},
		{
			Name:        "update",
			Usage:       "updates a package if it is compatible to the packages in $GOPATH/src",
			Description: furtherInformation,
			//Description: "go get the given package into a tmp directory and check if all dependancies are met. if so go get it into the real path, otherwise throw an error",
			Flags: append(globalFlags,
				cli.BoolFlag{"force", "force the update"},
				cli.BoolFlag{"ignore-init", "ignore changes of the init functions"},
				//cli.BoolFlag{"ignore-main", "ignore changes of the main functions"},
				cli.BoolFlag{"diff", "do not install the update, merily show the possible incompatible differences"},
			),
			Action: action(_update),
		},
		{
			Name:  "revisions",
			Usage: "write the revisions of the dependancies to $GOPATH/src/[package]/[file]",
			Flags: append(globalFlags,
				cli.StringFlag{"file", "dep-rev.json", "the file to which the definitions of the revisions is written. may be used as input for 'dep checkout'"},
				cli.BoolFlag{"stdout", "write the revisions to stdout"},
			),
			Description: furtherInformation,
			Action:      action(_revisions),
		},
		{
			Name:        "checkout",
			Usage:       "checkout revisions of imported packages as defined in [file] to $GOPATH/src",
			Description: furtherInformation,
			Flags: []cli.Flag{
				cli.StringFlag{"file", "./dep-rev.json", "the file with the definitions of the revisions, must be the format returned by 'dep checkout'"},
				cli.BoolFlag{"force", "also checkout packages that are already installed"},
			},
			Action: action(_checkout),
		},
		/*
			{
				Name:   "fix",
				Usage:  "tries to find a common revision that all dependant packages may work with. warning will be slow",
				Action: action(_fix),
			},
		*/
	}
}

func main() {
	/*
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("ERROR: %v, aborting\n", r)
				cleanup()
				os.Exit(1)
			}
		}()
	*/
	app.Run(os.Args)
}
