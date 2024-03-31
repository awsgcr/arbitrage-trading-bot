package general

import (
	"context"
	"golang.org/x/sync/errgroup"
	"jasonzhu.com/coin_labor/core/components/bus"
	"jasonzhu.com/coin_labor/core/components/log"
	"sync"
	"time"
)

type HealthState bool
type ExchangeFeature int

const (
	HealthStateUnhealthy HealthState = false
	HealthStateHealthy   HealthState = true
)
const (
	BinanceUserDataWatchFeature ExchangeFeature = iota
	BinanceMarketDepthWatchFeature
	MEXCUserDataWatchFeature
	MEXCMarketDepthWatchFeature
)

var DefaultHealthChecker = newHealthChecker()

type HealthData struct {
	ExchangeFeature ExchangeFeature
	State           HealthState
	Time            time.Time
}

type HealthReport struct {
	State HealthState
	Time  time.Time
}

type HealthChecker struct {
	lg               log.Logger
	ctx              context.Context
	healthDataMap    map[ExchangeFeature]HealthState
	healthDataMapRWM sync.RWMutex

	healthDataReceiverC chan HealthData
}

func newHealthReport(state HealthState) *HealthReport {
	return &HealthReport{
		State: state,
		Time:  time.Now(),
	}
}

func newHealthChecker() *HealthChecker {
	h := &HealthChecker{
		lg:  log.New("health_checker"),
		ctx: context.Background(),
		healthDataMap: map[ExchangeFeature]HealthState{
			BinanceUserDataWatchFeature:    HealthStateUnhealthy,
			BinanceMarketDepthWatchFeature: HealthStateUnhealthy,
			MEXCUserDataWatchFeature:       HealthStateUnhealthy,
			MEXCMarketDepthWatchFeature:    HealthStateUnhealthy,
		},
		healthDataReceiverC: make(chan HealthData),
	}
	h.run()
	return h
}

func (s *HealthChecker) Declare(feature ExchangeFeature, state HealthState) {
	data := HealthData{
		ExchangeFeature: feature,
		State:           state,
		Time:            time.Now(),
	}
	go func() {
		s.healthDataReceiverC <- data
	}()
}

func (s *HealthChecker) run() {
	group, _ := errgroup.WithContext(s.ctx)

	ticker := time.NewTicker(30 * time.Second)
	group.Go(func() error {
		for {
			select {
			case data := <-s.healthDataReceiverC:
				s.handleHealthData(data)
			case <-ticker.C:
				// heartbeat of HealthChecker for every 30 seconds
				s.lg.Info("heartbeat of HealthChecker")
			}
		}
	})
}

func (s *HealthChecker) handleHealthData(data HealthData) {
	s.healthDataMapRWM.Lock()
	defer s.healthDataMapRWM.Unlock()
	s.healthDataMap[data.ExchangeFeature] = data.State
	s.lg.Info("got health data", "exchangeFeature", data.ExchangeFeature, "state", data.State, "time", data.Time)

	if data.State == HealthStateUnhealthy {
		_ = bus.Publish(newHealthReport(HealthStateUnhealthy))
	} else if data.State == HealthStateHealthy {
		// confirm if all features are healthy, maybe for auto-recovery in the future
		if s.isAllFeaturesHealthy() {
			_ = bus.Publish(newHealthReport(HealthStateHealthy))
		}
	}
}

func (s *HealthChecker) IsAllFeaturesHealthy() bool {
	s.healthDataMapRWM.RLock()
	defer s.healthDataMapRWM.RUnlock()
	return s.isAllFeaturesHealthy()
}

func (s *HealthChecker) isAllFeaturesHealthy() bool {
	if s.healthDataMap[BinanceUserDataWatchFeature] == HealthStateHealthy &&
		s.healthDataMap[BinanceMarketDepthWatchFeature] == HealthStateHealthy &&
		s.healthDataMap[MEXCUserDataWatchFeature] == HealthStateHealthy {
		return true
	}
	return false
}
