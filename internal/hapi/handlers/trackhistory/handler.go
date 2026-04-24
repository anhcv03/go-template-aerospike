package trackhistory

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/model/httpmodel"
)

func FindByTrackIdRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/track-histories", findByTrackIdHandler(s))
}

// Find track-histories godoc
//
//	@Summary		Find by Id track-histories
//	@Description	Find by Id track-histories
//	@Tags			track-histories
//	@Accept			json
//	@Produce		json
//	@Param			track_id	query		string	true	"track id"
//	@Success		200			{object}	httpmodel.TrackHistory
//	@Failure		500			{object}	httpmodel.ErrorResponse
//	@Router			/track-histories [get]
func findByTrackIdHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		trackIdStr := c.QueryParam("track_id")
		if trackIdStr == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		size := c.QueryParam("size")
		// temporary set to 1000 if size is empty
		numOfHistory := 1000
		if size != "" {
			numOfHistory, _ = strconv.Atoi(size)
		}
		ctx := c.Request().Context()
		trackId, _ := strconv.Atoi(trackIdStr)
		u, err := s.MainService.FindByTrackId(ctx, trackId, numOfHistory)
		if err != nil {
			return c.JSON(http.StatusNotFound, err)
		}

		// return c.JSON(http.StatusOK, u)
		return c.JSON(http.StatusOK, httpmodel.TrackHistoriesListResponse{
			Code: "0",
			Data: httpmodel.TrackHistoriesListData{
				Items:  u,
				Limit:  999999,
				Offset: 0,
				Total:  int64(len(u)),
			},
			Message: "",
		})
	}
}
