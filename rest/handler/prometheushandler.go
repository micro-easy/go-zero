package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/micro-easy/go-zero/core/metric"
	"github.com/micro-easy/go-zero/core/timex"
	"github.com/micro-easy/go-zero/rest/internal/security"
)

const serverNamespace = "http_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http server requests duration(ms).",
		Labels:    []string{"path", "service"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http server requests error count.",
		Labels:    []string{"path", "code", "service"},
	})
)

func PromethousHandler(path, serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := timex.Now()
			cw := &security.WithCodeResponseWriter{Writer: w}
			defer func() {
				metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), path, serviceName)
				metricServerReqCodeTotal.Inc(path, strconv.Itoa(cw.Code), serviceName)
			}()

			next.ServeHTTP(cw, r)
		})
	}
}
