package main

import (
	"fmt"
	"github.com/metakeule/dep/depcore"
	"os"
)

func init() {
	if os.Getenv("GOROOT") == "" {
		exit(depcore.ErrorGOROOTNotSet)
	}
	if os.Getenv("GOPATH") == "" {
		exit(depcore.ErrorGOPATHNotSet)
	}
	if os.Getenv("DEP_TMP") == "" {
		exit(depcore.ErrorDEPTMPNotSet)
	}
}

// use it to return a specific error code
// that may be checked from the outside programm
// panics are handled in main
func exit(ø depcore.ErrorCode) {
	fmt.Printf("ERROR: %s, aborting\n", ø.Error())
	depcore.Cleanup()
	os.Exit(int(ø))
}
