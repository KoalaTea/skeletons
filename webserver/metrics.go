package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func newMetricsServer() *http.Server {
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	return &http.Server{
		// Localhost to seperate unauthenticated metrics endpoint and keep that unauthenticated data from exposure to external
		Addr:    "127.0.0.1:9999",
		Handler: router,
	}
}
