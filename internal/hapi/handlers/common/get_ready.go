package common

import (
	"fmt"
	"net/http"

	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi"
	"github.com/labstack/echo/v4"
)

func GetReadyRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/ready", getReadyHandler(s))
}

// Readiness check
// This endpoint returns 200 when our Service is ready to serve traffic (i.e. respond to queries).
// Does read-only probes apart from the general server ready state.
// Note that /-/ready is typically public (and not shielded by a mgmt-secret), we thus prevent information leakage here and only return `"Ready."`.
// Structured upon https://prometheus.io/docs/prometheus/latest/management_api/
func getReadyHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !s.Ready() {
			// We use 521 to indicate an error state
			// same as Cloudflare: https://support.cloudflare.com/hc/en-us/articles/115003011431#521error
			return c.String(521, "Not ready.")
		}

		err := ProbeReadiness(s.MainService.DbClient)
		// Finally return the health status according to the seen states
		if err != nil {
			// We use 521 to indicate this error state
			// same as Cloudflare: https://support.cloudflare.com/hc/en-us/articles/115003011431#521error
			return c.String(521, fmt.Sprintf("Not ready %s ", err.Error()))
		}

		return c.String(http.StatusOK, "Ready.")
	}
}
