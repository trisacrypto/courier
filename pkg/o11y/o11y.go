/*
Package metrics provides functions for exporting prometheus metrics.
*/
package o11y

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	prometheus.MustRegister(
		Passwords,
		Certificates,
		Requests,
		Durations,
		RequestSizeBytes,
		ReplySizeBytes,
	)
}

const (
	Namespace = "trisa"
	Subsystem = "courier"
)

const (
	code   = "code"
	method = "method"
	host   = "host"
	path   = "path"
)

var (
	// Passwords records the number of PKCS12 passwords posted to courier.
	Passwords = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "passwords",
		Help:      "counts the number of PCKS12 passwords successfully posted to courier",
	})

	// Certificates records the number of x509 certs posted to courier.
	Certificates = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "certificates",
		Help:      "counts the number of certificates successfully delivered to courier",
	})

	// Standard HTTP Request Metrics
	Requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "requests",
		Help:      "the number of HTTP requests processed, partitioned by status code and method",
	}, []string{code, method, host, path})

	Durations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "request_duration_seconds",
		Help:      "request latencies in seconds",
	}, []string{code, method, host, path})

	RequestSizeBytes = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "request_size_bytes",
		Help:      "the number of bytes sent in requests",
	}, []string{code, method, host, path})

	ReplySizeBytes = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      "reply_size_bytes",
		Help:      "the number of bytes sent in response to requests",
	}, []string{code, method, host, path})
)

// Prometheus returns the collector endpoint to add to the gin router.
func Prometheus() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// Metrics returns a middleware that can record o11y metrics for all reqeusts.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqSize := computeApproximateRequestSize(c.Request)

		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		labels := []string{
			"", c.Request.Method, c.Request.Host, path,
		}

		c.Next()

		labels[0] = strconv.Itoa(c.Writer.Status())

		repSize := float64(c.Writer.Size())
		dur := float64(time.Since(start)) / float64(time.Second)

		Requests.WithLabelValues(labels...).Inc()
		Durations.WithLabelValues(labels...).Observe(dur)
		RequestSizeBytes.WithLabelValues(labels...).Observe(reqSize)
		ReplySizeBytes.WithLabelValues(labels...).Observe(repSize)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) float64 {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return float64(s)
}
