package binance

import (
	"fmt"
	"jasonzhu.com/coin_labor/core/util"
	"jasonzhu.com/coin_labor/pkg/plugins/general"
	"testing"
)

func TestTradesManager_ListTrades(t *testing.T) {
	manager := NewTradesManager()
	trades, err := manager.ListTrades(general.NewSymbol(general.INJ), 10)
	if err != nil {
		fmt.Println(err)
	}

	start, _ := util.GetMsOfDateTime("2023-04-30 14:54:00")
	end, _ := util.GetMsOfDateTime("2025-04-27 18:18:18")
	for _, trade := range trades {
		day := util.UnixMillToStr(trade.Time)
		if trade.Time > start && trade.Time < end {
			fmt.Println(day, trade.Symbol, trade.IsBuyer, trade.Price, trade.Qty, trade.QuoteQty)
		}
	}
}
