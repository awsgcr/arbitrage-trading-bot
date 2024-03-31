package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"jasonzhu.com/coin_labor/core/components/log"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"testing"
)

func TestAccountManager_ListOrders(t *testing.T) {
	binance.UseTestnet = false
	secret := GetSecretsForExchanger(Binance)
	manager := &AccountManager{
		lg:     log.New("binance.account_manager"),
		secret: secret,
		client: getBinanceClient(secret),
	}
	info, _ := manager.GetAccountInfo()
	fmt.Println(info)

	eventC := make(chan *UserDataEvent)

	go func() {
		for {
			select {
			case e := <-eventC:
				bytes, _ := json.Marshal(e)
				fmt.Println("3333", string(bytes))
			}
		}
	}()

	manager.WsWatchUserDataChanges(context.Background(), eventC)
}
