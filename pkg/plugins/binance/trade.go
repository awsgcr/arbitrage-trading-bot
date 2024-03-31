package binance

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"jasonzhu.com/coin_labor/core/setting"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"strconv"
)

type TradesManager struct {
	secret *setting.Secret
	client *binance.Client
}

func NewTradesManager() *TradesManager {
	secret := GetSecretsForExchanger(Binance)
	s := &TradesManager{
		secret: secret,
		client: getBinanceClient(secret),
	}
	return s
}

func (s *TradesManager) ListTrades(symbol Symbol, limit int) ([]*Trade, error) {
	alias := getSymbolAlias(symbol)
	res, err := s.client.NewListTradesService().Symbol(alias).Limit(limit).Do(context.Background())
	if err != nil {
		return nil, err
	}

	var trades []*Trade
	for _, trade := range res {
		trades = append(trades, convertToTrade(trade))
	}

	return trades, nil
}

func convertToTrade(trade *binance.TradeV3) *Trade {
	return &Trade{
		Symbol:          trade.Symbol,
		Id:              strconv.FormatInt(trade.ID, 10),
		OrderId:         strconv.FormatInt(trade.OrderID, 10),
		OrderListId:     trade.OrderListId,
		Price:           NewDecimalFromStringIgnoreErr(trade.Price),
		Qty:             NewDecimalFromStringIgnoreErr(trade.Quantity),
		QuoteQty:        NewDecimalFromStringIgnoreErr(trade.QuoteQuantity),
		Commission:      trade.Commission,
		CommissionAsset: trade.CommissionAsset,
		Time:            trade.Time,
		IsBuyer:         trade.IsBuyer,
		IsMaker:         trade.IsMaker,
		IsBestMatch:     trade.IsBestMatch,
	}
}
