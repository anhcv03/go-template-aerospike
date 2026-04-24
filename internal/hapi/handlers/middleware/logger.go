package middleware

import (
	"strconv"
	"time"

	"gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	common "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/hapi/handlers/common"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var httpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration",
		Help:    "Duration of HTTP requests in ms",
		Buckets: []float64{100, 300, 500, 1000},
	},
	[]string{"method", "protocol", "path", "status_code", "origin", "ip", "user_id", "user_agent"},
)

func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		var err error

		defer func() {
			duration := time.Since(start).Milliseconds()
			method := c.Request().Method
			if method == "" {
				method = "-"
			}
			protocol := c.Request().Proto
			if protocol == "" {
				protocol = "-"
			}
			path := c.Request().URL.Path
			if path == "" {
				path = "-"
			}
			statusCode := strconv.Itoa(c.Response().Status)
			if statusCode == "" || statusCode == "0" {
				statusCode = "-"
			}
			contentLength := c.Response().Header().Get("content-length")
			if contentLength == "" || contentLength == "0" {
				contentLength = "-"
			}
			origin := c.Request().Header.Get("origin")
			if origin == "" {
				origin = c.Request().Header.Get("referer")
				if origin == "" {
					origin = "-"
				}
			}
			ip := c.RealIP()
			if ip == "" {
				ip = "-"
			}
			userID := c.Request().Header.Get("userid")
			if userID == "" {
				userID = "-"
			}
			userAgent := c.Request().Header.Get("user-agent")
			if userAgent == "" {
				userAgent = "-"
			}

			ctx := log.Logger.WithContext(c.Request().Context())

			if err != nil {
				config.PrintErrorLog(
					ctx,
					err,
					"%s %s %s %s %d %s %s %s %s",
					method,
					protocol,
					path,
					statusCode,
					duration,
					contentLength,
					origin,
					ip,
					userAgent,
				)
			} else {
				config.PrintDebugLog(
					ctx,
					"%s %s %s %s %d %s %s %s %s",
					method,
					protocol,
					path,
					statusCode,
					duration,
					contentLength,
					origin,
					ip,
					userAgent,
				)
			}

			common.SetHTTPMetric(
				method,
				protocol,
				path,
				statusCode,
				origin,
				ip,
				userID,
				userAgent,
				duration,
			)
		}()

		err = next(c)

		return err
	}
}
