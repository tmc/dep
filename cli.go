package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
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

func init() {
	app.Name = "dep"
	app.Usage = "manage package dependancies by exports"
	app.Flags = []cli.Flag{
		cli.BoolFlag{"all", "execute command for all packages in current GOPATH"},
		cli.StringFlag{"package", "", "execute command the given package"},
		cli.StringFlag{"dir", ".", "execute command the package in the given directory"},
		cli.BoolFlag{"recursive", "execute command also for all packages that are subdirectories of the package"},
	}
	app.Action = func(c *cli.Context) {
		if c.Bool("recursive") {
			RECURSIVE = true
		}

		if c.GlobalString("package") == "" && !c.GlobalBool("all") {
			fmt.Println("you need to pass either --package or --all")
			exit(InvalidOptions)
		}

		if c.GlobalString("package") != "" && c.GlobalBool("all") {
			fmt.Println("you need to pass either --package or --all, not both")
			exit(InvalidOptions)
		}

		if c.GlobalString("package") != "" && c.GlobalString("dir") != "" {
			fmt.Println("you can't pass --package and --dir, just one of them")
			exit(InvalidOptions)
		}

		if p := c.GlobalString("package"); p != "" {
			PACKAGE_PATH = p
		}

		if d := c.GlobalString("dir"); d != "" {
			PACKAGE_DIR = d
		}

		if c.GlobalBool("all") {
			ALL = true
		}

	}

	app.Commands = []cli.Command{
		{
			Name:   "store",
			Usage:  "store dependancies inside the package",
			Action: action(_store),
		},
		{
			Name:   "lint",
			Usage:  "report violations of the best practices",
			Action: action(_lint),
		},
		{
			Name:   "info",
			Usage:  "show dependancies",
			Action: action(_info),
		},
		{
			Name:   "register",
			Usage:  "register dependancies in GOPATH",
			Action: action(_register),
		},
		{
			Name:   "get",
			Usage:  "get a go package",
			Action: action(_get),
		},
		{
			Name:   "install",
			Usage:  "install a go package",
			Action: action(_install),
		},
		{
			Name:  "diff",
			Usage: "return difference of current dependacies to the stored one in GOPATH, returns error if the dependancy is not yet stored in GOPATH",
			Flags: []cli.Flag{
				cli.BoolFlag{"local", "check against the local dependancy in the package directory, returns error, if dependancy info is not in the local directory"},
			},
			Action: action(_diff),
		},
		{
			Name:        "update",
			Usage:       "update the given package if it is compatible to the other packages",
			Description: "go get the given package into a tmp directory and check if all dependancies are met. if so go get it into the real path, otherwise throw an error",
			Flags: []cli.Flag{
				cli.BoolFlag{"force", "force the update even if the exports changed in a possible incompatible way"},
				cli.BoolFlag{"ignore-init", "ignore changes in the init functions"},
				cli.BoolFlag{"ignore-main", "ignore changes in the main functions"},
				cli.BoolFlag{"diff", "do not install the update, but show the possible incompatible differences"},
			},
			Action: action(_update),
		},
		{
			Name:   "fix",
			Usage:  "try to find a common revision that all dependant packages may work with. warning will be slow",
			Action: action(_fix),
		},
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ERROR: %v, aborting\n", r)
			cleanup()
			os.Exit(1)
		}
	}()

	app.Run(os.Args)
}
