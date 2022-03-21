package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"io"
	"io/ioutil"
	"regexp"
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
	REFRESH_SECONDS = 60
	RepoNotFound    = "Repo not found"
)

type OAsisService struct {
	repos        []string
	formatString string
	oasFiles     []string
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
func (oasis *OAsisService) GetIndex(ctx echoV4.Context) error {
	return ctx.File("./temporary/index.html")
}

// (GET /json/{spec}).
func (oasis *OAsisService) GetJsonSpec(ctx echoV4.Context, spec OAsis.Spec) error {
	return ctx.File("./temporary/" + string(spec) + ".json")
}

// (GET /yaml/{spec}).
func (oasis *OAsisService) GetYamlSpec(ctx echoV4.Context, spec OAsis.Spec) error {
	return ctx.File("./temporary/" + string(spec) + ".yaml")
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
		spew.Dump(rawURL)

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

		err = os.Mkdir("./temporary", 0777)
		if err != nil {
			spew.Dump(err)

			continue
		}

		localfile, err := os.Create("./temporary/" + strings.ReplaceAll(file, "trunk/pkg/", ""))
		if err != nil {
			spew.Dump(err)
		}
		defer localfile.Close()

		size, err := io.Copy(localfile, resp.Body)
		if err != nil {
			spew.Dump(err)
		}

		spew.Dump(size)

		// create json version of spec
		byte, err := exec.Command("openapi-generator-cli", "generate", "-i", "./temporary/"+strings.ReplaceAll(file, "trunk/pkg/", ""), "-g", "openapi", "-o", "./temporary/").CombinedOutput()
		if err != nil {
			spew.Dump(byte)
			spew.Dump(err)
		}

		//generate dynamic-html documentation for spec
		err = exec.Command("openapi-generator-cli", "generate", "-i", "./temporary/"+strings.ReplaceAll(file, "trunk/pkg/", ""), "-g", "html2", "-o", "./temporary/").Run()
		if err != nil {
			spew.Dump(byte)
			spew.Dump(err)
		}

		//generate swagger-ui documentation for spec
		err = exec.Command("openapi-generator-cli", "generate", "-i", "./temporary/"+strings.ReplaceAll(file, "trunk/pkg/", ""), "-g", "swagger-ui", "-o", "./temporary/").Run()
		if err != nil {
			spew.Dump(byte)
			spew.Dump(err)
		}

		files, err := ioutil.ReadDir("./temporary")
		if err != nil {
			spew.Dump(err)
		}
	
		for _, file := range files {
			fmt.Println(file.Name(), file.IsDir())
		}
	}
}
