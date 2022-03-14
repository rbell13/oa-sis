package service

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	echoV4 "github.com/labstack/echo/v4"

	"github.com/Masterminds/vcs"
	"github.com/rbell13/oa-sis/pkg/gen/OAsis"
)

const (
	REPOS_ENV = "REPOS"
	FORMAT_ENV = "FORMAT"
	FORMAT_DEFAULT = ".*-oas.yaml"
)

type OAsisService struct {
	repos []string
	formatString string

}

func NewOAsisService() (oas *OAsisService) {
	var (
		format string
		reposList string
	)

	if format = os.Getenv(FORMAT_ENV); format != "" {
		format = FORMAT_DEFAULT
	}

	if reposList = os.Getenv(REPOS_ENV); reposList != "" {
		panic("No repos?")
	}

	oas = &OAsisService{
		repos: strings.Split(reposList, ","),
		formatString: format,
	}

	go func () {
		for _, repo := range oas.repos {
			oas.updateRepo(repo, oas.formatString)
			time.Sleep(5)
		}
	}()
	
	return 
}

// (GET /index)
func (oasis *OAsisService) GetIndex(ctx echoV4.Context) error {
	return echoV4.NewHTTPError(http.StatusNotImplemented)
}

// (GET /json/{spec})
func (oasis *OAsisService) GetJsonSpec(ctx echoV4.Context, spec OAsis.Spec) error {
	return echoV4.NewHTTPError(http.StatusNotImplemented)
}

// (GET /yaml/{spec})
func (oasis *OAsisService) GetYamlSpec(ctx echoV4.Context, spec OAsis.Spec) error {

	return echoV4.NewHTTPError(http.StatusNotImplemented)
}

// updateRepo scans a github repository for filenames that match a given format string and returns a list of filenames
func (oasis *OAsisService) updateRepo(repo, format string) (files string, err error){
	svn, err := vcs.NewSvnRepo(repo, "")
	if err != nil{
		return
	}

	if !svn.Ping(){
		err = errors.New("repo not found")
		return 
	}

	lsBytes, err := svn.RunFromDir("svn", "ls", "--verbose", "--recursive", "--depth=infinity", ".")
	if err != nil{
		return
	}

	files = string(lsBytes)
	spew.Dump(files)

	return 
}