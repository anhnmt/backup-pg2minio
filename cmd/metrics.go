package cmd

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Prometheus struct {
	lastBackupTimestamp prometheus.Gauge
	lastBackupSize      prometheus.Gauge
	backupDuration      prometheus.Histogram
	backupStatus        prometheus.GaugeVec
}

var defaultPrometheus atomic.Value

func init() {
	SetDefaultMetrics(&Prometheus{})
}

func DefaultMetrics() *Prometheus {
	return defaultPrometheus.Load().(*Prometheus)
}

func SetDefaultMetrics(p *Prometheus) {
	defaultPrometheus.Store(p)
}

func NewPrometheusMetrics(namespace, subsys string) (*Prometheus, *prometheus.Registry, error) {
	prom := &Prometheus{}
	promReg := prometheus.NewPedanticRegistry()
	errs := make([]error, 0)

	prom.lastBackupTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "backup_last_success_timestamp_seconds",
		Help:      "The timestamp of the last successful backup in seconds",
		Namespace: namespace,
		Subsystem: subsys,
	})

	prom.lastBackupSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "backup_last_success_size_bytes",
		Help:      "The size of the last successful backup in bytes",
		Namespace: namespace,
		Subsystem: subsys,
	})

	prom.backupDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      "backup_duration_seconds",
		Help:      "The duration of the backup",
		Namespace: namespace,
		Subsystem: subsys,
	})

	prom.backupStatus = *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "backup_last_status",
		Help:      "The status of the last backup (1=success, 0=failure).",
		Namespace: namespace,
		Subsystem: subsys,
	}, []string{"status"})

	errs = append(errs,
		promReg.Register(collectors.NewBuildInfoCollector()),
		promReg.Register(collectors.NewGoCollector()),
		promReg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})),
		promReg.Register(prom.lastBackupTimestamp),
		promReg.Register(prom.lastBackupSize),
		promReg.Register(prom.backupDuration),
		promReg.Register(prom.backupStatus),
	)

	if err := errors.Join(errs...); err != nil {
		return nil, nil, fmt.Errorf("registering metrics collectors: %w", err)
	}

	return prom, promReg, nil
}

func LastBackupTimestamp() {
	DefaultMetrics().lastBackupTimestamp.Set(float64(time.Now().Unix()))
}

func LastBackupSize(sizeBytes int64) {
	DefaultMetrics().lastBackupSize.Set(float64(sizeBytes))
}

func BackupDuration(start time.Time) {
	DefaultMetrics().backupDuration.Observe(time.Since(start).Seconds())
}

func SetBackupStatus(success bool) {
	status := 0.0
	if success {
		status = 1.0
	}
	DefaultMetrics().backupStatus.WithLabelValues("success").Set(status)
	DefaultMetrics().backupStatus.WithLabelValues("failure").Set(1.0 - status)
}
