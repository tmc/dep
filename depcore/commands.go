package depcore

import (
	"bytes"
	_ "code.google.com/p/go.exp/inotify"
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

/*
type Options1 struct {
	All         bool
	Recursive   bool
	PackagePath string
	Package     *exports.Package
	PackageDir  string
	GOPATH      string
	GOROOT      string
	TMPDIR      string
	DEP         string
	Env         *Environment
}
*/

type DB struct {
	*db.DB
	Environment *Environment
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
func (dB *DB) hasConflict(p *exports.Package) (errors map[string][3]string) {
	pkg := p
	//fmt.Printf("\n\n\n\nchecking conflict for package %#v\n", p.Path)
	imp, err := db.GetImported(dB.DB, pkg.Path)
	if err != nil {
		panic(err.Error())
	}
	// fmt.Printf("package %s in imported from %v places\n", p.Path, len(imp))

	errors = map[string][3]string{}

	for _, im := range imp {
		key := fmt.Sprintf("%s: %s", im.Package, im.Name)
		// fmt.Printf("package %#v imports: %s\n", im.Package, im.Name)
		if val, exists := pkg.Exports[im.Name]; exists {
			// fmt.Printf("\n\ta: %#v\n\tb: %#v\n", val, im.Value)
			if val != im.Value {
				errors[key] = [3]string{"changed", im.Value, val} //fmt.Sprintf("%s will change as required from %s (was %s, would be %s)", im.Name, im.Package, im.Value, val)
			}
			continue
		}
		errors[key] = [3]string{"removed", im.Value, ""} // fmt.Sprintf("%s would be missing, required by %s", im.Name, im.Package)
	}
	return
}

type Environment struct {
	*exports.Environment
	TMPDIR string
	DB     *DB
}

// environment during a tentative update
type TentativEnvironment struct {
	*Environment
	Original *Environment
}

func (env *Environment) NewTentative() *TentativEnvironment {
	t := &TentativEnvironment{}
	t.Original = env
	t.Environment = NewEnv(env.mkdirTempDir())
	return t
}

func (tempEnv *TentativEnvironment) checkConflicts(pkg *exports.Package) (errs map[string][3]string) {
	tempPkg := tempEnv.Pkg(pkg.Path)
	if !tempEnv.Original.PkgExists(pkg.Path) {
		panic(fmt.Sprintf("package %s is not installed in %s", pkg.Path, tempEnv.Original.GOPATH))
	}
	errs = tempEnv.DB.hasConflict(tempPkg)
	// errs = hasConflict(dB, tempPkg)
	return
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

func (env *Environment) _registerPackage(pkgMap map[string]*db.Pkg, pkg *exports.Package) (dbExps []*db.Exp, dbImps []*db.Imp) {
	p := &db.Pkg{}
	p.Package = pkg.Path
	//fmt.Printf("registering %s\n", pkg.Path)
	p.Json = asJson(pkg)
	pkgMap[pkg.Path] = p
	dbExps = []*db.Exp{}
	dbImps = []*db.Imp{}

	for im, _ := range pkg.ImportedPackages {
		if _, has := pkgMap[im]; has {
			continue
		}
		imPkg := env.Pkg(im)
		pExp, pImp := env._registerPackage(pkgMap, imPkg)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	pkgjs := pkg

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

func createDB(gopath string) error {
	_, dbFileerr := os.Stat(dEP(gopath))
	dB, err := db.Open(dEP(gopath))
	if err != nil {
		return err
	}
	defer dB.Close()
	if dbFileerr != nil {
		//fmt.Println(dbFileerr)
		db.CreateTables(dB)
	}
	return nil
}

func (dB *DB) registerPackages(pkgs ...*exports.Package) {
	dbExps := []*db.Exp{}
	dbImps := []*db.Imp{}
	pkgMap := map[string]*db.Pkg{}

	for _, pkg := range pkgs {
		pExp, pImp := dB.Environment._registerPackage(pkgMap, pkg)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	dbPkgs := []*db.Pkg{}
	// fmt.Printf("register packages for db %#v\n", dB.File)

	for _, dbPgk := range pkgMap {
		// fmt.Printf("register %#v\n", dbPgk.Package)
		dbPkgs = append(dbPkgs, dbPgk)
	}

	err := db.InsertPackages(dB.DB, dbPkgs, dbExps, dbImps)
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

func (o *Environment) mkdirTempDir() (tmpGoPath string) {
	depPath := o.TMPDIR
	fl, err := os.Stat(depPath)
	if err != nil {
		err = os.MkdirAll(depPath, 0755)
		if err != nil {
			panic(err.Error())
		}
		fl, err = os.Stat(depPath)
	}
	if err != nil {
		panic(err.Error())
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

func runGoTest(o *Environment, tmpDir string, dir string) ([]byte, error) {
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

func (o *Environment) getRevCmd(dir string, c string, args ...string) string {
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

func (o *Environment) getRevisionGit(dir string) string {
	return o.getRevCmd(dir, "git", "rev-parse", "HEAD")
}

func (o *Environment) getRevisionHg(dir string) string {
	return o.getRevCmd(dir, "hg", "tip", "--template", "{node}")
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

func (o *Environment) getRevisionBzr(dir string) string {
	res := o.getRevCmd(dir, "bzr", "log", "-l", "1", "--show-ids")
	sm := bzrRevRe.FindAllStringSubmatch(res, 1)
	return sm[0][1]
}

func (o *Environment) pkgRevision(dir string, parent string) (rev revision) {
	//dir := path.Join(exports.DefaultEnv.GOPATH, "src", pkgPath)
	vcs, root, err := vcsForDir(dir)
	if err != nil {
		panic(err.Error())
	}
	_ = root
	var r string
	switch vcs.cmd {
	case "git":
		r = o.getRevisionGit(dir)
	case "hg":
		r = o.getRevisionHg(dir)
	case "bzr":
		r = o.getRevisionBzr(dir)
	case "svn":
		panic("svn is currently not supported")
	default:
		panic("unknown vcs command " + vcs.cmd)

	}
	return revision{vcs.cmd, r, parent, ""}
}

func (o *Environment) indirectRev(revisions map[string]revision, pkg *exports.Package, parent string) {
	for im, _ := range pkg.ImportedPackages {
		if _, has := revisions[im]; !has {
			revisions[im] = o.pkgRevision(path.Join(o.GOPATH, "src", im), pkg.Path)
			o.indirectRev(revisions, o.Pkg(im), pkg.Path)
		}
	}
}

func (env *Environment) checkoutRevCmd(dir string, c string, args ...string) error {
	//fmt.Printf("running:\n\t%s %s\n in\n\t%#v\n", c, strings.Join(args, " "), dir)
	cmd := exec.Command(c, args...)
	cmd.Dir = dir
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, env.GOPATH),
		fmt.Sprintf(`GOROOT=%s`, env.GOROOT),
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		//return fmt.Errorf("checkoutRevCmd Error %s", stderr.String())
		return fmt.Errorf("%s", stderr.String())
	}
	return nil
}

func (env *Environment) checkoutBzr(dir string, rev string) error {
	// update -r {tag}
	return env.checkoutRevCmd(dir, "bzr", "update", "-r", rev)
}

func (env *Environment) checkoutGit(dir string, rev string) error {
	return env.checkoutRevCmd(dir, "git", "checkout", rev)
}

func (env *Environment) checkoutHg(dir string, rev string) error {
	// update -r
	return env.checkoutRevCmd(dir, "hg", "update", "-r", rev)
}

func _repoRoot(dir string) string {
	_, root, err := vcsForDir(dir)
	// fmt.Printf("root by vcsForDir is: %#v\n", root)
	if err != nil {
		panic("can't find repodir for " + dir + " : " + err.Error())
	}
	return root
}

/*
func _fix(c *cli.Context) ErrorCode {
	return 0
}
*/
