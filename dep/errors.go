package dep

import (
// "fmt"
// "github.com/codegangsta/cli"
// "github.com/metakeule/exports"
// "os"
// "path"
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
	UpdateConflict
)

var _cleanup = []func(){}

func addCleanup(fn func()) {
	_cleanup = append(_cleanup, fn)
}

// reverse executing all cleanups
func Cleanup() {
	for i := len(_cleanup) - 1; i >= 0; i-- {
		_cleanup[i]()
	}
}
