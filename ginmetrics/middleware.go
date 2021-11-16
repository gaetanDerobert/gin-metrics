package ginmetrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gaetanDerobert/gin-metrics/bloom"
)

var (
	metricURIRequestTotal = "gin_url_request_total"

	bloomFilter *bloom.BloomFilter
)

// Use set gin metrics middleware
func (m *Monitor) Use(r gin.IRoutes) {
	m.initGinMetrics()

	r.Use(m.monitorInterceptor)
	r.GET(m.metricPath, func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}

// UseWithoutExposingEndpoint is used to add monitor interceptor to gin router
// It can be called multiple times to intercept from multiple gin.IRoutes
// http path is not set, to do that use Expose function
func (m *Monitor) UseWithoutExposingEndpoint(r gin.IRoutes) {
	m.initGinMetrics()
	r.Use(m.monitorInterceptor)
}

// Expose adds metric path to a given router.
// The router can be different with the one passed to UseWithoutExposingEndpoint.
// This allows to expose metrics on different port.
func (m *Monitor) Expose(r gin.IRoutes) {
	r.GET(m.metricPath, func(ctx *gin.Context) {
		promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
	})
}

// initGinMetrics used to init gin metrics
func (m *Monitor) initGinMetrics() {
	bloomFilter = bloom.NewBloomFilter()
	_ = monitor.AddMetric(&Metric{
		Type:        Counter,
		Name:        metricURIRequestTotal,
		Description: "All request received with their query parameters",
		Labels:      []string{"url", "method", "code"},
	})
}

// monitorInterceptor as gin monitor middleware.
func (m *Monitor) monitorInterceptor(ctx *gin.Context) {
	if ctx.Request.URL.Path == m.metricPath {
		ctx.Next()
		return
	}
	startTime := time.Now()

	// execute normal process.
	ctx.Next()

	// after request
	m.ginMetricHandle(ctx, startTime)
}

func (m *Monitor) ginMetricHandle(ctx *gin.Context, start time.Time) {
	r := ctx.Request
	w := ctx.Writer

	// set uri request total
	_ = m.GetMetric(metricURIRequestTotal).Inc([]string{ctx.Request.URL.String(), r.Method, strconv.Itoa(w.Status())})
}
