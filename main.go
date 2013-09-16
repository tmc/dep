package main

import (
	"fmt"
	"github.com/metakeule/dep/dep"
	"os"
)

func init() {
	if os.Getenv("GOROOT") == "" {
		exit(dep.GOROOTNotSet)
	}
	if os.Getenv("GOPATH") == "" {
		exit(dep.GOPATHNotSet)
	}
	if os.Getenv("HOME") == "" {
		exit(dep.HOMENotSet)
	}
}

// use it to return a specific error code
// that may be checked from the outside programm
// panics are handled in main
func exit(ø dep.ErrorCode) {
	fmt.Printf("ERROR: %s, aborting\n", ø)
	dep.Cleanup()
	os.Exit(int(ø))
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
