package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Registry = prometheus.NewRegistry()

var (
	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "manage_company_http_requests_total",
			Help: "Tổng số HTTP request đã xử lý.",
		},
		[]string{"method", "route", "status"},
	)

	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "manage_company_http_request_duration_seconds",
			Help:    "Độ trễ HTTP request, tính bằng giây.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route"},
	)
)

var (
	WSConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "manage_company_ws_online_employees",
		Help: "Số nhân viên đang có ít nhất một kết nối trên instance này.",
	})

	WSOnlineEmployees = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "manage_company_ws_online_employees",
		Help: "Số nhân viên đang có ít nhất một kết nối trên instance này.",
	})

	WSMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "manage_company_ws_messages_total",
			Help: "Tổng số bản tin WebSocket",
		},
		[]string{"direction", "type"},
	)

	WSDropped = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "manage_company_ws_dropped_total",
		Help: "Số lần đóng kết nối vì hàng đợi gửi đầy.",
	})
)

var (
	JobsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "manage_jobs_processed_total",
			Help: "Tổng số job nền đã xử lý",
		},
		[]string{"job", "result"},
	)

	JobDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mamage_company_duration_seconds",
			Help:    "Thời gian xử lý job nên, tính bằng giây.",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300},
		},
		[]string{"job"},
	)

	QueueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "manage_company_depth",
			Help: "Số message đang chờ trong hàng đợi.",
		},
		[]string{"queue"},
	)
)

var (
	ScheduledJobRuns = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "manage_company_scheduled_job_runs_total",
			Help: "Tổng số lần chạy job định kỳ.",
		},
		[]string{"job", "result"},
	)

	ScheduledJobLastSuccess = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "manage_company_scheduled_job_last_success_timestamp_seconds",
			Help: "Thời điểm job định kỳ thành công gần nhất (Unix timestamp).",
		},
		[]string{"job"},
	)
)

var BuildInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "manage_company_build_info",
		Help: "Thông tin build, luôn bằng 1; dữ liệu nằm ở nhãn.",
	},
	[]string{"version", "git_sha", "build_time", "app"},
)

func init() {
	Registry.MustRegister(
		HTTPRequests, HTTPDuration,
		WSConnections, WSOnlineEmployees, WSMessages, WSDropped,
		JobsProcessed, JobDuration, QueueDepth,
		ScheduledJobRuns, ScheduledJobLastSuccess,
		BuildInfo,

		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}

func SetBuildInfo(app, version, gitSHA, buildTime string) {
	BuildInfo.WithLabelValues(version, gitSHA, buildTime, app).Set(1)
}

func Handler() http.Handler {
	return promhttp.HandlerFor(Registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	})
}

