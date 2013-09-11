package main

import (
	"bytes"
	_ "code.google.com/p/go.exp/inotify"
	"encoding/json"
	"fmt"
	"github.com/metakeule/cli"
	"github.com/metakeule/dep/db"
	"github.com/metakeule/exports"
	"io/ioutil"
	_ "launchpad.net/goamz/aws"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"
)

var _ = fmt.Printf

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

func _lint(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	return 0
}

func _show(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	fmt.Printf("%s\n", asJson(packages()...))
	return 0
}

func _registerPackage(pkgMap map[string]*db.Pkg, pkg *exports.Package) (dbExps []*db.Exp, dbImps []*db.Imp) {
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
		imPkg := exports.DefaultEnv.Pkg(im)
		pExp, pImp := _registerPackage(pkgMap, imPkg)
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

func registerPackages(pkgs ...*exports.Package) {
	dbExps := []*db.Exp{}
	dbImps := []*db.Imp{}
	pkgMap := map[string]*db.Pkg{}

	for _, pkg := range pkgs {
		pExp, pImp := _registerPackage(pkgMap, pkg)
		dbExps = append(dbExps, pExp...)
		dbImps = append(dbImps, pImp...)
	}

	dbPkgs := []*db.Pkg{}

	for _, dbPgk := range pkgMap {
		dbPkgs = append(dbPkgs, dbPgk)
	}

	err := db.InsertPackages(dbPkgs, dbExps, dbImps)
	if err != nil {
		panic(err.Error())
	}

	/*
		for _, pk := range dbPkgs {
			fmt.Println("registered: ", pk.Package)
		}
	*/
}

func _register(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	_, dbFileerr := os.Stat(DEP)

	err := db.Open(DEP)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables()
	}

	pkgs := packages()
	registerPackages(pkgs...)
	return 0
}

func _get(c *cli.Context) ErrorCode {
	return 0
}

func _install(c *cli.Context) ErrorCode {
	return 0
}

type pkgDiff struct {
	Path    string
	Exports []string
	Imports []string
}

func toJson(i interface{}) []byte {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	return b
}

func mapDiff(_old map[string]string, _new map[string]string) (diff []string) {
	diff = []string{}
	var visited = map[string]bool{}

	for k, v := range _old {
		visited[k] = true
		vNew, inNew := _new[k]
		if !inNew {
			diff = append(diff, "---"+k+": "+v)
			continue
		}
		if v != vNew {
			diff = append(diff, "---"+k+": "+v)
			diff = append(diff, "+++"+k+": "+vNew)
		}
	}

	for k, v := range _new {
		if !visited[k] {
			diff = append(diff, "+++"+k+": "+v)
		}
	}
	return
}

func _diff(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	_, dbFileerr := os.Stat(DEP)
	err := db.Open(DEP)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables()
	}

	pkgs := packages()

	res := []*pkgDiff{}

	for _, pk := range pkgs {

		dbpkg, exps, imps, e := db.GetPackage(pk.Path, true, true)
		if e != nil {
			panic("package not registered: " + pk.Path)
		}

		js := asJson(pk)

		// TODO: check the hash instead, escp. check the exports and imports hash
		if string(js) != string(dbpkg.Json) {
			//__diff(a, b)
			pkgjs := pk.PackageJSON()

			var oldExports = map[string]string{}

			for _, dbExp := range exps {
				oldExports[dbExp.Name] = dbExp.Value
			}

			pDiff := &pkgDiff{}
			pDiff.Path = pk.Path
			pDiff.Exports = mapDiff(oldExports, pkgjs.Exports)

			var oldImports = map[string]string{}

			for _, dbImp := range imps {
				oldImports[dbImp.Import+"#"+dbImp.Name] = dbImp.Value
			}
			pDiff.Imports = mapDiff(oldImports, pkgjs.Imports)

			if len(pDiff.Exports) > 0 || len(pDiff.Imports) > 0 {
				res = append(res, pDiff)
			}
		}
	}
	if len(res) > 0 {
		fmt.Printf("%s\n", toJson(res))
	}
	return 0
}

