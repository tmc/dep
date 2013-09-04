package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/metakeule/exports"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

var _ = fmt.Print

func pkgJson(path string) (b []byte, internal bool) {
	p := exports.Get(path)
	internal = p.Internal
	var err error
	b, err = json.MarshalIndent(p, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return
}

func scan(dir string) (b []byte, internal bool) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Println(dir)
	b, internal = pkgJson(exports.PkgPath(dir))
	b = append(b, []byte("\n")...)
	return
}

func checksum(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// for temporary installations
var depPATH = path.Join(os.Getenv("HOME"), ".dep")

var goPATH = os.Getenv("GOPATH")
var goROOT = os.Getenv("GOROOT")

// for registry files
var depRegistry = path.Join(goPATH, "dep")

// for registry files of core libraries
var depRegistryRoot = path.Join(depPATH, goROOT, "dep")

// TODO: initialize the dependancies for the core libs on the first start
// and if GOROOT has changed

func readRegisterFile(dir string, internal bool) (*exports.Package, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	trimmedPath := exports.PkgPath(dir)
	// homedir := os.Getenv("HOME")
	registerPath := path.Join(depRegistry, trimmedPath)

	if internal {
		registerPath = path.Join(depRegistryRoot, trimmedPath)
	}

	registerPath, _ = filepath.Abs(registerPath)
	file := path.Join(registerPath, "dep.json")
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	p := &exports.Package{}
	err = json.Unmarshal(b, p)
	return p, err
}

func writeRegisterFile(dir string, data []byte, internal bool) error {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	trimmedPath := exports.PkgPath(dir)
	// homedir := os.Getenv("HOME")
	registerPath := path.Join(depRegistry, trimmedPath)

	if internal {
		registerPath = path.Join(depRegistryRoot, trimmedPath)
	}

	registerPath, _ = filepath.Abs(registerPath)
	// fmt.Println(registerPath)
	err = os.MkdirAll(registerPath, 0755)
	if err != nil {
		return err
	}
	file := path.Join(registerPath, "dep.json")
	fmt.Printf("writing %s\n", file)
	err = ioutil.WriteFile(file, data, 0644)
	if err != nil {
		return err
	}

	chk := checksum(string(data))
	file = path.Join(registerPath, "dep.md5")
	fmt.Printf("writing %s\n", file)
	return ioutil.WriteFile(file, []byte(chk), 0644)
}

func writeDepFile(dir string, data []byte) error {
	file := path.Join(dir, "dep.json")
	f, _ := filepath.Abs(file)
	fmt.Printf("writing %s\n", f)
	return ioutil.WriteFile(file, data, 0644)
}

var commands = map[string]*flag.FlagSet{
	"info":     flag.NewFlagSet("info", flag.ContinueOnError),
	"init":     flag.NewFlagSet("init", flag.ContinueOnError),
	"register": flag.NewFlagSet("register", flag.ContinueOnError),
}

var commandHandles = map[string]func(*flag.FlagSet){
	"info": func(fs *flag.FlagSet) {
		dir := fs.Arg(0)
		b, _ := scan(dir)
		fmt.Printf("%s", b)
	},

	"init": func(fs *flag.FlagSet) {
		dir := fs.Arg(0)
		b, internal := scan(dir)
		if internal {
			panic("can't init internal package " + dir)
		}
		err := writeDepFile(dir, b)
		if err != nil {
			panic(err.Error())
		}
	},
	"register": func(fs *flag.FlagSet) {
		dir := fs.Arg(0)
		b, internal := scan(dir)
		err := writeRegisterFile(dir, b, internal)
		if err != nil {
			panic(err.Error())
		}
	},
}

func init() {
	if os.Getenv("HOME") == "" {
		panic("HOME environment variable not set")
	}

	if os.Getenv("GOPATH") == "" {
		panic("GOPATH environment variable not set")
	}

	if os.Getenv("GOROOT") == "" {
		panic("GOROOT environment variable not set")
	}
	//Init.Args()
}

func main() {
	flag.Parse()
	args := flag.Args()
	//fmt.Println(args)
	cmd := args[0]

	c, ok := commands[cmd]
	if !ok {
		fmt.Println("unkown command " + cmd)
		flag.Usage()
		os.Exit(0)
	}

	c.Parse(args[1:])
	commandHandles[cmd](c)

	//fmt.Println(args)
	//fmt.Println(Scan.Arg(0))
	// fmt.Println(initArgs)
	//fmt.Println(string(pkgJson(pkg)))
}
