package serverinterceptors

import (
	"context"
	"encoding/json"

	"github.com/micro-easy/go-zero/core/jaeger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryOpenTracingInterceptor returns a func that handles tracing with given tracer.
func UnaryOpenTracingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		if !jaeger.Enabled() {
			return handler(ctx, req)
		}
		var spanCtx opentracing.SpanContext
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		spanCtx, _ = jaeger.Tracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(md))
		span := jaeger.Tracer().StartSpan(info.FullMethod, ext.RPCServerOption(spanCtx))
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
		resp, err = handler(ctx, req)
		if err != nil {
			ext.LogError(span, err)
		}
		// set request and reply tag
		reqContent, _ := json.Marshal(req)
		replyContent, _ := json.Marshal(resp)
		span.SetTag("request-info", string(reqContent))
		span.SetTag("reply-info", string(replyContent))
		return resp, err
	}
}
