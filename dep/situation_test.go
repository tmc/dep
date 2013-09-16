package dep

import (
	"github.com/metakeule/exports"
	"os"
	"path"
	"testing"
)

func init() {
	gopath := os.Getenv("GOPATH")
	exports.DefaultEnv.GOPATH = path.Join(gopath, "src", "github.com", "metakeule", "dep", "gopath")
}

func TestCompatible(t *testing.T) {
	existsDep := exports.DefaultEnv.PkgExists("dep")
	if existsDep {
		t.Errorf("package %s exists, but should not", "dep")
	}
}
