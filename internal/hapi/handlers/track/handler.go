package track

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/model/httpmodel"
	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/types"
)

// func CreateRoute(s *hapi.Server) *echo.Route {
// 	return s.Router.Root.POST("/tracks", createHandler(s))
// }

// func DeleteByIDRoute(s *hapi.Server) *echo.Route {
// 	return s.Router.Root.DELETE("/tracks/:id", deleteByIDHandler(s))
// }

// func PatchByIDRoute(s *hapi.Server) *echo.Route {
// 	return s.Router.Root.PATCH("/tracks/:id", patchByIDHandler(s))
// }

func FindByIDRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/tracks/:id", findByIDHandler(s))
}

func FindAllRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/tracks", findAllHandler(s))
}

// Find all track godoc
//
//	@Summary		Find all track
//	@Description	Find all track
//	@Tags			tracks
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	pb.Track
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/tracks [get]
func findAllHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find all tracks")

		u, err := s.MainService.FindTrackAll(ctx)
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find track all")

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, httpmodel.TracksListResponse{
			Code: "0",
			Data: httpmodel.TracksListData{
				Items:  u,
				Limit:  999999,
				Offset: 0,
				Total:  int64(len(u)),
			},
			Message: "",
		})
	}
}

// Find drone by ID godoc
//
//	@Summary		Find drone by ID
//	@Description	Find drone by ID
//	@Tags			drones
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"drone id"
//	@Success		200	{object}	pb.Drone
//	@Failure		400	{object}	types.ErrorResponse
//	@Router			/drones/{id} [get]
func findByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": " invalid id",
			})
		}
		requestID := uuid.NewString()
		ctx := log.With().Str("x-request-id", requestID).Logger().WithContext(c.Request().Context())
		c.Response().Header().Set("x-request-id", requestID)

		config.PrintDebugLog(ctx, "Find track by id: %d", id)

		u, err := s.MainService.FindTrackByID(ctx, int32(id))
		if err != nil {
			config.PrintErrorLog(ctx, err, "Failed to find track by id: %d", id)

			return c.JSON(http.StatusNotFound, types.ErrorResponse{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			})
		}
		return c.JSON(http.StatusOK, httpmodel.TrackResponse{
			Code: "0",
			Data: u,
			Message: "",
		})
	}
}