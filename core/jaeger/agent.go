package jaeger

import (
	"fmt"
	"sync"
	"time"

	"github.com/micro-easy/go-zero/core/proc"
	"github.com/micro-easy/go-zero/core/syncx"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

var (
	once         sync.Once
	enabled      syncx.AtomicBool
	jaegerTracer opentracing.Tracer
)

// Enabled returns if jaeger is enabled.
func Enabled() bool {
	return enabled.True()
}

func Tracer() opentracing.Tracer {
	return jaegerTracer
}

// StartAgent starts a jaeger agent.
func StartAgent(conf Config) {
	once.Do(func() {
		cfg := &config.Configuration{
			Disabled:    conf.Disabled,
			ServiceName: conf.ServiceName,
			Sampler: &config.SamplerConfig{
				Type:                    conf.SamplerType,
				Param:                   conf.SamplerParam,
				SamplingRefreshInterval: time.Duration(conf.SamplingRefreshInterval) * time.Second,
			},
			Reporter: &config.ReporterConfig{
				LogSpans:           conf.LogSpans,
				LocalAgentHostPort: conf.LocalAgentHostPort,
			},
		}

		tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger), config.MaxTagValueLength(int(conf.MaxTagValueLength)))
		if err != nil {
			panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
		}

		jaegerTracer = tracer

		proc.AddShutdownListener(func() {
			closer.Close()
		})
		enabled.Set(true)
	})
}
