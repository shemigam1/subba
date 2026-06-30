package observability

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// poolCollector exposes pgx connection-pool saturation as Prometheus gauges. The
// Postgres pool is the first thing to strain under load, so it is always on the
// dashboard.
type poolCollector struct {
	pools                          map[string]*pgxpool.Pool
	acquired, idle, total, maxDesc *prometheus.Desc
}

// NewPoolCollector builds a collector over the named pools (e.g. "tenant", "admin").
func NewPoolCollector(pools map[string]*pgxpool.Pool) prometheus.Collector {
	label := []string{"pool"}
	return &poolCollector{
		pools:    pools,
		acquired: prometheus.NewDesc("pgxpool_acquired_conns", "Connections currently in use.", label, nil),
		idle:     prometheus.NewDesc("pgxpool_idle_conns", "Idle connections.", label, nil),
		total:    prometheus.NewDesc("pgxpool_total_conns", "Total open connections.", label, nil),
		maxDesc:  prometheus.NewDesc("pgxpool_max_conns", "Maximum allowed connections.", label, nil),
	}
}

func (c *poolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquired
	ch <- c.idle
	ch <- c.total
	ch <- c.maxDesc
}

func (c *poolCollector) Collect(ch chan<- prometheus.Metric) {
	for name, p := range c.pools {
		s := p.Stat()
		ch <- prometheus.MustNewConstMetric(c.acquired, prometheus.GaugeValue, float64(s.AcquiredConns()), name)
		ch <- prometheus.MustNewConstMetric(c.idle, prometheus.GaugeValue, float64(s.IdleConns()), name)
		ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(s.TotalConns()), name)
		ch <- prometheus.MustNewConstMetric(c.maxDesc, prometheus.GaugeValue, float64(s.MaxConns()), name)
	}
}
