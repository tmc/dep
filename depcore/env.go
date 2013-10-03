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
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

type Environment struct {
	*gdf.Environment
	TMPDIR     string
	db         *db
	tentative  *tentativeEnvironment
	IgnorePkgs map[string]bool
}

func NewEnv(gopath string) (ø *Environment) {
	if gopath == "" {
		panic("can't create environment for empty GOPATH")
	}
	ø = &Environment{}
	ø.Environment = gdf.NewEnv(runtime.GOROOT(), gopath)
	ø.TMPDIR = os.Getenv("DEP_TMP")
	ø.mkdb()
	ø.IgnorePkgs = map[string]bool{}
	ign, e := ioutil.ReadFile(filepath.Join(gopath, ".depignore"))
	if e != nil {
		return
	}
	lines := bytes.Split(ign, []byte("\n"))

	for _, line := range lines {
		ø.IgnorePkgs[string(line)] = true
	}

	return
}

var exampleRegExp = regexp.MustCompile("example(s?)$")

func (env *Environment) shouldIgnorePkg(pkg string) bool {
	if exampleRegExp.MatchString(pkg) {
		return true
	}
	return env.IgnorePkgs[pkg]
}

func (env *Environment) RevFile(pkg string) string {
	return path.Join(env.GOPATH, "src", pkg, revFileName)
}

// for each import, get the revisions
func (o *Environment) recursiveImportRevisions(revisions map[string]revision, pkg *gdf.Package, parent string) {
	for im, _ := range pkg.ImportedPackages {
		if _, has := revisions[im]; !has {
			p, err := o.Pkg(im)
			if err != nil {
				panic(fmt.Sprintf("package %s does not exist", im))
			}
			//d, _ := p.Dir()
			var d string
			var internal bool
			d, internal, err = o.PkgDir(im)
			if internal {
				continue
			}
			revisions[im] = o.getRevision(d, pkg.Path)
			o.recursiveImportRevisions(revisions, p, pkg.Path)
		}
	}
}

func (env *Environment) HasRevFile(pkg string) bool {
	_, err := os.Stat(env.RevFile(pkg))
	if err != nil {
		return false
	}
	return true
}

func (env *Environment) newTentative() (t *tentativeEnvironment) {
	if env.tentative != nil {
		panic("can't create more than one tentative environment for the same env")
	}
	env.tentative = &tentativeEnvironment{
		Original:    env,
		Environment: NewEnv(env.mkTempDir()),
	}
	env.tentative.Open()
	return env.tentative
}

func (o *Environment) NumPkgsInRegistry() int {
	return o.db.NumPackages()
}

func (o *Environment) pkgJson(path string) (b []byte, internal bool) {
	p, err := o.Pkg(path)
	if err != nil {
		panic(err.Error())
	}
	internal = p.Internal
	b, err = json.MarshalIndent(p, "", "   ")
	if err != nil {
		panic(err.Error())
	}
	return
}

func (o *Environment) scan(dir string) (b []byte, internal bool) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err.Error())
	}
	b, internal = o.pkgJson(o.PkgPath(dir))
	b = append(b, []byte("\n")...)
	return
}

// TODO: get rid of it
func (env *Environment) packageToDBFormat(pkgMap map[string]*dbPkg, pkg *gdf.Package, includeImported bool) (dbExps []*exp, dbImps []*imp) {
	p := &dbPkg{}
	p.Package = pkg.Path
	p.Json = niceJson(pkg)
	p.InitMd5 = pkg.InitMd5
	p.JsonMd5 = pkg.JsonMd5()
	pkgMap[pkg.Path] = p
	dbExps = []*exp{}
	dbImps = []*imp{}

	if includeImported {
		for im, _ := range pkg.ImportedPackages {
			if _, has := pkgMap[im]; has {
				continue
			}
			imPkg, err := env.Pkg(im)
			if err != nil {
				panic(fmt.Sprintf("%s imports not existing package %s", pkg.Path, im))
			}
			pExp, pImp := env.packageToDBFormat(pkgMap, imPkg, includeImported)
			dbExps = append(dbExps, pExp...)
			dbImps = append(dbImps, pImp...)
		}
	}

	pkgjs := pkg

	for k, v := range pkgjs.Exports {
		dbE := &exp{}
		dbE.Package = pkg.Path
		dbE.Name = k
		dbE.Value = v
		dbExps = append(dbExps, dbE)
	}

	for k, v := range pkgjs.Imports {
		dbI := &imp{}
		dbI.Package = pkg.Path
		arr := strings.Split(k, "#")
		dbI.Name = arr[1]
		dbI.Value = v
		dbI.Import = arr[0]
		dbImps = append(dbImps, dbI)
	}
	return
}

// open an environment
func (o *Environment) Open() {
	subDirs := [3]string{"src", "bin", "pkg"}
	for _, s := range subDirs {
		d := path.Join(o.GOPATH, s)
		stat, err := os.Stat(d)
		if err != nil {
			errMk := os.Mkdir(d, 0755)
			if errMk != nil {
				panic(errMk.Error())
			}
			continue
		}
		if !stat.IsDir() {
			panic(d + " is a file. but should be a directory")
		}
	}
	o.mkdb()
}

// close an environment
func (o *Environment) Close() {
	if o.db != nil {
		o.db.Close()
	}
	if o.tentative != nil {
		o.tentative.Close()
	}
}

