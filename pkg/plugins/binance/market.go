package binance

import (
	"context"
	"errors"
	"github.com/adshao/go-binance/v2"
	"jasonzhu.com/coin_labor/core/components/alerting"
	"jasonzhu.com/coin_labor/core/components/metrics"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"strconv"
)

var lg = plg.New("s", "market")

type MarketManager struct {
	GMarketManager
	client *binance.Client
}

func newMarketInfoManager() MarketInterface {
	s := &MarketManager{}
	s.GMarketManager = InitGMarketManager(s.fetchDepth, s.wsWatchDepth)
	s.client = getBinanceClient(nil)
	return s
}

func (s *MarketManager) fetchDepth(symbol Symbol, limit int) *DepthInfo {
	symbol2USDT := getSymbolAlias(symbol)
	if !isAssetSupported(symbol) {
		return NewDepthInfoWithErr(symbol, errors.New("not supported symbol"))
	}

	res, err := s.client.NewDepthService().Symbol(symbol2USDT).Limit(limit).Do(context.Background())
	if err != nil {
		return NewDepthInfoWithErr(symbol, err)
	}

	asksLen := len(res.Asks)
	asks := make([]*Ask, asksLen)
	for i := 0; i < asksLen; i++ {
		item := res.Asks[i]
		ask, err := NewPriceLevelFromString(item.Price, item.Quantity)
		if err != nil {
			continue
		}
		asks[i] = &ask
	}
	bidsLen := len(res.Bids)
	bids := make([]*Bid, bidsLen)
	for i := 0; i < bidsLen; i++ {
		item := res.Bids[i]
		bid, err := NewPriceLevelFromString(item.Price, item.Quantity)
		if err != nil {
			continue
		}
		bids[i] = &bid
	}

	return &DepthInfo{
		Symbol:       symbol,
		Asks:         asks,
		Bids:         bids,
		LastUpdateID: res.LastUpdateID,
		Err:          nil,
	}

}

func (s *MarketManager) wsWatchDepth(ctx context.Context, infoC chan *DepthInfo, limit int, symbols ...Symbol) error {
	wsDepthHandler := func(event *WsPartialDepthEvent) {
		var bids []*Bid
		var asks []*Ask
		for _, bid := range event.Bids {
			bb := bid
			bids = append(bids, &bb)
		}
		for _, ask := range event.Asks {
			aa := ask
			asks = append(asks, &aa)
		}
		info := &DepthInfo{
			Symbol:       newSymbolFromString(event.Symbol),
			LastUpdateID: event.LastUpdateID,
			Bids:         bids,
			Asks:         asks,
			Err:          nil,
		}
		lg.Debug("watch top depth with updating", "LastUpdateID", event.LastUpdateID, "len(bid)", len(bids), "len(ask)", len(asks))
		infoC <- info
		go func() {
			metrics.M_Coin_Market_Depth_Total.WithLabelValues(string(Binance), TradingSpot, string(info.Symbol.BaseAsset)).Inc()
		}()
	}
	errHandler := func(err error) {
		lg.Error("failed to fetch new message from binance websocket.", "err", err)
		DefaultHealthChecker.Declare(BinanceMarketDepthWatchFeature, HealthStateUnhealthy)
		alerting.NotifyRightNow(err, "error occurred when fetching Market Depth from binance websocket.")
	}
	var wsServe *WsServe
	var err error
	if len(symbols) == 1 {
		symbol2USDT := getSymbolAlias(symbols[0])
		wsServe, err = WsPartialDepthServe100Ms(symbol2USDT, strconv.Itoa(limit), wsDepthHandler, errHandler)
	} else {
		symbolLevels := make(map[string]string)
		for _, symbol := range symbols {
			symbolLevels[getSymbolAlias(symbol)] = strconv.Itoa(limit)
		}
		wsServe, err = WsCombinedPartialDepthServe100Ms(symbolLevels, wsDepthHandler, errHandler)
	}
	if err != nil {
		return err
	}

	DefaultHealthChecker.Declare(BinanceMarketDepthWatchFeature, HealthStateHealthy)
	// waiting stop signal
	<-wsServe.DoneC()
	DefaultHealthChecker.Declare(BinanceMarketDepthWatchFeature, HealthStateUnhealthy)
	return nil
}
