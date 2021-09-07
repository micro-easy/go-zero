package logx

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/micro-easy/go-zero/core/timex"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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
		NewTraceLogger(ctx).write(debugLog, levelDebug, fmt.Sprint(v...), traceLoggerCallerInnerDepth, false, DebugLevel)
	}
}

func CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(DebugLevel) {
		NewTraceLogger(ctx).write(debugLog, levelDebug, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, false, DebugLevel)
	}
}

func CtxInfo(ctx context.Context, v ...interface{}) {
	if shouldLog(InfoLevel) {
		NewTraceLogger(ctx).write(infoLog, levelInfo, fmt.Sprint(v...), traceLoggerCallerInnerDepth, false, InfoLevel)
	}
}

func CtxInfof(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		NewTraceLogger(ctx).write(infoLog, levelInfo, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, false, InfoLevel)
	}
}

func CtxWarn(ctx context.Context, v ...interface{}) {
	if shouldLog(WarnLevel) {
		NewTraceLogger(ctx).write(warnLog, levelWarn, fmt.Sprint(v...), traceLoggerCallerInnerDepth, false, WarnLevel)
	}
}

func CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(WarnLevel) {
		NewTraceLogger(ctx).write(warnLog, levelWarn, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, false, WarnLevel)
	}
}

func CtxError(ctx context.Context, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		NewTraceLogger(ctx).write(errorLog, levelError, fmt.Sprint(v...), traceLoggerCallerInnerDepth, false, ErrorLevel)
	}
}

func CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		NewTraceLogger(ctx).write(errorLog, levelError, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, false, ErrorLevel)
	}
}

func CtxSpanDebug(ctx context.Context, v ...interface{}) {
	if shouldLog(DebugLevel) {
		NewTraceLogger(ctx).write(debugLog, levelDebug, fmt.Sprint(v...), traceLoggerCallerInnerDepth, true, DebugLevel)
	}
}

func CtxSpanDebugf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(DebugLevel) {
		NewTraceLogger(ctx).write(debugLog, levelDebug, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, true, DebugLevel)
	}
}

func CtxSpanInfo(ctx context.Context, v ...interface{}) {
	if shouldLog(InfoLevel) {
		NewTraceLogger(ctx).write(infoLog, levelInfo, fmt.Sprint(v...), traceLoggerCallerInnerDepth, true, InfoLevel)
	}
}

func CtxSpanInfof(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		NewTraceLogger(ctx).write(infoLog, levelInfo, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, true, InfoLevel)
	}
}

func CtxSpanWarn(ctx context.Context, v ...interface{}) {
	if shouldLog(WarnLevel) {
		NewTraceLogger(ctx).write(warnLog, levelWarn, fmt.Sprint(v...), traceLoggerCallerInnerDepth, true, WarnLevel)
	}
}

func CtxSpanWarnf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(WarnLevel) {
		NewTraceLogger(ctx).write(warnLog, levelWarn, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, true, WarnLevel)
	}
}

func CtxSpanError(ctx context.Context, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		NewTraceLogger(ctx).write(errorLog, levelError, fmt.Sprint(v...), traceLoggerCallerInnerDepth, true, ErrorLevel)
	}
}

func CtxSpanErrorf(ctx context.Context, format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		NewTraceLogger(ctx).write(errorLog, levelError, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, true, ErrorLevel)
	}
}

func NewTraceLogger(ctx context.Context) *traceLogger {
	return &traceLogger{
		ctx: ctx,
	}
}

func (l *traceLogger) Error(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, fmt.Sprint(v...), traceLoggerCallerInnerDepth, false, ErrorLevel)
	}
}

func (l *traceLogger) Errorf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, false, ErrorLevel)
	}
}

func (l *traceLogger) Info(v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprint(v...), traceLoggerCallerInnerDepth, false, InfoLevel)
	}
}

func (l *traceLogger) Infof(format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, false, InfoLevel)
	}
}

func (l *traceLogger) Slow(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprint(v...), traceLoggerCallerInnerDepth, false, ErrorLevel)
	}
}

func (l *traceLogger) Slowf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprintf(format, v...), traceLoggerCallerInnerDepth, false, ErrorLevel)
	}
}

func (l *traceLogger) WithDuration(duration time.Duration) Logger {
	l.Duration = timex.ReprOfDuration(duration)
	return l
}

func (l *traceLogger) write(writer io.Writer, level, content string, callDepth int, withSpan bool, levelInt int) {
	l.Timestamp = getTimestamp()
	l.Level = level
	l.Content = formatWithCaller(content, callDepth)
	l.Trace, l.Span = getTraceIdSpanId(l.ctx)
	if withSpan {
		spanLog(l.ctx, level, l.Content, levelInt)
	}
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

func spanLog(ctx context.Context, logLevel, logContent string, logLevelInt int) {
	if ctx != nil {
		sp := opentracing.SpanFromContext(ctx)
		if sp != nil {
			sp.LogKV("level", logLevel, "content", logContent)
			if logLevelInt >= ErrorLevel {
				ext.Error.Set(sp, true)
			}
		}
	}
}
