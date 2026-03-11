package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var Registry = prometheus.NewRegistry()

func init() {
	Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		RequestDuration,
		RequestCount,
	)
}

var (
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "web3_wallet",
			Subsystem: "go_wallet_api",
			Name:      "request_duration",
			Help:      "request duration",
		},
		[]string{"path"},
	)

	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "web3_wallet",
			Subsystem: "go_wallet_api",
			Name:      "http_requests_total",
			Help:      "http request total",
		},
		[]string{"path", "method"},
	)
)
