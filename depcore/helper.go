package depcore

import (
	"bytes"
	"fmt"
	"github.com/metakeule/exports"
)

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
