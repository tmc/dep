package main

import (
	"bytes"
	"path/filepath"
	// "encoding/json"
	"fmt"
	"github.com/metakeule/exports"
	"io/ioutil"
	"path"
)

func getJson(pkg string) string {
	b, err := exports.DefaultEnv.Pkg(pkg).MarshalJSON()
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}

func loadJson(pkgPath string) (ø *exports.PackageJSON) {
	file, _ := filepath.Abs(path.Join(GOPATH, "src", pkgPath, "dep.json"))
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err.Error())
	}
	ø, err = exports.DefaultEnv.LoadJson(data)
	if err != nil {
		panic(err.Error())
	}
	return
}

func MapEqual(a map[string]string, b map[string]string) bool {

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

func packageDiff(old_ *exports.PackageJSON, new_ *exports.PackageJSON) string {
	var buffer bytes.Buffer
	if old_.Path != new_.Path {
		buffer.WriteString(
			fmt.Sprintf(
				"--- Path: %s\n+++ Path: %s\n",
				old_.Path,
				new_.Path))
	}

	if !MapEqual(old_.Exports, new_.Exports) {

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
