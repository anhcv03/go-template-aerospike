package common

import (
	"net/http"

	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi"
	"github.com/labstack/echo/v4"
)

func GetHealthyRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/healthy", getHealthyHandler(s))
}

func getHealthyHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !s.Ready() {
			// We use 521 to indicate an error state
			// same as Cloudflare: https://support.cloudflare.com/hc/en-us/articles/115003011431#521error
			return c.String(521, "Not ready.")
		}

		return c.String(http.StatusOK, "Ready")
	}
}
