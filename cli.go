package dep

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

var app = cli.NewApp()

func action(c *cli.Context, fn func(c *cli.Context) ErrorCode) func(c *cli.Context) {
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
		cli.BoolFlag("all", false, "execute command for all packages in current GOPATH"),
		cli.StringFlag("package", "", "execute command the given package"),
		cli.StringFlag("dir", ".", "execute command the package in the given directory"),
	}

	cl := cli.Command{}
	//cl.Flags
	app.Commands = []cli.Command{
		{
			Name:      "store",
			ShortName: "s",
			Usage:     "store dependancies inside the package",
			Action: func(c *cli.Context) {

			},
		},
		{
			Name:      "lint",
			ShortName: "l",
			Usage:     "report violations of the best practices",
			Action: func(c *cli.Context) {

			},
		},
		{
			Name:      "info",
			ShortName: "i",
			Usage:     "show dependancies",
			Action: func(c *cli.Context) {

			},
		},
		{
			Name:      "register",
			ShortName: "r",
			Usage:     "register dependancies in GOPATH",
			Action: func(c *cli.Context) {

			},
		},
		{
			Name:      "diff",
			ShortName: "d",
			Usage:     "return difference of current dependacies to the stored one in GOPATH, returns error if the dependancy is not yet stored in GOPATH",
			Flags: []cli.Flag{
				cli.BoolFlag("local", false, "check against the local dependancy in the package directory, returns error, if dependancy info is not in the local directory"),
			},
			Action: func(c *cli.Context) {

			},
		},
		{
			Name:        "update",
			ShortName:   "u",
			Usage:       "update the given package if it is compatible to the other packages",
			Description: "go get the given package into a tmp directory and check if all dependancies are met. if so go get it into the real path, otherwise throw an error",
			Flags: []cli.Flag{
				cli.BoolFlag("force", false, "force the update even if the exports changed in a possible incompatible way"),
				cli.BoolFlag("ignore-init", false, "ignore changes in the init functions"),
				cli.BoolFlag("ignore-main", false, "ignore changes in the main functions"),
				cli.BoolFlag("diff", false, "do not install the update, but show the possible incompatible differences"),
			},
			Action: func(c *cli.Context) {

			},
		},
		{
			Name:      "fix",
			ShortName: "f",
			Usage:     "try to find a common revision that all dependant packages may work with. warning will be slow",
			Action: func(c *cli.Context) {

			},
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
