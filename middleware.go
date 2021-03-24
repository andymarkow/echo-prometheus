package echoprometheus

import (
	"reflect"
	"strconv"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Config responsible to configure middleware
type Config struct {
	Namespace           string
	Buckets             []float64
	Subsystem           string
	PatternedStatusCode bool
	SingleNotFoundPath  bool
	Skipper             middleware.Skipper
}

//nolint: gomnd
func applyStatusCodePattern(status int) string {
	switch {
	case status < 200:
		return "1xx"
	case status < 300:
		return "2xx"
	case status < 400:
		return "3xx"
	case status < 500:
		return "4xx"
	}
	return "5xx"
}

func isNotFoundHandler(handler echo.HandlerFunc) bool {
	return reflect.ValueOf(handler).Pointer() == reflect.ValueOf(echo.NotFoundHandler).Pointer()
}

// NewConfig returns a new config with default values
/*
	Namespace: "echo",
	Subsystem: "http",
	// There is no need to add a highest bucket with +Inf bound, it will be added implicitly.
	Buckets: []float64{
		0.0005,
		0.001, // 1ms
		0.002,
		0.005,
		0.01, // 10ms
		0.02,
		0.05,
		0.1, // 100 ms
		0.2,
		0.5,
		1.0, // 1s
		2.0,
		5.0,
		10.0, // 10s
		15.0,
		20.0,
		30.0,
	},
	PatternedStatusCode: false, // 200 => 2xx
	SingleNotFoundPath:  false, // group all 404 as "/not-found"
	Skipper:             nil,

*/
func NewConfig() *Config {
	return &Config{
		Namespace: "echo",
		Subsystem: "http",
		// There is no need to add a highest bucket with +Inf bound, it will be added implicitly.
		Buckets: []float64{
			0.0005,
			0.001, // 1ms
			0.002,
			0.005,
			0.01, // 10ms
			0.02,
			0.05,
			0.1, // 100 ms
			0.2,
			0.5,
			1.0, // 1s
			2.0,
			5.0,
			10.0, // 10s
			15.0,
			20.0,
			30.0,
		},
		PatternedStatusCode: false, // 200 => 2xx
		SingleNotFoundPath:  false, // group all 404 as "/not-found"
		Skipper:             nil,
	}
}

// Middleware returns an echo middleware for instrumentation
func Middleware() echo.MiddlewareFunc {
	return MiddlewareWithConfig(NewConfig())
}

// MetricsMiddlewareWithConfig returns an echo middleware for instrumentation with config
func MiddlewareWithConfig(config *Config) echo.MiddlewareFunc {

	requestCounter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "requests_total",
			Help:      "Request counter by status, method and path",
		},
		[]string{"status", "method", "path"})

	requestDuration := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "request_duration_seconds",
			Help:      "Requests duration in seconds",
			Buckets:   config.Buckets,
		},
		[]string{"method", "path"})

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper != nil && config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			path := c.Path()

			// To avoid attack high cardinality of 404
			if config.SingleNotFoundPath && isNotFoundHandler(c.Handler()) {
				path = "/not-found"
			}

			timer := prometheus.NewTimer(requestDuration.WithLabelValues(req.Method, path))
			err := next(c)
			timer.ObserveDuration()

			if err != nil {
				c.Error(err)
			}

			status := strconv.Itoa(c.Response().Status)
			if config.PatternedStatusCode {
				status = applyStatusCodePattern(c.Response().Status)
			}

			requestCounter.WithLabelValues(status, req.Method, path).Inc()

			return err
		}
	}
}
