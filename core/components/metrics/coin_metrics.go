package metrics

import . "github.com/prometheus/client_golang/prometheus"

var (
	M_Coin_Opp_pipeline_Counter *CounterVec

	M_Coin_Order_Total *CounterVec
	//M_Coin_Order_Response_Total            *CounterVec
	M_Coin_Order_Executeion_Time_Summary   *SummaryVec
	M_Coin_Order_Executeion_Time_Histogram *HistogramVec

	M_Coin_UserDataWatch_Latency_Summary   *SummaryVec
	M_Coin_UserDataWatch_Latency_Histogram *HistogramVec

	M_Coin_Market_Depth_Total       *CounterVec
	M_Coin_Market_Latency_Summary   *SummaryVec
	M_Coin_Market_Latency_Histogram *HistogramVec
)

func init() {
	M_Coin_Opp_pipeline_Counter = NewCounterVec(
		CounterOpts{
			Name: "coin_opp_pipeline_counter",
		},
		[]string{"symbol", "stage", "description"},
	)

	M_Coin_Order_Total = NewCounterVec(
		CounterOpts{
			Name: "coin_order_total",
		},
		[]string{"exchange", "symbol", "type", "status"},
	)

	M_Coin_Order_Executeion_Time_Summary = NewSummaryVec(
		SummaryOpts{
			Name: "coin_order_execution_time_microsecond_summary",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.8:  0.02,
				0.9:  0.01,
				0.95: 0.005,
				0.99: 0.001,
			},
			//AgeBuckets: 1, // Default is 5 minutes.
		},
		[]string{"exchange", "symbol", "type", "status"},
	)

	M_Coin_Order_Executeion_Time_Histogram = NewHistogramVec(
		HistogramOpts{
			Name: "coin_order_execution_time_microsecond_histogram",
			Buckets: []float64{
				10 * 1000,
				30 * 1000,
				50 * 1000,
				100 * 1000,
				500 * 1000,
				1000 * 1000,
				2000 * 1000,
			},
		},
		[]string{"exchange", "symbol", "type", "status"},
	)

	M_Coin_UserDataWatch_Latency_Summary = NewSummaryVec(
		SummaryOpts{
			Name: "coin_user_data_watch_latency_summary",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.8:  0.02,
				0.9:  0.01,
				0.95: 0.005,
				0.99: 0.001,
			},
			//AgeBuckets: 1, // Default is 5 minutes.
		},
		[]string{"exchange"},
	)

	M_Coin_UserDataWatch_Latency_Histogram = NewHistogramVec(
		HistogramOpts{
			Name: "coin_user_data_watch_latency_histogram",
			Buckets: []float64{
				0.5, 1, 3, 5,
			},
		},
		[]string{"exchange"},
	)

	M_Coin_Market_Depth_Total = NewCounterVec(
		CounterOpts{
			Name: "coin_market_depth_total",
		},
		[]string{"exchange", "trading", "symbol"}, // type: spot / derivatives
	)

	M_Coin_Market_Latency_Summary = NewSummaryVec(
		SummaryOpts{
			Name: "coin_market_latency_microsecond_summary",
			Objectives: map[float64]float64{
				0.5:  0.05,
				0.8:  0.02,
				0.9:  0.01,
				0.95: 0.005,
				0.99: 0.001,
			},
			//AgeBuckets: 1, // Default is 5 minutes.
		},
		[]string{"exchange", "symbol", "type", "status"},
	)

	M_Coin_Market_Latency_Histogram = NewHistogramVec(
		HistogramOpts{
			Name: "coin_market_latency_microsecond_histogram",
			Buckets: []float64{
				10 * 1000,
				30 * 1000,
				50 * 1000,
				100 * 1000,
				500 * 1000,
				1000 * 1000,
				2000 * 1000,
			},
		},
		[]string{"exchange", "symbol", "type", "status"},
	)
}
