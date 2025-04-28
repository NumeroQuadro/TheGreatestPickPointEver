package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/config"
	"log"
	"net/http"
	"sync"
)

func StartMetricsServer(cfg *config.Config) {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting metrics server on %s", cfg.MetricsPort)
	if err := http.ListenAndServe(cfg.MetricsPort, nil); err != nil {
		panic(err)
	}
}

var (
	RpcCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_server_ok_error_rate",
		Help: "Total number of RPCs handled on the server",
	}, []string{"method", "code"})

	GrpcRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests handled by the server",
	})

	GrpcErrorsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "grpc_errors_total",
		Help: "Total number of gRPC errors",
	})

	ResponseTimeSummary = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "grpc_response_time_seconds",
		Help: "Summary of gRPC response times",
	})

	RequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "Request duration in seconds",
		Buckets: []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1, 2, 5},
	})

	GrpcRequestCountByStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_request_count",
			Help: "Number of gRPC requests labeled by status code",
		},
		[]string{"status"},
	)

	registerOnce sync.Once
)

func RegisterBusinessMetrics() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			GrpcRequestsTotal,
			GrpcErrorsTotal,
			ResponseTimeSummary,
			RequestDuration,
			GrpcRequestCountByStatus,
			RpcCounter,
		)
	})
}
