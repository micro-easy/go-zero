package handler

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/micro-easy/go-zero/core/jaeger"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func OpenTracingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !jaeger.Enabled() {
			next.ServeHTTP(w, r)
			return
		}
		spanCtx, _ := jaeger.Tracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		span := jaeger.Tracer().StartSpan(r.URL.Path, ext.RPCServerOption(spanCtx))
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(r.Context(), span)
		var buf bytes.Buffer

		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		ow := &opentracingResponseWriter{
			w:    w,
			buf:  &buf,
			code: http.StatusOK,
		}
		next.ServeHTTP(ow, r.WithContext(ctx))
		span.SetTag("request-info", string(bodyBytes))
		span.SetTag("reply-info", buf.String())
		span.SetTag("response-code", ow.code)
	})
}

type opentracingResponseWriter struct {
	w    http.ResponseWriter
	buf  *bytes.Buffer
	code int
}

func (w *opentracingResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *opentracingResponseWriter) Write(bs []byte) (int, error) {
	w.buf.Write(bs)
	return w.w.Write(bs)
}

func (w *opentracingResponseWriter) WriteHeader(code int) {
	w.w.WriteHeader(code)
	w.code = code
}
