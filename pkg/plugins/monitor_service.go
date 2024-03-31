package plugins

import (
	"context"
	"golang.org/x/sync/errgroup"
	"jasonzhu.com/coin_labor/core/components/bus"
	"jasonzhu.com/coin_labor/core/components/log"
	_ "jasonzhu.com/coin_labor/pkg/plugins/binance"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"time"
)

const (
	ServiceName = "ExMonitorService"
)

//func init() {
//	registry.Register(&registry.Descriptor{
//		Name:         ServiceName,
//		Instance:     &MonitorService{},
//		InitPriority: registry.Low,
//	})
//}

type MonitorService struct {
	lg  log.Logger
	Bus bus.Bus `inject:""`

	stopped bool
}

var watchingAssets = []Asset{
	ETH,
	BNB,
	YFI,
	SHIB,
	N_1INCH,
	UNI,
	AAVE,
	ALICE,
	AXS,
	COMP,
	ENJ,
	SAND,
	OMG,
	MANA,
	LINK,
	SNX,
}

func (s *MonitorService) Init() error {
	s.lg = log.New("service.monitor")
	s.stopped = false
	return nil
}

func (s *MonitorService) Run(ctx context.Context) (err error) {
	group, _ := errgroup.WithContext(ctx)

	binanceExchange := GetExPluginByExchange(Binance)
	binanceMarket := binanceExchange.GetMarketInfoManager()

	mexcExchange := GetExPluginByExchange(MEXC)
	mexcMarket := mexcExchange.GetMarketInfoManager()

	group.Go(func() error {
		ticker := time.NewTicker(time.Duration(500*len(watchingAssets)) * time.Millisecond)
		for range ticker.C {
			if s.stopped {
				break
			}

			for _, ass := range watchingAssets {
				asset := ass
				symbol := Symbol{
					BaseAsset:  asset,
					QuoteAsset: DefaultQuoteCoin,
				}
				group.Go(func() error {
					depth, err := mexcMarket.FetchDepth(symbol, 5)
					if err != nil {
						s.lg.Error("failed to fetch depth for MEXC", "err", err)
						return err
					}
					reference, err := binanceMarket.FetchDepth(symbol, 5)
					if err != nil {
						s.lg.Error("failed to fetch depth for Binance", "err", err)
						return err
					}
					s.compareWithReference(reference, depth)
					return nil
				})
			}
		}
		return nil
	})

	return s.waitingToStop(ctx)
}

func (s *MonitorService) compareWithReference(reference *DepthInfo, target *DepthInfo) {
	if reference == nil || target == nil {
		return
	}
	askB, bidB, _ := reference.Top()
	askBV, bidBV, _ := target.Top()
	if askB != nil && bidB != nil && askBV != nil && bidBV != nil {
		s.lg.Info("Compare MEXC",
			"symbol", reference.Symbol,
			"askB", askB.Price, "bidB", bidB.Price,
			"askBV", askBV.Price, "bidBV", bidBV.Price,
			"askBV/bidB", askBV.Price.Div(bidB.Price), "askB/bidBV", askB.Price.Div(bidBV.Price),
			"askBV-bidB", askBV.Price.Sub(bidB.Price), "askB-bidBV", askB.Price.Sub(bidBV.Price),
		)
	}
}

func (s *MonitorService) waitingToStop(ctx context.Context) error {
	<-ctx.Done()
	s.lg.Info("Stopping background thread")
	s.stopped = true
	s.lg.Info("Stopped")
	return nil
}
