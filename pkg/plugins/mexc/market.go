package mexc

import (
	"context"
	"errors"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"strconv"
)

type MarketManager struct {
	GMarketManager
	lg log.Logger
}

func newMarketInfoManager() MarketInterface {
	s := &MarketManager{
		lg: plg.New("s", "market"),
	}
	s.GMarketManager = InitGMarketManager(s.fetchDepth, nil)
	return s
}

/**
API Response Example:
{

  "lastUpdateId": 1377043284,
  "bids": [
        ["30225.77","2.132868"],
        ],
  "asks": [
        ["30225.80","1.130244"],
        ],
}
*/
func (s *MarketManager) fetchDepth(symbol Symbol, limit int) *DepthInfo {
	symbolAlias := getSymbolAlias(symbol)
	if !isAssetSupported(symbol.BaseAsset) {
		return NewDepthInfoWithErr(symbol, errors.New("not supported symbol"))
	}

	params := http.Params{
		"symbol": symbolAlias,
		"limit":  strconv.Itoa(limit),
	}
	j, err := httpGetData(orderBookEndpoint, params)
	if err != nil {
		return NewDepthInfoWithErr(symbol, err)
	}

	asksLen := len(j.Get("asks").MustArray())
	asks := make([]*Ask, asksLen)
	for i := 0; i < asksLen; i++ {
		item := j.Get("asks").GetIndex(i)
		ask, err := NewPriceLevelFromString(item.GetIndex(0).MustString(), item.GetIndex(1).MustString())
		if err != nil {
			continue
		}
		asks[i] = &ask
	}
	bidsLen := len(j.Get("bids").MustArray())
	bids := make([]*Bid, bidsLen)
	for i := 0; i < bidsLen; i++ {
		item := j.Get("bids").GetIndex(i)
		bid, err := NewPriceLevelFromString(item.GetIndex(0).MustString(), item.GetIndex(1).MustString())
		if err != nil {
			continue
		}
		bids[i] = &bid
	}

	return &DepthInfo{
		Symbol:       symbol,
		Asks:         asks,
		Bids:         bids,
		LastUpdateID: j.Get("lastUpdateId").MustInt64(),
		Err:          nil,
	}
}

func (s *MarketManager) wsWatchDepth(ctx context.Context, infoC chan *DepthInfo, limit int, symbols ...Symbol) error {
	return nil
}
