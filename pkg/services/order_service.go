package services

import (
	"context"
	"golang.org/x/sync/errgroup"
	"jasonzhu.com/coin_labor/core/components/bus"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/components/registry"
	"jasonzhu.com/coin_labor/pkg/components/binance_service/market"
	"jasonzhu.com/coin_labor/pkg/components/coin_ex"
	"jasonzhu.com/coin_labor/pkg/components/general"
	"jasonzhu.com/coin_labor/pkg/components/okx"
	"time"
)

const (
	ServiceName = "OrderService"

	Symbol = "ETHUSDT"
)

func init() {
	registry.Register(&registry.Descriptor{
		Name:         ServiceName,
		Instance:     &OrderService{},
		InitPriority: registry.Low,
	})
}

type OrderService struct {
	lg  log.Logger
	Bus bus.Bus `inject:""`

	stopped bool
	stopC   chan struct{}
	doneC   chan struct{}
}

func (s *OrderService) Init() error {
	s.lg = log.New("service.order")
	s.stopped = false
	//binance.UseTestnet = true
	return nil
}

func (s *OrderService) Run(ctx context.Context) (err error) {
	group, _ := errgroup.WithContext(ctx)

	binanceDepthService := market.NewDepthTopService(Symbol)
	s.doneC, s.stopC, err = binanceDepthService.Watch()
	if err != nil {
		s.lg.Error("failed to watch market", "err", err)
		return err
	}

	coinExDepthService := coin_ex.DepthService{}
	coinExDepthService.Init()
	group.Go(func() error {
		return coinExDepthService.Run(ctx)
	})

	okxService := okx.NewMarketClient(Symbol)
	err = okxService.Run(ctx)
	if err != nil {
		return err
	}

	group.Go(func() error {
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				if s.stopped {
					return nil
				}

				var (
					askB, askCEX, askOkx *general.Ask
					bidB, bidCEX, bidOkx *general.Bid
				)
				if binanceDepthService.OK() {
					ask, bid, err := binanceDepthService.Top()
					if err != nil {
						s.lg.Error("Binance, failed to get depth info", "err", err)
						continue
					}
					s.lg.Info("print Binance info", "LastUpdateID", binanceDepthService.LastUpdateID, "top bid", bid.Price, "top ask", ask.Price, "offset", ask.Price-bid.Price, "ask quantity", ask.Quantity, "bid quantity", bid.Quantity)
					askB = ask
					bidB = bid
				} else {
					s.lg.Error("Binance depth info is not ok", "error cnt", binanceDepthService.ErrorCnt)
					continue
				}

				if coinExDepthService.OK() {
					ask, bid, _ := coinExDepthService.Top()
					if ask != nil && bid != nil {
						s.lg.Info("Print CoinEX info", "LastUpdateID", coinExDepthService.LastUpdateID, "top bid", bid.Price, "top ask", ask.Price, "offset", ask.Price-bid.Price, "ask quantity", ask.Quantity, "bid quantity", bid.Quantity)
						askCEX = ask
						bidCEX = bid
					} else {
						askCEX = nil
						bidCEX = nil
					}
				} else {
					askCEX = nil
					bidCEX = nil
					s.lg.Error("coinExDepthService not OK", "errorCnt", coinExDepthService.ErrorCnt)
				}

				if okxService.OK() {
					ask, bid, err := okxService.Top()
					if err != nil {
						s.lg.Error("failed to get depth info of OKX", "err", err)
						askOkx = nil
						bidOkx = nil
					} else {
						s.lg.Info("Print OKX info", "LastUpdateID", okxService.LastUpdateID, "top bid", bid.Price, "top ask", ask.Price, "offset", ask.Price-bid.Price, "ask quantity", ask.Quantity, "bid quantity", bid.Quantity)
						askOkx = ask
						bidOkx = bid
					}
				} else {
					askOkx = nil
					bidOkx = nil
					s.lg.Error("okxService not OK", "errorCnt", okxService.ErrorCnt)
				}

				if askB != nil && bidB != nil {
					if askOkx != nil && bidOkx != nil {
						s.lg.Info("Compare OKX",
							"askB", askB.Price, "bidB", bidB.Price,
							"askOkx", askOkx.Price, "bidOkx", bidOkx.Price,
							"askOKX/bidB", askOkx.Price/bidB.Price, "askB/bidOKX", askB.Price/bidOkx.Price,
							"askOKX-bidB", askOkx.Price-bidB.Price, "askB-bidOKX", askB.Price-bidOkx.Price,
						)
					}
					if askCEX != nil && bidCEX != nil {
						s.lg.Info("Compare CoinEX",
							"askB", askB.Price, "bidB", bidB.Price,
							"askCEX", askCEX.Price, "bidCEX", bidCEX.Price,
							"askCEX/bidB", askCEX.Price/bidB.Price, "askB/bidCEX", askB.Price/bidCEX.Price,
							"askCEX-bidB", askCEX.Price-bidB.Price, "askB-bidCEX", askB.Price-bidCEX.Price,
						)
					}
				}
			}
		}
	})

	s.lg.Info("Started in the background")
	return s.waitingToStop(ctx)
}

func (s *OrderService) waitingToStop(ctx context.Context) error {
	<-ctx.Done()
	s.lg.Info("Stopping background thread")
	s.stopped = true
	s.stopC <- struct{}{}
	<-s.doneC
	s.lg.Info("Stopped")
	return nil
}
