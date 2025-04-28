package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var (
	OrdersCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "orders_created_total",
		Help: "Total number of orders created successfully",
	})
	OrdersFailedCreationTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "orders_failed_creation_total",
		Help: "Total number of orders failed creation due to business validation",
	})
	OrdersRefundedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "orders_refunded_total",
		Help: "Total number of orders refunded",
	})
	OrdersReturnedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "orders_returned_total",
		Help: "Total number of orders returned",
	})
	OrdersCompletedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "orders_completed_total",
		Help: "Total number of orders completed successfully",
	})

	registerOnce sync.Once
)

func RegisterBusinessMetrics() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			OrdersCreatedTotal,
			OrdersFailedCreationTotal,
			OrdersRefundedTotal,
			OrdersReturnedTotal,
			OrdersCompletedTotal,
		)
	})
}