func (o *Environment) mkTempDir() (dir string) {
	stat, err := os.Stat(o.TMPDIR)
	if err != nil {
		panic(err.Error())
	}
	if !stat.IsDir() {
		panic(o.TMPDIR + " is a file. but should be a directory")
	}

	dir, err = ioutil.TempDir(o.TMPDIR, "gopath_")
	if err != nil {
		panic(err.Error())
	}
	return
}

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
		fmt.Printf("error while running: %s %s in %s\n", c, strings.Join(args, " "), dir)
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

func (o *Environment) getRevisionBzr(dir string) string {
	res := o.getRevCmd(dir, "bzr", "log", "-l", "1", "--show-ids")
	sm := bzrRevRe.FindAllStringSubmatch(res, 1)
	return sm[0][1]
}

func (o *Environment) getRevision(dir string, parent string) (rev revision) {
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
	reporoot := _repoRoot(dir)
	return revision{vcs.cmd, r, parent, "", o.PkgPath(reporoot)}
}

// setup the env for a cmd
func (env *Environment) cmdEnv() []string {
	return []string{
		fmt.Sprintf(`GOPATH=%s`, env.GOPATH),
		fmt.Sprintf(`GOROOT=%s`, env.GOROOT),
		fmt.Sprintf(`PATH=%s`, os.Getenv("PATH")),
	}
}

// runs go test in the given dir
func (o *Environment) goTest(dir string) ([]byte, error) {
	cmd := exec.Command("go", "test")
	cmd.Env = o.cmdEnv()
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

// returns all packages in env.GOPATH/src
func (env *Environment) allPackages() (a []*gdf.Package) {
	a = []*gdf.Package{}
	prs := newSubPackages(env)
	err := filepath.Walk(path.Join(env.GOPATH, "src"), prs.Walker)
	if err != nil {
		panic(err.Error())
	}
	for p, _ := range prs.packages {
		pk, e := env.Pkg(p)
		if e != nil {
			continue
		}
		a = append(a, pk)
	}
	return
}

func (env *Environment) cleandb() {
	env.Close()
	os.Remove(path.Join(env.GOPATH, "dep.db"))
	env.Open()
}

// creates the db file if it is not there
func (env *Environment) mkdb() {
	dbFile := path.Join(env.GOPATH, "dep.db")
	_, dbFileerr := os.Stat(dbFile)
	dB, err := db_open(env, dbFile)
	if err != nil {
		panic(err.Error())
	}
	if dbFileerr != nil {
		dB.CreateTables()
	}
	env.db = dB
}

func (env *Environment) Init() (conflicts map[string]map[string][3]string) {
	env.cleandb()
	env.mkdb()
	ps := env.allPackages()
	env.db.registerPackages(true, ps...)
	return env.checkIntegrity(ps...)
}

func (env *Environment) CheckIntegrity() (conflicts map[string]map[string][3]string) {
	env.mkdb()
	ps := env.allPackages()
	return env.checkIntegrity(ps...)
}

// checks the integrity of all packages
// by adding them to the db and checking for conflicts
func (env *Environment) checkIntegrity(ps ...*gdf.Package) (conflicts map[string]map[string][3]string) {
	//fmt.Println("check integrity")
	//env.mkdb()
	conflicts = map[string]map[string][3]string{}
	///*
	conflicts["#dep-registry-orphan#"] = map[string][3]string{}
	conflicts["#dep-registry-inconsistency#"] = map[string][3]string{}

	defer func() {
		if len(conflicts["#dep-registry-orphan#"]) == 0 {
			delete(conflicts, "#dep-registry-orphan#")
		}
		if len(conflicts["#dep-registry-inconsistency#"]) == 0 {
			delete(conflicts, "#dep-registry-inconsistency#")
		}
	}()
	//*/
	pkgs := map[string]bool{}

	//ps := env.allPackages()
	//env.db.registerPackages(ps...)
	for _, p := range ps {
		d, er := env.Diff(p, false)
		if er != nil {
			conflicts[p.Path] = map[string][3]string{
				"#dep-registry-inconsistency#": [3]string{"missing", er.Error(), ""},
			}
			return
		}

		if d != nil && len(d.Exports) > 0 {
			conflicts[p.Path] = map[string][3]string{
				"#dep-registry-inconsistency#": [3]string{"exports", strings.Join(d.Exports, "\n"), ""},
			}
			return
		}

		if d != nil && len(d.Imports) > 0 {
			conflicts[p.Path] = map[string][3]string{
				"#dep-registry-inconsistency#": [3]string{"imports", strings.Join(d.Imports, "\n"), ""},
			}
			return
		}

		pkgs[p.Path] = true
		errs := env.db.hasConflict(p, map[string]bool{})
		if len(errs) > 0 {
			conflicts[p.Path] = errs
			return
		}
	}

	if len(conflicts) > 0 {
		return
	}

	dbpkgs, err := env.db.GetAllPackages()
	if err != nil {
		panic(err.Error())
	}

	for _, dbp := range dbpkgs {
		if !pkgs[dbp.Package] {
			conflicts["#dep-registry-orphan#"][dbp.Package] = [3]string{"orphan", dbp.Package, ""}
			continue
		}
	}

	return
}

func (o *Environment) getJson(pkg string) string {
	p, err := o.Pkg(pkg)
	if err != nil {
		panic(err.Error())
	}
	var b []byte
	b, err = json.Marshal(p)
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}

func (o *Environment) loadJson(pkgPath string) (ø *gdf.Package) {
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
