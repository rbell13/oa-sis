package service

import (
	"net/http"

	echoV4 "github.com/labstack/echo/v4"

	"github.com/rbell13/oa-sis/pkg/gen/OAsis"
)

type OAsisService struct {
}

func NewOAsisService() *OAsisService {
	return &OAsisService{}
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
