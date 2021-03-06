package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type _sugar interface {
	Error(string, ...interface{})
	Out(string, ...interface{})
	Warn(string, ...interface{})
	DefaultPackagePath() string
	Package(pkgPath string)
	Ask(question string) bool
	ExitUnless(question string)
	Json(i interface{})
	PrintConflicts(conflicts map[string]map[string][3]string)
	ReadLines(reader io.Reader) ([]string, error)
	PrintChanged(changed map[string][2]string)
}

var S _sugar = sugar(0)

type sugar int

// readLines reads a whole file into memory
// and returns a slice of its lines.
func (s sugar) ReadLines(in io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writes an error with a linefeed and exits
func (s sugar) Error(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+" aborting\n", vals...)
	os.Exit(1)
}

// writes output with a linefeed and exits
func (s sugar) Out(format string, vals ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", vals...)
	os.Exit(0)
}

func (s sugar) Warn(format string, vals ...interface{}) {
	if !Args.NoWarn {
		fmt.Fprintf(os.Stderr, "--- Warning: "+format+" ---\n\n", vals...)
	}
}

func (s sugar) DefaultPackagePath() string {
	dir, err := os.Getwd()
	if err != nil {
		s.Error("can't get working directory: " + err.Error())
	}
	dir, err = filepath.Abs(dir)
	return env.PkgPath(dir)
}

func (s sugar) Package(pkgPath string) {
	var err error
	pkg, err = env.GetPkg(pkgPath)
	if err != nil {
		panic(err.Error())
	}
	return
}

func (s sugar) PrintChanged(changed map[string][2]string) {
	fmt.Println("The following repos have been changed: (old revision => new revision)")
	for repo, v := range changed {
		fmt.Printf("\n\t%s %s => %s", repo, v[0], v[1])
	}
	fmt.Println("\nDon't forget to run 'go test ./... in them. If you wan't to see what changed, run 'dep check'. To register the changed packages run 'dep register' for each package until 'dep check' complaints no more.\n")
}

func (s sugar) Json(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		s.Error(err.Error())
	}
	fmt.Fprintf(os.Stdout, "%s\n", b)
}

func (s sugar) Ask(question string) bool {
	if Args.Yes {
		return true
	}
	fmt.Println(question + " (y/n)?")
	answer := ""
	_, err := fmt.Scanln(&answer)
	if err != nil {
		s.Error(err.Error())
	}
	answer = strings.ToLower(answer)
	answer = strings.TrimSpace(answer)
	if answer == "y" || answer == "yes" {
		return true
	}
	return false
}

func (s sugar) ExitUnless(question string) {
	if !S.Ask(question) {
		fmt.Println("aborted")
		os.Exit(1)
	}
}

func (s sugar) PrintConflicts(conflicts map[string]map[string][3]string) {
	if len(conflicts) > 0 {
		fmt.Println("#####     ERROR     #####")
	}
	for k, v := range conflicts {
		if k == "#dep-registry-orphan#" {
			for kk, _ := range v {
				fmt.Printf("orphaned package in registry: %s\n", kk)
			}
			continue
		}
		for kk, vv := range v {
			if kk == "#dep-registry-inconsistency#" {
				switch vv[0] {
				case "missing":
					fmt.Printf("missing package in registry: %s\n", k)
				case "imports":
					fmt.Printf("the imports of %s have changed:\n\n%s\n", k, vv[1])
				case "exports":
					fmt.Printf("the exports of %s have changed:\n\n%s\n", k, vv[1])
				default:
					s.Error("unknown #dep-registry-inconsistency# key: " + vv[0])
				}
				continue
			}

			switch vv[0] {
			case "error":
				fmt.Printf("can't get packages that are importing symbols from %s, error: %s\n", kk, vv[1])
			case "changed":
				imA := strings.Split(kk, ":")
				imP, imName := imA[0], imA[1]
				fmt.Printf("\n\npackage %#v\nimports symbol %#v\nof %#v\n\n   but that changed from\n\n%s\n\n     to\n\n%s\n", imP, imName, k, vv[1], vv[2])
			case "removed":
				imA := strings.Split(kk, ":")
				imP, imName := imA[0], imA[1]
				fmt.Printf("\n\npackage %#v\nimports symbol %#v\nof %#v\n\n   that now does no longer exists, but was\n\n%s\n", imP, imName, k, vv[1])
			default:
				s.Error("unknown conflict type " + vv[0])
			}

		}
	}
}
