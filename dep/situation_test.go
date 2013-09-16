package dep

import (
	"github.com/metakeule/exports"
	"os"
	"path"
	"runtime"
	"testing"
)

var opt = &Options{}

func init() {
	opt.GOPATH = path.Join(os.Getenv("GOPATH"), "src", "github.com", "metakeule", "dep", "gopath")
	opt.GOROOT = runtime.GOROOT()
	opt.Env = exports.NewEnv(opt.GOROOT, opt.GOPATH)
}

func TestCompatible(t *testing.T) {
	//pkg := opt.Env.Pkg("github.com/metakeule/deptest_situation/working/compatible_change")

	//dir, _ := pkg.Dir()
	//goget(opt, "github.com/metakeule/deptest_situation/working/compatible_change")

	dir := path.Join(opt.GOPATH, "src", "github.com/metakeule/deptest_situation/working/compatible_change")
	rev := revision{}
	rev.VCM = "git"
	rev.Rev = "d5cf87993ac713a5aef12e39abf08542f3187587"
	checkoutRevision(opt, dir, rev)

	dir = path.Join(opt.GOPATH, "src", "github.com/metakeule/deptest_mod_a/compatible")
	rev = revision{}
	rev.VCM = "git"
	rev.Rev = "4976ff78ef93be4c04774ba18bffdbf2a96c5cc0"
	checkoutRevision(opt, dir, rev)

	/*
		existsDep := opt.Env.PkgExists("dep")
		if existsDep {
			t.Errorf("package %s exists, but should not", "dep")
		}
	*/
}
