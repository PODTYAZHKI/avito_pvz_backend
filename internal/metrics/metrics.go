package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})

	PvzCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pvz_created_total",
		Help: "Total number of PVZ created",
	})

	ReceptionCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reception_created_total",
		Help: "Total number of receptions created",
	})

	ProductAddedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "product_added_total",
		Help: "Total number of products added",
	})
)

func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":9000", nil); err != nil {
			panic(err)
		}
	}()
}
