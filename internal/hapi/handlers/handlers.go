package handlers

import (
	hapi "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi"
	common "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/handlers/common"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/handlers/track"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/handlers/trackhistory"

	"github.com/labstack/echo/v4"
)

func AttackAllRoutes(s *hapi.Server) {
	s.Router.Routes = []*echo.Route{
		// GET /-/version
		common.GetVersionRoute(s),
		// GET /-/ready
		common.GetReadyRoute(s),
		// GET /-/healthy
		common.GetHealthyRoute(s),

		track.FindAllRoute(s),
		track.FindByIDRoute(s),

		trackhistory.FindByTrackIdRoute(s),
	}
}
