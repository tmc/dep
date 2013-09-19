package depcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metakeule/exports"
	"io/ioutil"
	"path"
	"path/filepath"
)

func (o *Environment) getJson(pkg string) string {
	b, err := json.Marshal(o.Pkg(pkg))
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}

func (o *Environment) loadJson(pkgPath string) (ø *exports.Package) {
	file, _ := filepath.Abs(path.Join(o.GOPATH, "src", pkgPath, "dep.json"))
	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err.Error())
	}
	ø, err = o.LoadJson(data)
	if err != nil {
		panic(err.Error())
	}
	return
}

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
