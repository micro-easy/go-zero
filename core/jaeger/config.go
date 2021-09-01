package jaeger

type Config struct {
	ServiceName        string  `json:",optional"`
	SamplerType        string  `json:",default=remote"`
	SamplerParam       float64 `json:",default=0.01"`
	LogSpans           bool    `json:",default=false"`
	LocalAgentHostPort string  `json:",default=127.0.0.1:6831"`
	Disabled           bool    `json:",default=false"`
}
