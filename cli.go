package main

import (
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/dep"
	"github.com/metakeule/exports"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

var app = cli.NewApp()

func action(fn func(*cli.Context, *dep.Options) dep.ErrorCode) func(c *cli.Context) {
	return func(c *cli.Context) {
		// panics are handled in main
		errCode := fn(c, options)
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

func parseGlobalFlags(c *cli.Context) {

	if c.Bool("r") {
		options.Recursive = true
		//RECURSIVE = true
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
		options.PackagePath = p
		//		PACKAGE_PATH = p
		//PACKAGE = exports.DefaultEnv.Pkg(PACKAGE_PATH)
		options.Package = options.Env.Pkg(options.PackagePath)
		//PACKAGE_DIR = path.Join(GOPATH, "src", PACKAGE_PATH)
		options.PackageDir = path.Join(options.GOPATH, "src", options.PackagePath)
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

		//PACKAGE_DIR = d
		options.PackageDir = d
		//		PACKAGE_PATH = exports.DefaultEnv.PkgPath(PACKAGE_DIR)
		options.PackagePath = options.Env.PkgPath(options.PackageDir)
		//PACKAGE = exports.DefaultEnv.Pkg(PACKAGE_PATH)
		options.Package = options.Env.Pkg(options.PackagePath)
	}
	return
}

var options *dep.Options

func init() {
	options = &dep.Options{}
	options.HOME = os.Getenv("HOME")
	options.GOPATH = os.Getenv("GOPATH")
	options.GOROOT = runtime.GOROOT()
	options.Env = exports.NewEnv(runtime.GOROOT(), options.GOPATH)

	ErrorCodeInfos[dep.GOROOTNotSet] = "$GOROOT environment variable is not set"
	ErrorCodeInfos[dep.GOPATHNotSet] = "$GOPATH environment variable is not set"
	ErrorCodeInfos[dep.HOMENotSet] = "$HOME environment variable is not set"
	ErrorCodeInfos[dep.GOPATHInvalid] = fmt.Sprintf("$GOPATH directory %s is not conforming to the standard layout (bin,pkg,src directories)", options.GOPATH)
	ErrorCodeInfos[dep.InvalidOptions] = "given options are invalid"
	ErrorCodeInfos[dep.PackageInternal] = fmt.Sprintf("package %s is internal", options.PackagePath)
	ErrorCodeInfos[dep.PackageInvalid] = fmt.Sprintf("package %s is invalid", options.PackagePath)
	ErrorCodeInfos[dep.PackageNotInGOPATH] = fmt.Sprintf("package not in $GOPATH directory %s", options.GOPATH)
	ErrorCodeInfos[dep.DirNotAPackage] = fmt.Sprintf("directory %s is not a package", options.PackageDir)
	ErrorCodeInfos[dep.DependancyNotInPackageDir] = fmt.Sprintf("dep files not in package %s", options.PackagePath)
	ErrorCodeInfos[dep.DependancyNotInGOPATH] = fmt.Sprintf("dep files not in $GOPATH/dep directory %s", path.Join(options.GOPATH, "dep"))
	ErrorCodeInfos[dep.DependancyInfosCorrupt] = fmt.Sprintf("dep infos are corrupt for package  %s", options.PackagePath)
	ErrorCodeInfos[dep.UpdateConflict] = "update conflict"

	app.Name = "dep"
	app.Usage = "manage go package dependancies via imports and exports"
	app.Version = "0.0.1"
	app.Action = func(c *cli.Context) {
		parseGlobalFlags(c)
		if c.Args()[0] == "intro" {
			fmt.Println(`here is an introduction into dep`)
		}
	}
	app.Commands = []cli.Command{
		/*
			{
				Name:        "store",
				Usage:       "writes dependancies to $GOPATH/src/[package]/dep.json",
				Description: furtherInformation,
				Action:      action(_store),
				Flags:       globalFlags,
			},
		*/
		{
			Name:        "lint",
			Usage:       "shows violations of the best practices $GOPATH/src/[package]",
			Description: furtherInformation,
			Flags:       globalFlags,
			Action:      action(dep.Lint),
		},
		{
			Name:        "show",
			Usage:       "shows the dependancies of $GOPATH/src/[package]",
			Description: furtherInformation,
			Flags:       globalFlags,
			Action:      action(dep.Show),
		},
		{
			Name:        "register",
			Usage:       "registers the dependancies in $GOPATH/dep.db",
			Description: furtherInformation,
			Flags:       globalFlags,
			Action:      action(dep.Register),
		},
		{
			Name:        "get",
			Usage:       "get a go package and register it",
			Description: "passes all arguments to 'go get', see 'go help get'",
			Action:      action(dep.Get),
		},
		{
			Name:        "install",
			Usage:       "installs a go package and register it",
			Description: "passes all arguments to 'go install', see 'go help get'",
			Action:      action(dep.Install),
		},
		{
			Name:        "diff",
			Usage:       "shows the differences of current dependacies to the registered dependancies in $GOPATH/dep.db",
			Description: furtherInformation,
			Flags: append(globalFlags,
				cli.BoolFlag{"local", "compare with the local dependancies in $GOPATH/src/[package]/dep.json"},
			),
			Action: action(dep.Diff),
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
				cli.BoolFlag{"keep-temp-gopath", "keep the temporary GOPATH for inspection"},
			),
			Action: action(dep.Update),
		},
		{
			Name:  "revisions",
			Usage: "write the revisions of the dependancies to $GOPATH/src/[package]/[file]",
			Flags: append(globalFlags,
				cli.StringFlag{"file", "dep-rev.json", "the file to which the definitions of the revisions is written. may be used as input for 'dep checkout'"},
				cli.BoolFlag{"stdout", "write the revisions to stdout"},
				cli.BoolFlag{"include-indirect", "include indirect dependancies (e.g. depencancies of dependancies"},
			),
			Description: furtherInformation,
			Action:      action(dep.Revisions),
		},
		{
			Name:        "checkout",
			Usage:       "checkout revisions of imported packages as defined in [file] to $GOPATH/src",
			Description: furtherInformation,
			Flags: []cli.Flag{
				cli.StringFlag{"file", "./dep-rev.json", "the file with the definitions of the revisions, must be the format returned by 'dep checkout'"},
				cli.BoolFlag{"force", "also checkout packages that are already installed"},
			},
			Action: action(dep.Checkout),
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
