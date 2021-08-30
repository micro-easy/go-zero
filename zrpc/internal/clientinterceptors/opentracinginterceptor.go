package clientinterceptors

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/micro-easy/go-zero/core/jaeger"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// opentracing interceptor is an interceptor that controls tracing.
func OpenTracingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption)   error {
		if !jaeger.Enabled() {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		span, _ := opentracing.StartSpanFromContextWithTracer(ctx, jaeger.Tracer(), method)
		defer span.Finish()
		ext.SpanKindRPCClient.Set(span)

		ctx = injectSpanContext(ctx, span)
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			ext.LogError(span, err)
		}
		reqContent, _ := json.Marshal(req)
		replyContent, _ := json.Marshal(reply)
		span.SetTag("request-info", string(reqContent))
		span.SetTag("reply-info", string(replyContent))
		return err
	}
}

func injectSpanContext(ctx context.Context, span opentracing.Span) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	mdWriter := metadataReaderWriter{md}
	span.Tracer().Inject(span.Context(),
		opentracing.HTTPHeaders,
		mdWriter)
	return metadata.NewOutgoingContext(ctx, md)
}

type metadataReaderWriter struct {
	metadata.MD
}

func (w metadataReaderWriter) Set(key, val string) {
	// The GRPC HPACK implementation rejects any uppercase keys here.
	//
	// As such, since the HTTP_HEADERS format is case-insensitive anyway, we
	// blindly lowercase the key (which is guaranteed to work in the
	// Inject/Extract sense per the OpenTracing spec).
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

func (w metadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}
