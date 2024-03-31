package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	M_DDB_Write_Requests_Total         *prometheus.CounterVec
	M_DDB_Write_Records_Total          *prometheus.CounterVec
	M_DDB_Write_Execution_Time_Summary *prometheus.SummaryVec
	M_DDB_Write_Execution_Time         *prometheus.HistogramVec

	M_DDB_Read_Requests_Total         *prometheus.CounterVec
	M_DDB_Read_Records_Total          *prometheus.CounterVec
	M_DDB_Read_Execution_Time_Summary *prometheus.SummaryVec
	M_DDB_Read_Execution_Time         *prometheus.HistogramVec
)

func init() {
	M_DDB_Write_Requests_Total = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddb_write_requests_total",
			Help: "requests to ddb",
		},
		[]string{"table", "type", "status"},
	)
	M_DDB_Write_Records_Total = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddb_write_records_total",
			Help: "records write to ddb",
		},
		[]string{"table", "type", "status"},
	)
	M_DDB_Write_Execution_Time_Summary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "ddb_write_execution_time_microsecond_summary",
			Help: "summary of ddb write execution duration",
			Objectives: map[float64]float64{
				0.5: 0.05,
				//0.8:  0.02,
				0.9: 0.01,
				//0.95: 0.005,
				0.99: 0.001,
			},
			AgeBuckets: 1, // Default is 5 minutes.
		},
		[]string{"table", "type", "status"},
	)
	M_DDB_Write_Execution_Time = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ddb_write_execution_time_microsecond",
			Help:    "histogram of ddb write execution duration",
			Buckets: []float64{10 * 1000, 20 * 1000, 30 * 1000},
		},
		[]string{"table", "type", "status"},
	)

	// Read
	M_DDB_Read_Requests_Total = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddb_read_requests_total",
			Help: "read requests to ddb",
		},
		[]string{"table", "type", "status"},
	)
	M_DDB_Read_Records_Total = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ddb_read_records_total",
			Help: "read records to ddb",
		},
		[]string{"table", "type", "status"},
	)
	M_DDB_Read_Execution_Time_Summary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "ddb_read_execution_time_microsecond_summary",
			Help: "summary of ddb read execution duration",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
			AgeBuckets: 1, // Default is 5 minutes.
		},
		[]string{"table", "type", "status"},
	)
	M_DDB_Read_Execution_Time = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ddb_read_execution_time_microsecond",
			Help:    "histogram of ddb read execution duration",
			Buckets: []float64{5 * 1000, 9 * 1000, 12 * 1000},
		},
		[]string{"table", "type", "status"},
	)
}

func initAppMetricVars() {
	prometheus.MustRegister(
		M_Coin_Opp_pipeline_Counter,
		M_Coin_Order_Total,
		M_Coin_Order_Executeion_Time_Summary,
		M_Coin_Order_Executeion_Time_Histogram,
		M_Coin_UserDataWatch_Latency_Summary,
		M_Coin_UserDataWatch_Latency_Histogram,
		M_Coin_Market_Depth_Total,
		M_Coin_Market_Latency_Summary,
		M_Coin_Market_Latency_Histogram,
	)
}

func initDDBMetricVars() {
	prometheus.MustRegister(
		M_DDB_Write_Requests_Total,
		M_DDB_Write_Records_Total,
		M_DDB_Write_Execution_Time_Summary,
		M_DDB_Write_Execution_Time,

		M_DDB_Read_Requests_Total,
		M_DDB_Read_Records_Total,
		M_DDB_Read_Execution_Time_Summary,
		M_DDB_Read_Execution_Time,
	)
}
