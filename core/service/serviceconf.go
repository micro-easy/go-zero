package service

import (
	"log"

	"github.com/micro-easy/go-zero/core/jaeger"
	"github.com/micro-easy/go-zero/core/load"
	"github.com/micro-easy/go-zero/core/logx"
	"github.com/micro-easy/go-zero/core/prometheus"
	"github.com/micro-easy/go-zero/core/stat"
)

const (
	DevMode  = "dev"
	TestMode = "test"
	PreMode  = "pre"
	ProMode  = "pro"
)

type ServiceConf struct {
	Name       string
	Log        logx.LogConf
	Mode       string            `json:",default=pro,options=dev|test|pre|pro"`
	MetricsUrl string            `json:",optional"`
	Prometheus prometheus.Config `json:",optional"`
	Jaeger     jaeger.Config     `json:",optional"`
}

func (sc ServiceConf) MustSetUp() {
	if err := sc.SetUp(); err != nil {
		log.Fatal(err)
	}
}

func (sc ServiceConf) SetUp() error {
	if len(sc.Log.ServiceName) == 0 {
		sc.Log.ServiceName = sc.Name
	}

	if len(sc.Jaeger.ServiceName) == 0 {
		sc.Jaeger.ServiceName = sc.Name
	}

	if err := logx.SetUp(sc.Log); err != nil {
		return err
	}

	sc.initMode()
	prometheus.StartAgent(sc.Prometheus)
	jaeger.StartAgent(sc.Jaeger)

	if len(sc.MetricsUrl) > 0 {
		stat.SetReportWriter(stat.NewRemoteWriter(sc.MetricsUrl))
	}

	return nil
}

func (sc ServiceConf) initMode() {
	switch sc.Mode {
	case DevMode, TestMode, PreMode:
		load.Disable()
		stat.SetReporter(nil)
	}
}
