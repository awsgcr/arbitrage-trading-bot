package mexc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"jasonzhu.com/coin_labor/core/setting"
	"jasonzhu.com/coin_labor/core/util"
	"jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"testing"
	"time"
)

var symbol = Symbol{
	BaseAsset:  OG,
	QuoteAsset: DefaultQuoteCoin,
}

func TestMarket(t *testing.T) {
	// BaseInfo
	infoManager, _ := newBaseInfoManager()
	printObj(infoManager.ServerTime())

	symbolsBasicInfo := infoManager.GetSymbolsBasicInfo()
	printObj(symbolsBasicInfo)
	basicInfo, err := infoManager.GetSymbolBasicInfo(Symbol{BaseAsset: OG, QuoteAsset: DefaultQuoteCoin})
	printOrFatal(t, err, basicInfo)

	// Market
	manager := newMarketInfoManager()
	depth, err := manager.FetchDepth(symbol, 10)
	if err != nil {
		t.Fatal(err)
	}
	printObj(depth)

	// Account Balances
	secret := GetSecretsForExchanger(MEXC)
	accountManager := &AccountManager{
		lg:     plg.New("s", "account"),
		secret: secret,
		client: http.NewHMACClient(secret, baseAPIMainURL, apiKeyHeader),
	}
	info, err := accountManager.GetAccountInfo()
	if err != nil {
		t.Fatal(err)
	}
	printObj(info)

	// Order
	orderManager := NewOrderManager()
	//orders, err := orderManager.ListAllOrders(symbol)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//printObj("all Orders", orders)

	limitOrder := NewLimitOrder(symbol, SideTypeBuy,
		TimeInForceTypeGTC,
		decimal.NewFromFloat(3.001), // 精度3， 对应 quoteAssetPrecision
		decimal.NewFromFloat(3.01),  // 精度2， 对应 baseAssetPrecision
	)
	order, err := orderManager.CreateOrder(*limitOrder)
	printOrFatal(t, err, "create order", order)
	res, err := orderManager.ListOpenOrdersOfSymbol(symbol)
	printOrFatal(t, err, "ListOpenOrdersOfSymbol", res)
	getOrder, err := orderManager.GetOrder(symbol, order.OrderID, order.ClientOrderID)
	printOrFatal(t, err, "GetOrder", getOrder)
	cancelOrder, err := orderManager.CancelOrder(symbol, order.OrderID, order.ClientOrderID)
	time.Sleep(1 * time.Second)
	printOrFatal(t, err, "cancelOrder", cancelOrder)
	openOrders, err := orderManager.ListOpenOrdersOfSymbol(symbol)
	printOrFatal(t, err, "ListOpenOrdersOfSymbol", openOrders)
	getOrder2, err := orderManager.GetOrder(symbol, order.OrderID, order.ClientOrderID)
	printOrFatal(t, err, "GetOrder", getOrder2)

}

// Market Order Testing.
func TestOrderManager_CreateOrder(t *testing.T) {
	orderManager := NewOrderManager()
	//order := NewMarketOrder(NewSymbol(INJ), SideTypeSell, decimal.NewFromInt(1), decimal.NewFromFloat(0.7)) // Support Market
	order := NewMarketOrder(NewSymbol(WOO), SideTypeSell, decimal.NewFromFloat(0.2732), decimal.NewFromFloat(30)) // Not support Market
	createOrder, err := orderManager.CreateOrder(*order)
	fmt.Println(createOrder, err)
}

// ------------------------------------------ WebSocket ------------------------------------------------------
func TestAccountManager_WsWatchUserDataChanges(t *testing.T) {
	setting.AlertingEnabled = true
	// Account UserDataStream.
	var err error
	secret := GetSecretsForExchanger(MEXC)
	accountManager := &AccountManager{
		lg:     plg.New("s", "account"),
		secret: secret,
		client: http.NewHMACClient(secret, baseAPIMainURL, apiKeyHeader),
	}
	err = accountManager.createListenKey()
	if err != nil {
		t.Fatal(err)
	}
	printObj(accountManager.userStreamListenKey)

	userDataEventC := make(chan *UserDataEvent)
	watchUserDataEventC(userDataEventC)
	err = accountManager.WsWatchUserDataChanges(context.Background(), userDataEventC)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("end")
}

func watchUserDataEventC(c chan *UserDataEvent) {
	go func() {
		for e := range c {
			printObj(e)
		}
	}()
}

func printOrFatal(t *testing.T, err error, a ...any) {
	if err != nil {
		t.Fatal(err)
	}
	printObj(a...)
}

func printObj(o ...any) {
	if len(o) > 1 {
		b, _ := json.Marshal(o[1])
		fmt.Println(o[0], string(b))
		return
	}
	b, err := json.Marshal(o)
	if err != nil {
		fmt.Println("json error", err)
	}
	fmt.Println(string(b))
}

var assets = []string{
	//"LTC", "INJ", "WOO",
	"HBAR", "BICO", "CTK", "MKR", "KSM", "NEO", "BAR", "AVA", "KAVA", "AVA", "GMX", "ONE", "OGN", "RNDR", "ETC", "IOST", "QNT", "ENS", "WOO", "SSV", "ZRX", "CELR", "IOTX",
}
var dot = ","

func TestBinancePlugin_GetMarketInfoManager(t *testing.T) {
	manager := newMarketInfoManager()

	for _, asset := range assets {
		depth, err := manager.FetchDepth(NewSymbol(Asset(asset)), 5)
		if err != nil {
			fmt.Println(fillWithBlank(asset, 6), dot, err)
			continue
		}
		ask, err := depth.TopAsk()
		if err != nil {
			fmt.Println(fillWithBlank(asset, 6), dot, err)
			continue
		}
		fmt.Println(fillWithBlank(asset, 6), dot, dot, ask.Price, dot, depth.Asks[1].Price, dot, depth.Asks[2].Price)
		time.Sleep(100 * time.Millisecond)
	}
}

func fillWithBlank(s string, size int) string {
	if len(s) > size {
		return s[0:size]
	} else {
		num := size - len(s)
		blank := ""
		for i := 0; i < num; i++ {
			blank += " "
		}
		return s + blank
	}
}

func TestTradeManager_ListTradesOfSymbol(t *testing.T) {
	manager := NewTradeManager()
	res, err := manager.ListTradesOfSymbol(NewSymbol(INJ), 500)
	if err != nil {
		t.Fatal(err)
	}

	var cnt = 0
	var total = decimal.Zero
	start, _ := util.GetMsOfDateTime("2023-04-27 14:54:00")
	end, _ := util.GetMsOfDateTime("2023-04-27 18:18:18")
	for _, trade := range res {
		day := util.UnixMillToStr(trade.Time)
		if trade.Time > start && trade.Time < end {
			fmt.Println(day, trade.Symbol, trade.Price, trade.Qty, trade.QuoteQty)
			cnt++
			total = total.Add(trade.QuoteQty)
		}
	}
	fmt.Println((end-start)/1000/60, cnt, total)
}
