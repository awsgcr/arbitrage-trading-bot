package metrics

import (
	"context"
	"jasonzhu.com/coin_labor/core/components/registry"
)

func init() {
	registry.RegisterService(&InternalMetricsService{})
	initMetricVars()
	initAppMetricVars()
}

type InternalMetricsService struct {
}

func (s *InternalMetricsService) Init() error {
	return nil
}

func (s *InternalMetricsService) Run(ctx context.Context) error {
	M_Instance_Start.Inc()

	<-ctx.Done()
	return ctx.Err()
}
