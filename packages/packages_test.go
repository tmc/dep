package packages

import (
	// "fmt"
	"github.com/metakeule/exports/typ"
	"testing"
)

var testPkg = "github.com/metakeule/dep/example/q"

func TestImports(t *testing.T) {
	pkg := Get(testPkg)
	expected := "github.com/metakeule/dep/example/p"
	if !pkg.Imports[expected] {
		t.Errorf("package %s has not import %s", testPkg, expected)
	}

	notExpected := "fmt"
	if pkg.Imports[notExpected] {
		t.Errorf("package %s has internal import %s but should not", testPkg, notExpected)
	}
}

func TestInternal(t *testing.T) {
	pkg := Get(testPkg)
	if pkg.Internal {
		t.Errorf("package %s has internal but should not", testPkg)
	}

	pkg = Get("fmt")

	if !pkg.Internal {
		t.Errorf("package fmt has not internal but should")
	}
}

func TestPath(t *testing.T) {
	pkg := Get(testPkg)

	if pkg.Path != testPkg {
		t.Errorf("package %s should has wrong path: %s", testPkg, pkg.Path)
	}
}

func TestExports(t *testing.T) {
	pkg := Get(testPkg)

	if len(pkg.Exports) != 3 {
		t.Errorf("package %s should have 3 exports, but has: %v", testPkg, len(pkg.Exports))
	}

	if pkg.Exports["A"] == nil {
		t.Errorf("package %s should have export with name A, but has none", testPkg)
	}

	if _, ok := pkg.Exports["A"].(*typ.Value); !ok {
		t.Errorf("package %s should have export A of type *typ.Value, but has %T", testPkg, pkg.Exports["A"])
	}

	if pkg.Exports["B"] == nil {
		t.Errorf("package %s should have export with name B, but has none", testPkg)
	}

	if _, ok := pkg.Exports["B"].(*typ.Value); !ok {
		t.Errorf("package %s should have export B of type *typ.Value, but has %T", testPkg, pkg.Exports["B"])
	}

	if pkg.Exports["C"] == nil {
		t.Errorf("package %s should have export with name C, but has none", testPkg)
	}

	if _, ok := pkg.Exports["C"].(*typ.StructType); !ok {
		t.Errorf("package %s should have export C of type *typ.StructType, but has %T", testPkg, pkg.Exports["C"])
	}
}
