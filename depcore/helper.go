package depcore

import (
	"bytes"
	"fmt"
	// "github.com/metakeule/dep/db"
	"encoding/json"
	"github.com/metakeule/exports"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func toJson(i interface{}) []byte {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return b
}

var bzrRevRe = regexp.MustCompile(`revision-id:\s*([^\s]+)`)

// maps a package path to a vcs and a revision
type revision struct {
	VCM    string
	Rev    string
	Parent string
	Tag    string // TODO check if revision is a tag and put it into the rev
}

var revFileName = "dep-rev.json"

func _repoRoot(dir string) string {
	_, root, err := vcsForDir(dir)
	if err != nil {
		panic("can't find repodir for " + dir + " : " + err.Error())
	}
	return root
}

// checks if two maps are equal
func mapEqual(a map[string]string, b map[string]string) bool {
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	for k, _ := range b {
		_, exists := a[k]
		if !exists {
			return false
		}
	}
	return true
}

// makes a diff for two packages
func packageDiff(old_ *exports.Package, new_ *exports.Package) string {
	var buffer bytes.Buffer
	if old_.Path != new_.Path {
		buffer.WriteString(
			fmt.Sprintf(
				"--- Path: %s\n+++ Path: %s\n",
				old_.Path,
				new_.Path))
	}

	if !mapEqual(old_.Exports, new_.Exports) {
		visited := map[string]bool{}
		for old_key, old_val := range old_.Exports {
			visited[old_key] = true
			new_val, ok := new_.Exports[old_key]
			if !ok {
				buffer.WriteString(fmt.Sprintf("--- Exports: %s: %s\n", old_key, old_val))
				continue
			}
			if old_val != new_val {
				buffer.WriteString(
					fmt.Sprintf(
						"--- Exports: %s: %s\n+++ Exports: %s: %s\n",
						old_key, old_val, old_key, new_val))
			}
		}

	}
	return buffer.String()
}

func getmasterRevision(pkg string, dir string) string {
	cmd := exec.Command("git", "rev-parse", "master")
	cmd.Env = []string{
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}
	cmd.Dir = dir
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		panic(stdout.String() + "\n" + stderr.String())
	}
	return strings.Trim(stdout.String(), "\n\r")
}
