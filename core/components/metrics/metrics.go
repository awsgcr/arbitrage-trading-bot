package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"jasonzhu.com/coin_labor/core/setting"
)

/**

@author Jason
@version 2021-10-05 16:07
*/

var exporterName = setting.ApplicationName

var (
	M_Instance_Start prometheus.Counter

	M_DB_DataSource_QueryById prometheus.Counter

	// Timers
	M_Alerting_Execution_Time prometheus.Summary

	// M_Pipe_Version is a gauge that contains build info about this binary
	M_Pipe_Build_Version *prometheus.GaugeVec
)

func init() {
	M_Instance_Start = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "instance_start_total",
		Help:      "counter for started instances",
		Namespace: exporterName,
	})

	M_DB_DataSource_QueryById = newCounterStartingAtZero(prometheus.CounterOpts{
		Name:      "db_datasource_query_by_id_total",
		Help:      "counter for getting datasource by id",
		Namespace: exporterName,
	})

	M_Alerting_Execution_Time = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:      "alerting_execution_time_milliseconds",
		Help:      "summary of alert exeuction duration",
		Namespace: exporterName,
	})

	M_Pipe_Build_Version = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "build_info",
		Help:      "A metric with a constant '1' value labeled by version, revision, branch, and goversion from which Pipe was built.",
		Namespace: exporterName,
	}, []string{"version", "revision", "branch", "goversion", "edition"})
}

func initMetricVars() {
	prometheus.MustRegister(
		M_Instance_Start,
		M_Alerting_Execution_Time,
		M_DB_DataSource_QueryById,
		M_Pipe_Build_Version)

}

func newCounterVecStartingAtZero(opts prometheus.CounterOpts, labels []string, labelValues ...string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(opts, labels)

	for _, label := range labelValues {
		counter.WithLabelValues(label).Add(0)
	}

	return counter
}

func newCounterStartingAtZero(opts prometheus.CounterOpts, labelValues ...string) prometheus.Counter {
	counter := prometheus.NewCounter(opts)
	counter.Add(0)

	return counter
}
