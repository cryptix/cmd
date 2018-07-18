// depCheckout clones overriden packages from a dep project (Gopkg.toml) into os.Args[1]
// it `go get`s the originals, adds the `source` as a git remote named fork.
// then it fetches them and resets to the specified `revision``
package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cryptix/go/logging"
	"github.com/golang/dep"
	"github.com/golang/dep/gps"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/vcs"
)

var check = logging.CheckFatal

func main() {
	wd, err := os.Getwd()
	check(errors.Wrap(err, "GetWD failed"))
	log.Print("wd: ", wd)

	gp := os.Getenv("GOPATH")
	log.Print("GOPATH:", gp)

	if len(os.Args) < 2 {
		check(errors.Errorf("usage: %s <forkPath>", os.Args[0]))
	}
	forkPath := os.Args[1]

	ctx := &dep.Ctx{
		WorkingDir:   wd,
		GOPATH:       gp,
		GOPATHs:      strings.Split(gp, ":"),
		ExplicitRoot: wd,
		Out:          log.New(os.Stderr, "", 0),
		Err:          log.New(os.Stderr, "", 0),
		Verbose:      true,
	}

	proj, err := ctx.LoadProject()
	check(errors.Wrap(err, "load proj"))

	for pr, prop := range proj.Manifest.Ovr {
		rev, ok := prop.Constraint.(gps.Revision)
		if prop.Source != "" && ok {
			log.Print("Package: ", pr)
			log.Print("Fork Source: ", prop.Source)
			pkg := string(pr)
			loc, err := getPkgLocation(forkPath, pkg)
			if err != nil {
				log.Print("could not locate forked:", err)
				check(goGet(forkPath, pkg))
				if pkg == "github.com/qor/auth_themes" {
					pkg = "github.com/qor/auth_themes/clean"
				}
				loc, err = getPkgLocation(forkPath, pkg)
				check(errors.Wrap(err, "2nd pkg locate failed"))
			}
			log.Print("OS location:", loc)
			remoteURL, err := gitRemoteGetURL(loc, "fork")
			if err != nil {
				log.Println("no remote fork:", err)
				repoRoot, err := vcs.RepoRootForImportDynamic(prop.Source, true)
				check(errors.Wrap(err, "vcs.RepoFromImport failed"))
				err = gitRemoteAdd(loc, "fork", repoRoot.Repo)
				check(errors.Wrap(err, "failed to add fork"))
			} else {
				log.Println("Remote:", remoteURL)
			}
			err = gitFetch(loc, "fork")
			check(err)

			err = gitResetHard(loc, rev)
			check(err)
			log.Print("Done")
			log.Print()
		}
	}
}

func getPkgLocation(gopath, pkg string) (string, error) {
	cmd := exec.Command("go", "list", "-f", `{{.Dir}}`, pkg)
	cmd.Env = []string{"GOPATH=" + gopath}

	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), errors.Wrapf(err, "pkg(%s) not located", pkg)
}

func goGet(gopath, pkg string) error {
	cmd := exec.Command("go", "get", "-d", pkg+"/...")
	cmd.Env = []string{"GOPATH=" + gopath}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run(), "goGet failed")
}

func gitRemoteGetURL(path, remote string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	return string(out), errors.Wrap(err, "git remote get-url failed")
}

func gitRemoteAdd(path, remote, url string) error {
	cmd := exec.Command("git", "remote", "add", remote, url)
	cmd.Dir = path
	return errors.Wrap(cmd.Run(), "git remote add failed")
}

func gitFetch(path, remote string) error {
	cmd := exec.Command("git", "fetch", "-v", remote)
	cmd.Dir = path
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run(), "gitFetch failed")
}

func gitResetHard(path string, rev gps.Revision) error {
	cmd := exec.Command("git", "reset", "--hard", string(rev))
	cmd.Dir = path
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run(), "gitFetch failed")
}
