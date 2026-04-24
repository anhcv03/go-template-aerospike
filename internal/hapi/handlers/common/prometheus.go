package common

import (
	config "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"

	"github.com/prometheus/client_golang/prometheus"
)

var Registry = prometheus.NewRegistry()

var httpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration",
		Help:    "Duration of HTTP requests in ms",
		Buckets: []float64{100, 300, 500, 1000},
	},
	[]string{"method", "protocol", "path", "status_code", "origin", "ip", "user_id", "user_agent"},
)

func init() {
	prometheus.
		WrapRegistererWith(prometheus.Labels{"service_name": config.SVC_ORDER}, Registry).
		MustRegister(httpRequestDuration)
}

func SetHTTPMetric(
	method string,
	protocol string,
	path string,
	statusCode string,
	origin string,
	ip string,
	userID string,
	userAgent string,
	duration int64,
) {
	httpRequestDuration.
		WithLabelValues(method, protocol, path, statusCode, origin, ip, userID, userAgent).
		Observe(float64(duration))
}

func SetGRPCMetric(
	method string,
	protocol string,
	statusCode string,
	duration int64,
) {
	httpRequestDuration.
		WithLabelValues(method, protocol, "-", statusCode, "-", "-", "-", "-").
		Observe(float64(duration))
}