func mkdirTempDir() (tmpGoPath string) {
	depPath := path.Join(HOME, ".dep")
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

	- make a tempdir in .deb with subdirs pkg, src and bin
	- set GOPATH to the tempdir and go get the needed package
	- check for package and all their dependancies, if
		- their dependant packages would be fine with the new exports
	- if all is fine, move the missing src entries to the real GOPATH and install the package
	- if there are some conflicts, show them
	- remove -r the tempdir
*/

// TODO: add verbose flag for verbose output
func hasConflict(p *exports.Package) (errors map[string][3]string) {
	pkg := p.PackageJSON()
	imp, err := db.GetImported(pkg.Path)
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

func checkConflicts(tempEnv *exports.Environment, pkg *exports.Package) (errs map[string][3]string) {
	tempPkg := tempEnv.Pkg(pkg.Path)
	if !exports.DefaultEnv.PkgExists(pkg.Path) {
		panic(fmt.Sprintf("package %s is not installed"))
	}

	registerPackages(pkg)
	errs = hasConflict(tempPkg)
	return
}

func runGoTest(tmpDir string, dir string) ([]byte, error) {
	//args := []string{"get", pkg.Path}
	//cmd := exec.Command("go", args...)
	cmd := exec.Command("go", "test")
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, tmpDir),
		fmt.Sprintf(`GOROOT=%s`, GOROOT),
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// TODO: checkout certain revisions if there is a dep.rev file
func _update(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	tmpDir := mkdirTempDir()

	defer func() {
		if !c.Bool("keep-temp-gopath") {
			err := os.RemoveAll(tmpDir)
			if err != nil {
				panic(err.Error())
			}
		}
	}()

	pkgs := packages()
	tempEnv := exports.NewEnv(runtime.GOROOT(), tmpDir)

	_, dbFileerr := os.Stat(DEP)
	err := db.Open(DEP)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	if dbFileerr != nil {
		fmt.Println(dbFileerr)
		db.CreateTables()
	}

	conflicts := map[string]map[string][3]string{}

	// TODO: check if the package and its dependancies are installed in the
	// default path, if so, check, they are registered / updated in the database.
	// if not, register /update them
	// TODO make a db connection to get the conflicting
	// packages.
	// it might be necessary to make an update of the db infos first
	visited := map[string]bool{}

	for _, pkg := range pkgs {
		if !visited[pkg.Path] {
			visited[pkg.Path] = true
			args := []string{"get", "-u", pkg.Path}
			//args = append(args, c.Args()...)

			cmd := exec.Command("go", args...)
			cmd.Env = []string{
				fmt.Sprintf(`GOPATH=%s`, tmpDir),
				fmt.Sprintf(`GOROOT=%s`, GOROOT),
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
			// make flag with --skip-gotest
			tempPkgPath := path.Join(tempEnv.GOPATH, "src", pkg.Path)
			// go test should fail, if the newly installed packages are not compatible
			res, e := runGoTest(tmpDir, tempPkgPath)
			if e != nil {
				panic("Error while running 'go test " + tempPkgPath + "'\n" + string(res) + "\n" + e.Error())
			}

			errs := checkConflicts(tempEnv, pkg)
			if len(errs) > 0 {
				conflicts[pkg.Path] = errs
			}
		}

		for im, _ := range pkg.Imports {
			if !visited[pkg.Path] {
				imPkg := exports.DefaultEnv.Pkg(im)
				tempPkgPath := path.Join(tempEnv.GOPATH, "src", im)
				// go test should fail, if the newly installed packages are not compatible
				res, e := runGoTest(tmpDir, tempPkgPath)
				if e != nil {
					panic("Error while running 'go test " + tempPkgPath + "'\n" + string(res) + "\n" + e.Error())
				}

				errs := checkConflicts(tempEnv, imPkg)
				if len(errs) > 0 {
					conflicts[pkg.Path] = errs
				}
			}
		}
	}

	if len(conflicts) > 0 {
		b, e := json.MarshalIndent(conflicts, "", "  ")
		if e != nil {
			panic(e.Error())
		}
		fmt.Printf("%s\n", b)
		return UpdateConflict
	}

	// if we got here, everything is fine and we may do our update

	// TODO: it might be better to simply move the installed packages
	// instead of go getting them again, because they might
	// have changed in the meantime
	for _, pkg := range pkgs {
		// update all dependant packages as well
		args := []string{"get", "-u", pkg.Path}
		//args = append(args, c.Args()...)

		cmd := exec.Command("go", args...)
		cmd.Env = []string{
			fmt.Sprintf(`GOPATH=%s`, GOPATH),
			fmt.Sprintf(`GOROOT=%s`, GOROOT),
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
	}
	return 0
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

func getRevCmd(dir string, c string, args ...string) string {
	cmd := exec.Command(c, args...)
	cmd.Dir = dir
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, GOPATH),
		fmt.Sprintf(`GOROOT=%s`, GOROOT),
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

func getRevisionGit(dir string) string {
	return getRevCmd(dir, "git", "rev-parse", "HEAD")
}

func getRevisionHg(dir string) string {
	return getRevCmd(dir, "hg", "tip", "--template", "{node}")
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

func getRevisionBzr(dir string) string {
	res := getRevCmd(dir, "bzr", "log", "-l", "1", "--show-ids")
	sm := bzrRevRe.FindAllStringSubmatch(res, 1)
	return sm[0][1]
}

func pkgRevision(pkgPath string, parent string) (rev revision) {
	dir := path.Join(exports.DefaultEnv.GOPATH, "src", pkgPath)
	vcs, root, err := vcsForDir(dir)
	if err != nil {
		panic(err.Error())
	}
	_ = root
	var r string
	switch vcs.cmd {
	case "git":
		r = getRevisionGit(dir)
	case "hg":
		r = getRevisionHg(dir)
	case "bzr":
		r = getRevisionBzr(dir)
	case "svn":
		panic("svn is currently not supported")
	default:
		panic("unknown vcs command " + vcs.cmd)

	}
	return revision{vcs.cmd, r, parent, ""}
}

func indirectRev(revisions map[string]revision, pkg *exports.Package, parent string) {
	for im, _ := range pkg.Imports {
		if _, has := revisions[im]; !has {
			revisions[im] = pkgRevision(im, pkg.Path)
			indirectRev(revisions, exports.DefaultEnv.Pkg(im), pkg.Path)
		}
	}
}

func _revisions(c *cli.Context) ErrorCode {
	parseGlobalFlags(c)
	file := c.String("file")
	stdout := c.Bool("stdout")
	inclIndirect := c.Bool("include-indirect")
	pkgs := packages()
	allrevisions := map[string]revision{}

	for _, pkg := range pkgs {
		revisions := map[string]revision{}
		for im, _ := range pkg.Imports {
			revisions[im] = pkgRevision(im, pkg.Path)
			if inclIndirect {
				indirectRev(revisions, exports.DefaultEnv.Pkg(im), pkg.Path)
				continue
			}
		}

		if stdout {
			for k, v := range revisions {
				if _, exists := allrevisions[k]; !exists {
					allrevisions[k] = v
				}
			}
			continue
		}
		data, err := json.MarshalIndent(revisions, "", "  ")
		if err != nil {
			panic(err.Error())
		}

		dir, _ := pkg.Dir()
		filename := path.Join(dir, file)
		err = ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			panic(err.Error())
		}
	}
	if stdout {
		data, err := json.MarshalIndent(allrevisions, "", "  ")
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%s\n", data)
	}
	return 0
}

func checkoutRevCmd(dir string, c string, args ...string) string {
	cmd := exec.Command(c, args...)
	cmd.Dir = dir
	cmd.Env = []string{
		fmt.Sprintf(`GOPATH=%s`, GOPATH),
		fmt.Sprintf(`GOROOT=%s`, GOROOT),
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

func checkoutBzr(dir string, rev string) error {
	// update -r {tag}
	return checkoutRevCmd(dir, "bzr", "update", "-r", rev)
}

func checkoutGit(dir string, rev string) error {
	return checkoutRevCmd(dir, "git", "checkout", rev)
}

func checkoutHg(dir string, rev string) error {
	// update -r
	return checkoutRevCmd(dir, "hg", "update", "-r", rev)
}

func repoRoot(dir string) string {
	_, root, err := vcsForDir(dir)
	if err != nil {
		panic(err.Error())
	}
	return root
}

// looks for revisions in the given file and checks out the
// packages that are not already installed
// TODO: we need to check for the repo of a package and control,
// if the repo is not already checked out
// same for update
// TODO ignore packages everywhere that have /example/ or /examples/ in their path same for /test/ and /tests/
func _checkout(c *cli.Context) ErrorCode {
	file := c.String("file")
	force := c.Bool("force")

	data, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err.Error())
	}

	revisions := map[string]revision{}
	err = json.Unmarshal(data, &revisions)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Printf("%#v\n", revisions)
	doneRepos := map[string]bool{}

	for pkg, rev := range revisions {
		exists := exports.DefaultEnv.PkgExists(pkg)

		if force || !exists {
			dir := path.Join(exports.DefaultEnv.GOPATH, "src", pkg)
			if !exists {
				// install package, but only if repo does not exist

			}

			r := repoRoot(dir)
			if doneRepos[r] {
				continue
			}
			doneRepos[r] = true

			// checkout revision
			var checkoutErr error
			switch rev.VCM {
			case "bzr":
				checkoutErr = checkoutBzr(r, rev.Rev)
			case "git":
				checkoutErr = checkoutGit(r, rev.Rev)
			case "hg":
				checkoutErr = checkoutHg(r, rev.Rev)
			case "svn":
				panic("unsupported VCM svn for package " + pkg)
			default:
				panic("unsupported VCM " + rev.VCM + " for package " + pkg)
			}

			if checkoutErr != nil {
				panic("can't checkout " + pkg + " rev " + rev.Rev + ":\n" + checkoutErr.Error())
			}
		}
	}

	// vcsByCmd(cmd)
	/*
		vcs, root, err := vcsForDir(p)
		if err != nil {
			panic(err.Error())
		}
		if err := vcs.tagSync(root, "die version"); err != nil {
			panic(err.Error())
		}
	*/
	return 0
}

/*
func _fix(c *cli.Context) ErrorCode {
	return 0
}
*/
