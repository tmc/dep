package dep

import (
	"bytes"
	_ "code.google.com/p/go.exp/inotify"
	"database/sql"
	"encoding/json"
	"fmt"
	// "github.com/metakeule/cli"
	// "path/filepath"
	// "github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	"io/ioutil"
	_ "launchpad.net/goamz/aws"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

type Options struct {
	All         bool
	Recursive   bool
	PackagePath string
	Package     *exports.Package
	PackageDir  string
	GOPATH      string
	GOROOT      string
	HOME        string
	DEP         string
	Env         *exports.Environment
}

var _ = fmt.Printf

/*
func _store(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	pkgs := packages()

	for _, pkg := range pkgs {
		b := asJson(pkg)
		b = append(b, []byte("\n")...)

		dir := path.Join(GOPATH, "src", pkg.Path)
		err := writeDepFile(dir, b)
		if err != nil {
			panic(err.Error())
		}
	}
	return 0
}
*/

func _registerPackage(o *Options, pkgMap map[string]*db.Pkg, pkg *exports.Package) (dbExps []*db.Exp, dbImps []*db.Imp) {
	p := &db.Pkg{}
	p.Package = pkg.Path
	p.Json = asJson(pkg)
	pkgMap[pkg.Path] = p
	dbExps = []*db.Exp{}
	dbImps = []*db.Imp{}

	for im, _ := range pkg.Imports {
		if _, has := pkgMap[im]; has {
			continue
		}
		imPkg := o.Env.Pkg(im)
		pExp, pImp := _registerPackage(o, pkgMap, imPkg)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	pkgjs := pkg.PackageJSON()

	for k, v := range pkgjs.Exports {
		dbE := &db.Exp{}
		dbE.Package = pkg.Path
		dbE.Name = k
		dbE.Value = v
		dbExps = append(dbExps, dbE)
	}

	for k, v := range pkgjs.Imports {
		dbI := &db.Imp{}
		dbI.Package = pkg.Path
		arr := strings.Split(k, "#")
		dbI.Name = arr[1]
		dbI.Value = v
		dbI.Import = arr[0]
		dbImps = append(dbImps, dbI)
	}
	return
}

func registerPackages(o *Options, dB *sql.DB, pkgs ...*exports.Package) {
	dbExps := []*db.Exp{}
	dbImps := []*db.Imp{}
	pkgMap := map[string]*db.Pkg{}

	for _, pkg := range pkgs {
		pExp, pImp := _registerPackage(o, pkgMap, pkg)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	dbPkgs := []*db.Pkg{}

	for _, dbPgk := range pkgMap {
		dbPkgs = append(dbPkgs, dbPgk)
	}

	err := db.InsertPackages(dB, dbPkgs, dbExps, dbImps)
	if err != nil {
		panic(err.Error())
	}

	/*
		for _, pk := range dbPkgs {
			fmt.Println("registered: ", pk.Package)
		}
	*/
}

func toJson(i interface{}) []byte {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return b
}

func mkdirTempDir(o *Options) (tmpGoPath string) {
	depPath := path.Join(o.HOME, ".dep")
	fl, err := os.Stat(depPath)
	if err != nil {
		err = os.MkdirAll(depPath, 0755)
		if err != nil {
			panic(err.Error())
		}
	}
	if !fl.IsDir() {
		panic(depPath + " is a file. but should be a directory")
	}

	tmpGoPath, err = ioutil.TempDir(depPath, "gopath_")
	if err != nil {
		panic(err.Error())
	}
	err = os.Mkdir(path.Join(tmpGoPath, "src"), 0755)
	if err != nil {
		panic(err.Error())
	}
	err = os.Mkdir(path.Join(tmpGoPath, "bin"), 0755)
	if err != nil {
		panic(err.Error())
	}
	err = os.Mkdir(path.Join(tmpGoPath, "pkg"), 0755)
	if err != nil {
		panic(err.Error())
	}
	return
}

/*
	TODO
	- check for all packages within the repoDir
	- check for package and all their dependancies, if
		- their dependant packages would be fine with the new exports
	- if all is fine, move the missing src entries to the real GOPATH and install the package
	- if there are some conflicts, show them
*/

// TODO: add verbose flag for verbose output
func hasConflict(dB *sql.DB, p *exports.Package) (errors map[string][3]string) {
	pkg := p.PackageJSON()
	imp, err := db.GetImported(dB, pkg.Path)
	if err != nil {
		panic(err.Error())
	}

	errors = map[string][3]string{}

	for _, im := range imp {
		key := fmt.Sprintf("%s: %s", im.Package, im.Name)

		if val, exists := pkg.Exports[im.Name]; exists {
			if val != im.Value {
				errors[key] = [3]string{"changed", im.Value, val} //fmt.Sprintf("%s will change as required from %s (was %s, would be %s)", im.Name, im.Package, im.Value, val)
			}
			continue
		}
		errors[key] = [3]string{"removed", im.Value, ""} // fmt.Sprintf("%s would be missing, required by %s", im.Name, im.Package)
	}
	return
}

func checkConflicts(o *Options, dB *sql.DB, tempEnv *exports.Environment, pkg *exports.Package) (errs map[string][3]string) {
	tempPkg := tempEnv.Pkg(pkg.Path)
	if !o.Env.PkgExists(pkg.Path) {
		panic(fmt.Sprintf("package %s is not installed"))
	}

	registerPackages(o, dB, pkg)
	errs = hasConflict(dB, tempPkg)
	return
}

func runGoTest(o *Options, tmpDir string, dir string) ([]byte, error) {
	cmd := exec.Command("go", "test")
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, tmpDir),
		fmt.Sprintf(`GOROOT=%s`, o.GOROOT),
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// maps a package path to a tag
//type tags map[string]string

//var tagFileName = "dep-tags.json"

/*
	better take tags instead of revisions
	(for git revisions might be used as well)

	git tags => show all tags
	bzr tags => show all tags
	hg tags => show all tags
	svn branches lying around in repo/tags
*/

/*
git rev-parse HEAD
svn info | grep "Revision" | awk '{print $2}'
bzr log -r last:1 | grep revno
//hg tags | grep "tip" | awk '{print $2}'
//hg log -r tip (erste)
// better:
hg tip --template '{node}'
*/

func getRevCmd(o *Options, dir string, c string, args ...string) string {
	cmd := exec.Command(c, args...)
	cmd.Dir = dir
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, o.GOPATH),
		fmt.Sprintf(`GOROOT=%s`, o.GOROOT),
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}

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

func getRevisionGit(o *Options, dir string) string {
	return getRevCmd(o, dir, "git", "rev-parse", "HEAD")
}

func getRevisionHg(o *Options, dir string) string {
	return getRevCmd(o, dir, "hg", "tip", "--template", "{node}")
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

func getRevisionBzr(o *Options, dir string) string {
	res := getRevCmd(o, dir, "bzr", "log", "-l", "1", "--show-ids")
	sm := bzrRevRe.FindAllStringSubmatch(res, 1)
	return sm[0][1]
}

func pkgRevision(o *Options, dir string, parent string) (rev revision) {
	//dir := path.Join(exports.DefaultEnv.GOPATH, "src", pkgPath)
	vcs, root, err := vcsForDir(dir)
	if err != nil {
		panic(err.Error())
	}
	_ = root
	var r string
	switch vcs.cmd {
	case "git":
		r = getRevisionGit(o, dir)
	case "hg":
		r = getRevisionHg(o, dir)
	case "bzr":
		r = getRevisionBzr(o, dir)
	case "svn":
		panic("svn is currently not supported")
	default:
		panic("unknown vcs command " + vcs.cmd)

	}
	return revision{vcs.cmd, r, parent, ""}
}

func indirectRev(o *Options, revisions map[string]revision, pkg *exports.Package, parent string) {
	for im, _ := range pkg.Imports {
		if _, has := revisions[im]; !has {
			revisions[im] = pkgRevision(o, path.Join(o.GOPATH, "src", im), pkg.Path)
			indirectRev(o, revisions, o.Env.Pkg(im), pkg.Path)
		}
	}
}

func checkoutRevCmd(o *Options, dir string, c string, args ...string) error {
	cmd := exec.Command(c, args...)
	cmd.Dir = dir
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, o.GOPATH),
		fmt.Sprintf(`GOROOT=%s`, o.GOROOT),
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s", stderr)
	}
	return nil
}

func checkoutBzr(o *Options, dir string, rev string) error {
	// update -r {tag}
	return checkoutRevCmd(o, dir, "bzr", "update", "-r", rev)
}

func checkoutGit(o *Options, dir string, rev string) error {
	return checkoutRevCmd(o, dir, "git", "checkout", rev)
}

func checkoutHg(o *Options, dir string, rev string) error {
	// update -r
	return checkoutRevCmd(o, dir, "hg", "update", "-r", rev)
}

func _repoRoot(dir string) string {
	_, root, err := vcsForDir(dir)
	if err != nil {
		panic(err.Error())
	}
	return root
}

/*
func _fix(c *cli.Context) ErrorCode {
	return 0
}
*/
