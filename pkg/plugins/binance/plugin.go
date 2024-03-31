package binance

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"jasonzhu.com/coin_labor/pkg/plugins/general"
)

func init() {
	binance.WebsocketKeepalive = true
	if plugin, err := newBinancePlugin(); err == nil {
		general.Register(&general.ExPlugin{
			ExName:   general.Binance,
			Instance: plugin,
			Ranking:  300,
		})
	}
}

var UseTestnet = false

type BinancePlugin struct {
	baseInfoManager general.BaseInterface
	marketManager   general.MarketInterface
	accountManager  general.AccountInterface
	orderManager    general.OrderInterface
}

func newBinancePlugin() (general.ExManager, error) {
	manager, err := newBaseInfoManager()
	if err != nil {
		fmt.Printf("failed to init plugin [%s]\n", general.Binance)
		return nil, err
	}
	marketManager := newMarketInfoManager()
	accountManager := newAccountManager()
	orderManager := NewOrderManager()
	return &BinancePlugin{
		baseInfoManager: manager,
		marketManager:   marketManager,
		accountManager:  accountManager,
		orderManager:    orderManager,
	}, nil
}

func (p *BinancePlugin) ExchangeAlias() general.Exchange {
	return general.Binance
}

func (p *BinancePlugin) GetBaseInfoManager() general.BaseInterface {
	return p.baseInfoManager
}

func (p *BinancePlugin) GetMarketInfoManager() general.MarketInterface {
	return p.marketManager
}

func (p *BinancePlugin) GetAccountManager() general.AccountInterface {
	return p.accountManager
}

func (p *BinancePlugin) GetOrderInterface() general.OrderInterface {
	return p.orderManager
}
