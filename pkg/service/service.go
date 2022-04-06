package service

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/vcs"
	"github.com/davecgh/go-spew/spew"
	echoV4 "github.com/labstack/echo/v4"
	"github.com/rbell13/oa-sis/pkg/gen/OAsis"
)

const (
	REPOS_ENV       = "REPOS"
	FORMAT_ENV      = "FORMAT"
	FORMAT_DEFAULT  = ".*-oas.yaml"
	REFRESH_SECONDS = 30
	DOCS            = "./docs"
	RepoNotFound    = "Repo not found"
)

type OAsisService struct {
	repos        []string
	formatString string
	oasFiles     []string
}

type indexModel struct {
	Repos     []string
	Specs     []string
	GoVersion string
	GitCommit string
	BuildDate string
}

func NewOAsisService() (oas *OAsisService) {
	var (
		format    string
		reposList string
	)

	if format = os.Getenv(FORMAT_ENV); format == "" {
		format = FORMAT_DEFAULT
	}

	if reposList = os.Getenv(REPOS_ENV); reposList == "" {
		panic("No repos?")
	}

	oas = &OAsisService{
		repos:        strings.Split(reposList, ","),
		formatString: format,
	}

	go func() {
		for _, repo := range oas.repos {
			var err error

			oas.oasFiles, err = oas.updateRepo(repo, oas.formatString)
			if err != nil {
				panic(err)
			}

			getOASFromRemote(context.Background(), repo, oas.oasFiles)
		}

		time.Sleep(time.Second * REFRESH_SECONDS)
	}()

	return
}

// (GET /index).
// Serve template page with list of available specs.
func (oasis *OAsisService) GetIndex(ctx echoV4.Context) error {
	var dirs []string
	files, err := ioutil.ReadDir(DOCS)
	if err != nil {
		return echoV4.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}

	return ctx.Render(http.StatusOK, "index.html", indexModel{
		Repos:     oasis.repos,
		Specs:     dirs,
		GoVersion: runtime.Version(),
		// GitCommit: gitCommit(),
		BuildDate: time.Now().Format(time.RFC3339),
	})
}

// (GET /json/{spec}).
func (oasis *OAsisService) GetJsonSpec(ctx echoV4.Context, spec OAsis.Spec) error {
	spew.Dump("./" + filepath.Join(DOCS, string(spec), "openapi.json"))

	return ctx.File("./" + filepath.Join(DOCS, string(spec), "openapi.json"))
}

// (GET /yaml/{spec}).
func (oasis *OAsisService) GetYamlSpec(ctx echoV4.Context, spec OAsis.Spec) error {
	spew.Dump("./" + filepath.Join(DOCS, string(spec), "openapi.yaml"))

	return ctx.File("./" + filepath.Join(DOCS, string(spec), "openapi.yaml"))
}

// updateRepo scans a github repository for filenames that match a given format string and returns a list of filenames.
func (oasis *OAsisService) updateRepo(repo, format string) (files []string, err error) {
	svn, err := vcs.NewSvnRepo(repo, "")
	if err != nil {
		return
	}

	if !svn.Ping() {
		return nil, errors.New(RepoNotFound)
	}

	lsBytes, err := svn.RunFromDir("svn", "ls", "-R", "--recursive", "--depth=infinity", repo)
	if err != nil {
		return
	}

	files = regexp.MustCompile(format).FindStringSubmatch(string(lsBytes))

	return
}

func getOASFromRemote(ctx context.Context, repo string, files []string) {
	repourl, err := url.Parse(repo)
	if err != nil {
		spew.Dump(err)
	}

	for _, file := range files {
		rawURL := fmt.Sprintf("https://raw.githubusercontent.com" + repourl.Path + "/main/" + strings.Trim(file, "trunk/"))
		filename := filepath.Base(file)
		filenameNoExt := strings.TrimSuffix(filename, filepath.Ext(filename))

		req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
		if err != nil {
			spew.Dump(err)

			continue
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			spew.Dump(err)

			continue
		}
		defer resp.Body.Close()

		specBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			spew.Dump(err)

			continue
		}

		dir := filepath.Join(DOCS, filenameNoExt)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		err = ioutil.WriteFile(filepath.Join(dir, filename), specBytes, os.ModePerm)
		if err != nil {
			spew.Dump(err)
			panic(err)
		}
		// Create json version of spec
		err = exec.Command("openapi-generator-cli", "generate", "-i", filepath.Join(dir, filename), "-g", "openapi", "-o", dir).Run()
		if err != nil {
			spew.Dump(err)
			panic(err)
		}
		// Create yaml version of spec
		err = exec.Command("openapi-generator-cli", "generate", "-i", filepath.Join(dir, filename), "-g", "openapi-yaml", "-o", dir).Run()
		if err != nil {
			spew.Dump(err)
			panic(err)
		}

		// generate dynamic-html documentation for spec
		err = exec.Command("openapi-generator-cli", "generate", "-i", filepath.Join(dir, filename), "-g", "dynamic-html", ">", filepath.Join(dir, filenameNoExt+".html")).Run()
		if err != nil {
			spew.Dump(string(specBytes))
			spew.Dump(err)
		}

		// generate swagger-ui documentation for spec
		err = exec.Command("openapi-generator-cli", "generate", "-i", filepath.Join(dir, filename), "-g", "swagger-ui", ">", filepath.Join(dir, filenameNoExt+".swagger.html")).Run()
		if err != nil {
			spew.Dump(string(specBytes))
			spew.Dump(err)
		}
	}
}
