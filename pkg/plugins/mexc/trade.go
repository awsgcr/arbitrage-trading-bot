package mexc

import (
	"context"
	"github.com/bitly/go-simplejson"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/setting"
	. "jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"net/http"
)

type TradeManager struct {
	secret *setting.Secret
	client *Client
	lg     log.Logger
}

func NewTradeManager() *TradeManager {
	secret := GetSecretsForExchanger(MEXC)
	return &TradeManager{
		secret: secret,
		lg:     plg.New("s", "Trade"),
		client: NewHMACClient(secret, baseAPIMainURL, apiKeyHeader),
	}
}

func (s *TradeManager) ListTradesOfSymbol(symbol Symbol, limit int) (res []*Trade, err error) {
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: myTradesEndpoint,
		SecType:  SecTypeSigned,
	}
	r.SetParam("symbol", getSymbolAlias(symbol))
	if limit > 0 {
		r.SetParam("limit", limit)
	}
	data, err := s.client.CallAPI(context.Background(), r)
	if err != nil {
		return nil, err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	size := len(j.MustArray())
	var Trades = make([]*Trade, size)
	for i := 0; i < size; i++ {
		item := j.GetIndex(i)
		Trades[i] = convertToTrade(item)
	}
	return Trades, nil
}

func convertToTrade(item *simplejson.Json) *Trade {
	return &Trade{
		Symbol:          item.Get("symbol").MustString(),
		Id:              item.Get("id").MustString(),
		OrderId:         item.Get("orderId").MustString(),
		OrderListId:     item.Get("orderListId").MustInt64(),
		Price:           NewDecimalFromStringIgnoreErr(item.Get("price").MustString()),
		Qty:             NewDecimalFromStringIgnoreErr(item.Get("qty").MustString()),
		QuoteQty:        NewDecimalFromStringIgnoreErr(item.Get("quoteQty").MustString()),
		Commission:      item.Get("commission").MustString(),
		CommissionAsset: item.Get("commissionAsset").MustString(),
		Time:            item.Get("time").MustInt64(),
		IsBuyer:         item.Get("isBuyer").MustBool(),
		IsMaker:         item.Get("isMaker").MustBool(),
		IsBestMatch:     item.Get("isBestMatch").MustBool(),
		ClientOrderId:   item.Get("clientOrderId").MustString(),
	}
}
