package logx

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/micro-easy/go-zero/core/timex"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
)

const traceLoggerCallerInnerDepth = 4

type traceLogger struct {
	logEntry
	Trace string `json:"trace,omitempty"`
	Span  string `json:"span,omitempty"`
	ctx   context.Context
}

func CtxDebug(ctx context.Context, v ...interface{}) {
	if shouldLog(DebugLevel) {
		NewTraceLogger(ctx).write(debugLog, levelDebug, fmt.Sprint(v...), traceLoggerCallerInnerDepth)
	}
}

func CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(DebugLevel) {
		NewTraceLogger(ctx).write(debugLog, levelDebug, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth)
	}
}

func CtxInfo(ctx context.Context, v ...interface{}) {
	if shouldLog(InfoLevel) {
		NewTraceLogger(ctx).write(infoLog, levelInfo, fmt.Sprint(v...), traceLoggerCallerInnerDepth)
	}
}

func CtxInfof(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		NewTraceLogger(ctx).write(infoLog, levelInfo, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth)
	}
}

func CtxWarn(ctx context.Context, v ...interface{}) {
	if shouldLog(WarnLevel) {
		NewTraceLogger(ctx).write(warnLog, levelWarn, fmt.Sprint(v...), traceLoggerCallerInnerDepth)
	}
}

func CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(WarnLevel) {
		NewTraceLogger(ctx).write(warnLog, levelWarn, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth)
	}
}

func CtxError(ctx context.Context, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		NewTraceLogger(ctx).write(errorLog, levelError, fmt.Sprint(v...), traceLoggerCallerInnerDepth)
	}
}

func CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		NewTraceLogger(ctx).write(errorLog, levelError, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth)
	}
}

func NewTraceLogger(ctx context.Context) *traceLogger {
	return &traceLogger{
		ctx: ctx,
	}
}

func (l *traceLogger) Error(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, fmt.Sprint(v...), traceLoggerCallerInnerDepth)
	}
}

func (l *traceLogger) Errorf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth)
	}
}

func (l *traceLogger) Info(v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprint(v...), traceLoggerCallerInnerDepth)
	}
}

func (l *traceLogger) Infof(format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth)
	}
}

func (l *traceLogger) Slow(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprint(v...), traceLoggerCallerInnerDepth)
	}
}

func (l *traceLogger) Slowf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth)
	}
}

func (l *traceLogger) WithDuration(duration time.Duration) Logger {
	l.Duration = timex.ReprOfDuration(duration)
	return l
}

func (l *traceLogger) write(writer io.Writer, level, content string, callDepth int) {
	l.Timestamp = getTimestamp()
	l.Level = level
	l.Content = formatWithCaller(content, callDepth)
	l.Trace, l.Span = getTraceIdSpanId(l.ctx)
	outputJson(writer, l)
}

func WithContext(ctx context.Context) Logger {
	return &traceLogger{
		ctx: ctx,
	}
}

func getTraceIdSpanId(ctx context.Context) (string, string) {
	if ctx != nil {
		sp := opentracing.SpanFromContext(ctx)
		if sp != nil {
			spCtx := sp.Context()
			if spCtx != nil {
				spanCtx, ok := spCtx.(jaeger.SpanContext)
				if ok {
					return spanCtx.TraceID().String(), spanCtx.SpanID().String()
				}
			}
		}
	}
	return "", ""
}
