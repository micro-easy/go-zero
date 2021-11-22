package serverinterceptors

import (
	"context"
	"strconv"
	"time"

	"github.com/micro-easy/go-zero/core/metric"
	"github.com/micro-easy/go-zero/core/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const serverNamespace = "rpc_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rpc server requests duration(ms).",
		Labels:    []string{"method", "service"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "rpc server requests code count.",
		Labels:    []string{"method", "code", "service"},
	})
)

func UnaryPrometheusInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		interface{}, error) {
		startTime := timex.Now()
		resp, err := handler(ctx, req)
		metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), info.FullMethod, serviceName)
		metricServerReqCodeTotal.Inc(info.FullMethod, strconv.Itoa(int(status.Code(err))), serviceName)
		return resp, err
	}
}
