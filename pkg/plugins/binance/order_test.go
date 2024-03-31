package binance

import (
	"encoding/json"
	"fmt"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"testing"
)

func TestOrderManager_ListOrders(t *testing.T) {
	secret := GetSecretsForExchanger(Binance)
	manager := &OrderManager{
		secret: secret,
		client: getBinanceClient(secret),
	}
	var orders []*Order
	var b []byte
	//orders, _ = manager.ListOpenOrders()
	//b, _ = json.Marshal(orders)
	//fmt.Println("ListOpenOrders:", string(b))

	// 当前委托订单，必须指定币对才能看到
	orders, _ = manager.ListOpenOrdersOfSymbol(Symbol{BaseAsset: ETH, QuoteAsset: DefaultQuoteCoin})
	b, _ = json.Marshal(orders)
	fmt.Println("ListOpenOrdersOfSymbol:", string(b), orders)

	//allOrders, _ := manager.ListAllOrders(ETH)
	//b3, _ := json.Marshal(allOrders)
	//fmt.Println("ListAllOrders:", string(b3))

	//order, _ := manager.GetOrder(ETH, "", "web_f642937e597545cbbbe698c9d3316884")
	//b2, _ := json.Marshal(order)
	//fmt.Println(string(b2))

	status, err := manager.CancelOrder(Symbol{
		BaseAsset: ETH, QuoteAsset: DefaultQuoteCoin,
	}, "", "web_139696a33179453fa1b1904901afc419")
	fmt.Println(status, err)

	//ten := decimal.NewFromInt(10)
	//order, err := manager.CreateOrder(OrderPlan{
	//	Symbol:      ETH,
	//	Side:        SideTypeBuy,
	//	OrderType:   OrderTypeLimit,
	//	TimeInForce: TimeInForceTypeGTC,
	//	Price:       &ten,
	//	Quantity:    &ten,
	//})
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(order.OrderID, order.ClientOrderID)
}
