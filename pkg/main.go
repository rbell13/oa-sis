package main

import (
	"github.com/foolin/goview/supports/echoview-v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rbell13/oa-sis/pkg/gen/OAsis"
	OAsisService "github.com/rbell13/oa-sis/pkg/service"
)

func main() {
	e := echo.New()
	service := OAsisService.NewOAsisService()
	OAsis.RegisterHandlers(e, service)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Renderer = echoview.Default()

	e.Logger.Fatal(e.Start(":8080"))
}
