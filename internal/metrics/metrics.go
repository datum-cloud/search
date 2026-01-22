package metrics

import (
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	namespace = "search"
)

var (
	// ExampleResourceTotal tracks the total number of search queries
	ExampleResourceTotal = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Namespace:      namespace,
			Name:           "query_total",
			Help:           "Total number of search queries",
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"status"},
	)

	// ExampleResourceDuration tracks the duration of search queries
	ExampleResourceDuration = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Namespace:      namespace,
			Name:           "query_duration_seconds",
			Help:           "Duration of search queries in seconds",
			StabilityLevel: metrics.ALPHA,
			Buckets:        metrics.ExponentialBuckets(0.001, 2, 14),
		},
		[]string{"operation"},
	)
)

// init registers all custom metrics with the legacy registry
// This ensures they're included in the /metrics endpoint
func init() {
	legacyregistry.MustRegister(
		ExampleResourceTotal,
		ExampleResourceDuration,
	)
}
