package common

import (
	"net/http"

	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("get-version")

func GetVersionRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/version", getVersionHandler(s))
}

func getVersionHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, span := tracer.Start(c.Request().Context(), "getVersion")
		defer span.End()
		return c.String(http.StatusOK, config.GetFormattedBuildArgs())
	}
}
