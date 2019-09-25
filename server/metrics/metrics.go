package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/runatlantis/atlantis/server/logging"

	"net/http"
)

var (
	OpsProcessedHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "atlantis_ops_seconds",
			Help:    "Histogram for plan/apply operations",
			Buckets: []float64{5, 10, 30, 60, 300, 600},
		},
		[]string{"repo", "directory", "workspace", "command", "success"},
	)
	OpsInProgress = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "atlantis_ops_in_progress_total",
			Help: "The total number of in-progress plan/apply operations",
		},
		[]string{"repo", "directory", "workspace", "command"},
	)
)

func Init(logger *logging.SimpleLogger) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		logger.Info("Metrics server started - listening on port 2122")
		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			logger.Err("unable to start metrics server: %s", err)
		}
	}()
}
