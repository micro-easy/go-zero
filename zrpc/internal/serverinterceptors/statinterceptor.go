package serverinterceptors

import (
	"context"
	"encoding/json"
	"time"

	"github.com/micro-easy/go-zero/core/logx"
	"github.com/micro-easy/go-zero/core/stat"
	"github.com/micro-easy/go-zero/core/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const serverSlowThreshold = time.Millisecond * 500

func UnaryStatInterceptor(metrics *stat.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {

		var reply interface{}
		startTime := timex.Now()
		defer func() {
			duration := timex.Since(startTime)
			metrics.Add(stat.Task{
				Duration: duration,
			})
			logDuration(ctx, info.FullMethod, req, reply, duration, err)
		}()

		reply, err = handler(ctx, req)
		return
	}
}

func logDuration(ctx context.Context, method string, req, reply interface{}, duration time.Duration, retErr error) {
	var addr string
	client, ok := peer.FromContext(ctx)
	if ok {
		addr = client.Addr.String()
	}

	reqBytes, _ := json.Marshal(req)
	replyBytes, _ := json.Marshal(reply)

	if retErr != nil {
		logx.WithContext(ctx).WithDuration(duration).Errorf("%s - %s - %s - %v", addr, method, string(reqBytes), retErr)
	} else if duration > serverSlowThreshold {
		logx.WithContext(ctx).WithDuration(duration).Slowf("[RPC] slowcall - %s - %s - %s - %v",
			addr, method, string(reqBytes), string(replyBytes))
	} else {
		logx.WithContext(ctx).WithDuration(duration).Infof("%s - %s - %s - %v", addr, method, string(reqBytes), string(replyBytes))
	}
}
