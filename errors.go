package main

import (
	"fmt"
	// "github.com/codegangsta/cli"
	// "github.com/metakeule/exports"
	"os"
	"path"
)

type ErrorCode int

const (
	_                      = iota
	_                      = iota
	GOROOTNotSet ErrorCode = iota
	GOPATHNotSet
	HOMENotSet
	GOPATHInvalid
	InvalidOptions
	PackageInternal
	PackageInvalid
	PackageNotInGOPATH
	DirNotAPackage
	DependancyNotInPackageDir
	DependancyNotInGOPATH
	DependancyInfosCorrupt
)

var (
	HOME   = os.Getenv("HOME")
	GOPATH = os.Getenv("GOPATH")
	GOROOT = os.Getenv("GOROOT")
	// the path of the current package
	PACKAGE_PATH string
	// the physical path of the package
	PACKAGE_DIR string
	// are all packages affected
	ALL       bool
	RECURSIVE bool
)

func init() {
	if GOROOT == "" {
		exit(GOROOTNotSet)
	}
	if GOPATH == "" {
		exit(GOPATHNotSet)
	}
	if HOME == "" {
		exit(HOMENotSet)
	}
}

var ErrorCodeInfos = map[ErrorCode]string{
	GOROOTNotSet:              "$GOROOT environment variable is not set",
	GOPATHNotSet:              "$GOPATH environment variable is not set",
	HOMENotSet:                "$HOME environment variable is not set",
	GOPATHInvalid:             fmt.Sprintf("$GOPATH directory %s is not conforming to the standard layout (bin,pkg,src directories)", GOPATH),
	InvalidOptions:            "given options are invalid",
	PackageInternal:           fmt.Sprintf("package %s is internal", PACKAGE_PATH),
	PackageInvalid:            fmt.Sprintf("package %s is invalid", PACKAGE_PATH),
	PackageNotInGOPATH:        fmt.Sprintf("package not in $GOPATH directory %s", GOPATH),
	DirNotAPackage:            fmt.Sprintf("directory %s is not a package", PACKAGE_DIR),
	DependancyNotInPackageDir: fmt.Sprintf("dep files not in package %s", PACKAGE_PATH),
	DependancyNotInGOPATH:     fmt.Sprintf("dep files not in $GOPATH/dep directory %s", path.Join(GOPATH, "dep")),
	DependancyInfosCorrupt:    fmt.Sprintf("dep infos are corrupt for package  %s", PACKAGE_PATH),
}

func (ø ErrorCode) String() string {
	return ErrorCodeInfos[ø]
}

var _cleanup = []func(){}

func addCleanup(fn func()) {
	_cleanup = append(_cleanup, fn)
}

// reverse executing all cleanups
func cleanup() {
	for i := len(_cleanup) - 1; i >= 0; i-- {
		_cleanup[i]()
	}
}

// use it to return a specific error code
// that may be checked from the outside programm
// panics are handled in main
func exit(ø ErrorCode) {
	fmt.Printf("ERROR: %s, aborting\n", ø)
	cleanup()
	os.Exit(int(ø))
}
