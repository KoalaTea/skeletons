package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "A counter for requests from the wrapped client.",
		},
		[]string{"request_uri", "code", "method"},
	)
	// histVec has no labels, making it a zero-dimensional ObserverVec.
	httpLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_requests_latency",
			Help:    "A histogram of request latencies.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"request_uri", "method"},
	)
	inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(httpRequests, httpLatency, inFlightGauge)
}

func instrumentHttpMetrics(request_uri string, handler http.Handler) http.Handler {
	return promhttp.InstrumentHandlerInFlight(inFlightGauge,
		promhttp.InstrumentHandlerDuration(httpLatency.MustCurryWith(prometheus.Labels{"request_uri": request_uri}),
			promhttp.InstrumentHandlerCounter(httpRequests.MustCurryWith(prometheus.Labels{"request_uri": request_uri}), handler)))
}

func newMetricsServer() *http.Server {
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	return &http.Server{
		// Localhost to seperate unauthenticated metrics endpoint and keep that unauthenticated data from exposure to external
		Addr:    "127.0.0.1:9999",
		Handler: router,
	}
}
