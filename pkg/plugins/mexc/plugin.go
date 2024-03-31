package mexc

import (
	"fmt"
	"jasonzhu.com/coin_labor/pkg/plugins/general"
)

func init() {
	if plugin, err := NewMEXCPlugin(); err == nil {
		general.Register(&general.ExPlugin{
			ExName:   general.MEXC,
			Instance: plugin,
			Ranking:  300,
		})
	}
}

type MEXCPlugin struct {
	baseInfoManager general.BaseInterface
	marketManager   general.MarketInterface
	accountManager  general.AccountInterface
	orderManager    general.OrderInterface
}

func NewMEXCPlugin() (*MEXCPlugin, error) {
	baseInfoManager, err := newBaseInfoManager()
	if err != nil {
		fmt.Printf("failed to init plugin [%s]\n", general.MEXC)
		return nil, err
	}
	marketManager := newMarketInfoManager()
	accountManager := newAccountManager()
	orderManager := NewOrderManager()
	return &MEXCPlugin{
		baseInfoManager: baseInfoManager,
		marketManager:   marketManager,
		accountManager:  accountManager,
		orderManager:    orderManager,
	}, nil
}

func (p *MEXCPlugin) ExchangeAlias() general.Exchange {
	return general.MEXC
}

func (p *MEXCPlugin) GetBaseInfoManager() general.BaseInterface {
	return p.baseInfoManager
}

func (p *MEXCPlugin) GetMarketInfoManager() general.MarketInterface {
	return p.marketManager
}

func (p *MEXCPlugin) GetAccountManager() general.AccountInterface {
	return p.accountManager
}

func (p *MEXCPlugin) GetOrderInterface() general.OrderInterface {
	return p.orderManager
}
