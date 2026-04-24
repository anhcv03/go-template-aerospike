package router

import (
	hapi "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi"
	handlers "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/handlers"
	logger "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/handlers/middleware"

	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

func Init(s *hapi.Server) {
	s.Echo = echo.New()

	s.Echo.GET("/swagger/*", echoSwagger.WrapHandler)

	s.Echo.Any("/socket.io/", func(c echo.Context) error {
		s.MainService.SocketServer.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	s.Echo.Debug = s.Config.HttpConfig.EchoDebug
	s.Echo.HideBanner = true

	// if s.Config.JWTTokenConfig.ValidateJwt {
	// 	s.Echo.Use(echojwt.WithConfig(echojwt.Config{
	// 		SigningKey:  []byte(s.Config.JWTTokenConfig.JwtSecret),
	// 		TokenLookup: fmt.Sprintf("cookie:%s", s.Config.JWTTokenConfig.CookieName),
	// 	}))
	// }

	if s.Config.HttpConfig.EnableRecoverMiddleware {
		s.Echo.Use(echoMiddleware.Recover())
	} else {
		log.Warn().Msg("Disabling recover middleware due to environment config")
	}

	if s.Config.HttpConfig.EnableCORSMiddleware {
		s.Echo.Use(middleware.CORSWithConfig(echoMiddleware.CORSConfig{
			AllowOrigins:  []string{"*"},
			AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, "X-Total-Count"},
			ExposeHeaders: []string{"X-Total-Count"},
		}))
	} else {
		log.Warn().Msg("Disabling CORS middleware due to environment config")
	}

	s.Echo.Use(otelecho.Middleware("init-router"))
	s.Echo.Use(logger.LoggerMiddleware)

	s.Echo.GET("/prometheus", echo.WrapHandler(promhttp.Handler()))

	s.Router = &hapi.Router{
		Routes:     nil,
		Root:       s.Echo.Group(""),
		Management: s.Echo.Group("/-"),
	}

	handlers.AttackAllRoutes(s)
}
