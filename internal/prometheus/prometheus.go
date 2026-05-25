package prometheus

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusMetrics struct {
	TotalRequests        *prometheus.CounterVec
	RequestsDuration     *prometheus.HistogramVec
	ActiveConnections    prometheus.Gauge
	TotalSearches        prometheus.Counter
	TotalBlockedSearches prometheus.Counter
	TotalStoplistHits    prometheus.Counter
	TopSize              prometheus.Gauge
}

func NewPrometheusMetrics() *PrometheusMetrics {
	totalRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "total_requests",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	prometheus.MustRegister(totalRequests)

	requestsDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "requests_duration",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	prometheus.MustRegister(requestsDuration)

	activeConnections := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Current number of active HTTP connections",
		},
	)
	prometheus.MustRegister(activeConnections)

	totalSearches := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_searches",
			Help: "Total number of search events received",
		},
	)
	prometheus.MustRegister(totalSearches)

	totalBlockedSearches := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_blocked_searches",
			Help: "Total number of spam searches blocked",
		},
	)
	prometheus.MustRegister(totalBlockedSearches)

	totalStoplistHits := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_stoplist_hits",
			Help: "Total number of stoplist hits (blocked words)",
		},
	)
	prometheus.MustRegister(totalStoplistHits)

	topSize := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "top_size",
			Help: "Current number of unique searches in top",
		},
	)
	prometheus.MustRegister(topSize)

	return &PrometheusMetrics{
		TotalRequests:        totalRequests,
		RequestsDuration:     requestsDuration,
		ActiveConnections:    activeConnections,
		TotalSearches:        totalSearches,
		TotalBlockedSearches: totalBlockedSearches,
		TotalStoplistHits:    totalStoplistHits,
		TopSize:              topSize,
	}
}

func (pm *PrometheusMetrics) AddSearch() {
	pm.TotalSearches.Inc()
}

func (pm *PrometheusMetrics) BlockSearch() {
	pm.TotalBlockedSearches.Inc()
}

func (pm *PrometheusMetrics) StoplistHit() {
	pm.TotalStoplistHits.Inc()
}

func (pm *PrometheusMetrics) SetTopSize(size int64) {
	pm.TopSize.Set(float64(size))
}

func (pm *PrometheusMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		pm.ActiveConnections.Inc()
		defer pm.ActiveConnections.Dec()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		pm.TotalRequests.WithLabelValues(c.Request.Method, endpoint, status).Inc()
		pm.RequestsDuration.WithLabelValues(c.Request.Method, endpoint).Observe(duration)
	}
}
