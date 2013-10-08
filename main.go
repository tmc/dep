package main

import (
	"flag"
	"github.com/go-dep/dep/depcore"
	"github.com/go-dep/gdf"
	"os"
	"path/filepath"
	"strings"
)

type _args struct {
	Verbose   bool
	Json      bool
	Yes       bool
	NoWarn    bool
	Override  string
	Panic     bool
	PkgPath   string
	SkipCheck bool
}

var (
	env  *depcore.Environment
	pkg  *gdf.Package
	Args = &_args{}
)

func initV1() {
	if os.Getenv("GOPATH") == "" {
		S.Error("GOPATH not set")
	}
	if os.Getenv("DEP_TMP") == "" {
		S.Error("DEP_TMP not set")
	}
	env = depcore.NewEnv(strings.Split(os.Getenv("GOPATH"), ";")[0])
	env.TMPDIR = os.Getenv("DEP_TMP")

	flag.BoolVar(&Args.Verbose, "verbose", false, "print details about the actions taken")
	flag.BoolVar(&Args.Json, "json", false, "print in readable json format")
	flag.BoolVar(&Args.Yes, "y", false, "answer all questions with 'yes'")
	flag.BoolVar(&Args.NoWarn, "no-warn", false, "suppress warnings")
	flag.BoolVar(&Args.Panic, "panic", false, "panic on errors")
	flag.BoolVar(&Args.SkipCheck, "skip-check", false, "skip the dep check at the beginning of dep get")
	flag.StringVar(&Args.Override, "override", "", "pass an overwrite file")
}

var cmdsWithoutPkgDir = map[string]bool{
	"backups-cleanup":  true,
	"gopath-cleanup":   true,
	"registry-cleanup": true,
	"init":             true,
	"check":            true,
	"dump":             true,
}

var cmdsWithoutExistingPkg = map[string]bool{
	"unregister": true,
	"get":        true,
}

func main() {
	flag.Usage = func() { S.Out(usage) }

	flag.Parse()
	if Args.Verbose && Args.Json {
		S.Error("-verbose and -json option are mutually exclusive")
	}

	if !Args.Panic {
		defer func() {
			if r := recover(); r != nil {
				S.Error("%v", r)
			}
		}()
	}

	if Args.Verbose {
		depcore.VERBOSE = true
		gdf.VERBOSE = true
	}

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		S.Out(usage)
	}
	cmd := args[0]

	env.Open()
	if len(env.IgnorePkgs) > 0 {
		S.Warn("ignoring packages in %s", filepath.Join(env.GOPATH, ".depignore"))
	}

	if cmd != "get" && Args.Override != "" {
		S.Error("flag -override is only for dep get command")
	}

	if cmd != "get" && Args.SkipCheck {
		S.Error("flag -skip-check is only for dep get command")
	}

	if !cmdsWithoutPkgDir[cmd] {
		if len(args) == 2 {
			Args.PkgPath = args[1]
		} else {
			Args.PkgPath = S.DefaultPackagePath()
		}

		if !cmdsWithoutExistingPkg[cmd] {
			S.Package(Args.PkgPath)
		}

	}
	defer env.Close()

	if err := runCmd(cmd); err != nil {
		S.Error(err.Error())
	}
}
