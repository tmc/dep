package depcore

import (
	"bytes"
	"fmt"
	// "github.com/go-dep/dep/db"
	"encoding/json"
	"github.com/go-dep/gdf"
	"os"
	"os/exec"
	"path"
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
	VCM      string
	Rev      string
	Parent   string
	Tag      string // TODO check if revision is a tag and put it into the rev
	RepoRoot string
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
func packageDiff(old_ *gdf.Package, new_ *gdf.Package) string {
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

type testEnv struct{ inner *Environment }

func NewTestEnv() *testEnv {
	t := &testEnv{
		NewEnv(
			path.Join(
				os.Getenv("GOPATH"),
				"src",
				"github.com",
				"go-dep",
				"dep",
				"gopath"))}
	t.prepare()
	return t
}

func (env *testEnv) prepare() {
	os.RemoveAll(env.inner.GOPATH)
	os.MkdirAll(env.inner.GOPATH, 0755)
	env.inner.Open()
}

func (env *testEnv) Get(pkg, rev string) error {
	g := newPackageGetter(env.inner, pkg)
	r := revision{}
	r.RepoRoot = g.repoPath(pkg)
	r.Rev = rev
	r.VCM = "git"
	return g.getPkgRev(r)
}

func (ev *testEnv) Update(pkg, rev string) (changed map[string][2]string, err error) {
	defer ev.inner.Close()
	env := ev.inner

	g := newPackageGetter(env, pkg)

	//err := g.getPkgRev(revision{VCM: "git", Rev: rev, RepoRoot: g.repoPath(pkg)})
	//err := g.getImport(pkg, rev, map[string]bool{})
	err = g.getByRev(pkg, rev, "git")

	if err != nil {
		panic(err.Error())
	}

	var dir string
	var internal bool
	dir, internal, err = env.PkgDir(pkg)
	if err != nil {
		panic(err.Error())
	}
	if internal {
		panic(fmt.Sprintf("can't update internal package %s", pkg))
	}
	master := getmasterRevision(pkg, dir)

	// check, if revisions are correct
	if env.getRevisionGit(path.Join(env.GOPATH, "src", pkg)) != rev {
		panic(fmt.Sprintf("revision %#v not checked out for package %#v\n", rev, pkg))
	}

	depsBefore, eb := g.trackedRevisions(pkg)
	if eb != nil {
		panic(eb.Error())
	}

	for d, drev := range depsBefore {
		if r := env.getRevisionGit(path.Join(env.GOPATH, "src", d)); r != drev.Rev {
			panic(fmt.Sprintf("revision before update %#v not checked out, expected: %#v for dependancy package %#v\n", r, drev.Rev, d))
		}
	}

	conflicts := env.Init()
	if len(conflicts) > 0 {
		data, _ := json.MarshalIndent(conflicts, "", "  ")
		fmt.Printf("%s\n", data)
		panic(fmt.Sprintf("GOPATH %s is not integer", env.GOPATH))
	}

	changed, e := env.db.updatePackage(pkg, func(candidates ...*gdf.Package) bool {
		return true
	})
	if e != nil {
		//		fmt.Printf("normal error in updating package\n")
		err = e
		return
	}

	pdir, pinternal, perr := env.PkgDir(pkg)
	if perr != nil {
		panic(fmt.Sprintf("can't update package %s: %s", pkg, perr))
	}

	if pinternal {
		panic(fmt.Sprintf("can't update internal package %s", pkg))
	}

	if r := env.getRevisionGit(pdir); r != master {
		err = fmt.Errorf("revision after update %#v not matching master: %#v in package %#v\n", r, master, pkg)
		return
	}
	depsAfter, e := g.trackedRevisions(pkg)
	if e != nil {
		panic(e.Error())
	}

	for d, drev := range depsAfter {
		if r := env.getRevisionGit(path.Join(env.GOPATH, "src", d)); r != drev.Rev {
			err = fmt.Errorf("revision after update %#v not matching expected: %#v for dependancy package %#v\n", r, drev.Rev, d)
			return
		}
	}

	return
}
