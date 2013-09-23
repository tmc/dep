package depcore

/*
import (
	"bytes"
	_ "code.google.com/p/go.exp/inotify"
	"encoding/json"
	"fmt"
	// "github.com/metakeule/cli"
	// "github.com/metakeule/exports"
	"io/ioutil"
	_ "launchpad.net/goamz/aws"
	"os"
	"os/exec"
	// "path"
)
*/
// looks for revisions in the given file and checks out the
// packages that are not already installed
// TODO: we need to check for the repo of a package and control,
// if the repo is not already checked out
// same for update
// TODO ignore packages everywhere that have /example/ or /examples/ in their path same for /test/ and /tests/
/*
func CLICheckout(c *cli.Context, o *Options) ErrorCode {
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

	visited := map[string]bool{}

	for pkg, rev := range revisions {
		exists := o.Env.PkgExists(pkg)

		if force || !exists {
			dir := path.Join(o.GOPATH, "src", pkg)
			r := _repoRoot(dir)
			if doneRepos[r] {
				continue
			}
			if !exists {
				// install package, but only if repo does not exist
				// question: what happens if a pkg was added to repo-dir, i.e. repo dir exists and pkg dir not?
				visited[pkg] = true
				args := []string{"get", pkg}
				//args = append(args, c.Args()...)

				cmd := exec.Command("go", args...)
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
			}

			doneRepos[r] = true

			// checkout revision
			var checkoutErr error
			switch rev.VCM {
			case "bzr":
				checkoutErr = checkoutBzr(o, r, rev.Rev)
			case "git":
				checkoutErr = checkoutGit(o, r, rev.Rev)
			case "hg":
				checkoutErr = checkoutHg(o, r, rev.Rev)
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
	return 0
}
*/

/*
func loadRevisions(file string) (revs map[string]revision, err error) {
	revs = map[string]revision{}
	var data []byte
	data, err = ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &revs)
	return
}

func (env *Environment) CLICheckout(file string, force bool) error {
	revisions, err := loadRevisions(file)
	if err != nil {
		panic(err.Error())
	}
	//fmt.Printf("%#v\n", revisions)
	doneRepos := map[string]bool{}
	visited := map[string]bool{}

	for pkg, rev := range revisions {
		exists := env.PkgExists(pkg)

		if force || !exists {
			dir := env.PkgDir(pkg)
			r := _repoRoot(dir)
			if doneRepos[r] {
				continue
			}

			// TODO:  do whatever update does
			if !exists {
				// install package, but only if repo does not exist
				// question: what happens if a pkg was added to repo-dir, i.e. repo dir exists and pkg dir not?
				visited[pkg] = true
				args := []string{"get", "-d", pkg}
				//args = append(args, c.Args()...)

				cmd := exec.Command("go", args...)
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
					panic(stdout.String() + "\n" + stderr.String())
				}
			}

			doneRepos[r] = true

			// checkout revision
			var checkoutErr error
			switch rev.VCM {
			case "bzr":
				checkoutErr = env.checkoutBzr(r, rev.Rev)
			case "git":
				checkoutErr = env.checkoutGit(r, rev.Rev)
			case "hg":
				checkoutErr = env.checkoutHg(r, rev.Rev)
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
	return nil
}
*/