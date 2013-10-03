package depcore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/metakeule/gdf"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type packageGetter struct {
	env       *Environment
	pkgPath   string
	pkg       *gdf.Package
	revisions map[string]string
}

func newPackageGetter(env *Environment, pkgPath string) *packageGetter {
	return &packageGetter{env: env, pkgPath: pkgPath, revisions: map[string]string{}}
}

func (get *packageGetter) repoPath(pkgPath string) string {
	dir, _, err := get.env.PkgDir(pkgPath)

	if err != nil {
		panic(err.Error())
	}

	str, err := relativePath(path.Join(get.env.GOPATH, "src"), dir)

	if err != nil {
		panic(err.Error())
	}
	return str
}

func (get *packageGetter) execGo(args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Env = get.env.cmdEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"go %s returns error: %s",
			strings.Join(args, " "),
			stdout.String()+"\n"+stderr.String(),
		)
	}
	return nil
}

// With go get -d we don't install the packages
func (get *packageGetter) goGet(pkg string) error     { return get.execGo("get", "-u", "-d", pkg) }
func (get *packageGetter) goInstall(pkg string) error { return get.execGo("install", pkg) }

// get an import with an optional revision
// the done map has repoPaths that already have been handled
func (get *packageGetter) getImport(pkgPath string, rev *revision, done map[string]bool) (err error) {
	if !get.env.PkgExists(pkgPath) {
		err = get.goGet(pkgPath)
		if err != nil {
			return
		}
	}

	repoPath := get.repoPath(pkgPath)
	// check if repo path already has been handled
	if done[repoPath] {
		return
	}

	if rev != nil {
		get.checkout(*rev)
	}

	done[repoPath] = true
	err = get.getImports(pkgPath, done)
	return
}

// gets the imports (recursively) respecting the revisions in revfile (if there is one)
// the done map has repoPaths that already have been handled, but is not used
// here (simply passed to getImport)
func (get *packageGetter) getImports(pkgPath string, done map[string]bool) (err error) {
	var pkg *gdf.Package
	pkg, err = get.env.Pkg(pkgPath)
	if err != nil {
		return err
	}
	// if there is rev file, get the imported revisions
	if get.env.HasRevFile(pkgPath) {
		var revisions map[string]revision
		revisions, err = get.trackedRevisions(pkgPath)
		if err != nil {
			return err
		}
		for imported, rev := range revisions {
			err = get.getImport(imported, &rev, done)
			if err != nil {
				return
			}
		}
		return
	}

	// if there is no rev file, get the imported packages
	// although go get -u -d installs the dependent packages
	// it might be that older revisions required other packages
	// that have not been fetched as by go get
	for imported, _ := range pkg.ImportedPackages {
		err = get.getImport(imported, nil, done)
		if err != nil {
			return
		}
	}
	return
}

// gets the main package but only a certain revision from it
func (get *packageGetter) getPkgRev(rev revision) (err error) {
	err = get.goGet(get.pkgPath)
	if err != nil {
		return
	}

	get.checkout(rev)

	// important: add own repo to repopath, so that imports of subpackages
	// of the own repo do not trigger an additional go get
	done := map[string]bool{get.repoPath(get.pkgPath): true}
	err = get.getImports(get.pkgPath, done)
	if err != nil {
		return
	}

	get.cleanupRepos(done)
	err = get.goInstall(get.pkgPath)
	return
}

// gets the main package and its imports (recursively) with the right revisions
func (get *packageGetter) get() (err error) {
	err = get.goGet(get.pkgPath)
	if err != nil {
		return
	}

	// important: add own repo to repopath, so that imports of subpackages
	// of the own repo do not trigger an additional go get
	done := map[string]bool{get.repoPath(get.pkgPath): true}
	err = get.getImports(get.pkgPath, done)
	if err != nil {
		return
	}

	get.cleanupRepos(done)
	err = get.goInstall(get.pkgPath)
	return
}

func (get *packageGetter) scanRepos(pkgPath string, repos map[string]bool) {
	repos[get.repoPath(pkgPath)] = true
	pkg := get.env.MustPkg(pkgPath)
	for imported, _ := range pkg.ImportedPackages {
		get.scanRepos(imported, repos)
	}
}

// remove old imports that we got via goGet but that aren't needed
// after switching revisions
func (get *packageGetter) cleanupRepos(done map[string]bool) {
	neededRepos := map[string]bool{}
	get.scanRepos(get.pkgPath, neededRepos)
	for r, _ := range done {
		if !neededRepos[r] {
			os.RemoveAll(path.Join(get.env.GOPATH, "src", r))
		}
	}
}

// checkous revision rev of directory d
func (get *packageGetter) checkout(rev revision) {
	var checkoutErr error
	switch rev.VCM {
	case "bzr":
		checkoutErr = get.checkoutBzr(rev.RepoRoot, rev.Rev)
	case "git":
		checkoutErr = get.checkoutGit(rev.RepoRoot, rev.Rev)
	case "hg":
		checkoutErr = get.checkoutHg(rev.RepoRoot, rev.Rev)
	case "svn":
		panic("unsupported VCM svn for repository " + rev.RepoRoot)
	default:
		panic("unsupported VCM " + rev.VCM + " for repository " + rev.RepoRoot)
	}

	if checkoutErr != nil {
		panic("can't checkout " + rev.RepoRoot + " rev " + rev.Rev + ":\n" + checkoutErr.Error())
	}
}

// reads the tracked revisions for imports as defined in the revFile
func (get *packageGetter) trackedRevisions(pkg string) (r map[string]revision, err error) {
	r = map[string]revision{}
	data, e := ioutil.ReadFile(path.Join(get.env.GOPATH, "src", pkg, revFileName))
	if e != nil {
		return
	}
	err = json.Unmarshal(data, &r)
	return
}

// does a checkout for Bzr VCM
func (get *packageGetter) checkoutBzr(dir string, rev string) error {
	return get.checkoutCmd(dir, "bzr", "update", "-r", rev)
}

// does a checkout for Git VCM
func (get *packageGetter) checkoutGit(dir string, rev string) error {
	return get.checkoutCmd(dir, "git", "checkout", rev)
}

// does a checkout for Hg VCM
func (get *packageGetter) checkoutHg(dir string, rev string) error {
	return get.checkoutCmd(dir, "hg", "update", "-r", rev)
}

func (get *packageGetter) checkoutCmd(dir string, c string, args ...string) error {
	cmd := exec.Command(c, args...)
	cmd.Dir = dir
	cmd.Env = get.env.cmdEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s", stderr.String())
	}
	return nil
}
