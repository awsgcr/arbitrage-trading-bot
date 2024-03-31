package mexc

import (
	"context"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/components/metrics"
	"jasonzhu.com/coin_labor/core/setting"
	. "jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"net/http"
	"time"
)

type OrderManager struct {
	secret *setting.Secret
	client *Client
	lg     log.Logger
}

func NewOrderManager() OrderInterface {
	secret := GetSecretsForExchanger(MEXC)
	return &OrderManager{
		secret: secret,
		lg:     plg.New("s", "order"),
		client: NewHMACClient(secret, baseAPIMainURL, apiKeyHeader),
	}
}

// ListOpenOrdersOfSymbol
/**
[
{}
]
*/
func (s *OrderManager) ListOpenOrdersOfSymbol(symbol Symbol) (res []*Order, err error) {
	start := time.Now()
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: openOrdersEndpoint,
		SecType:  SecTypeSigned,
	}
	r.SetParam("symbol", getSymbolAlias(symbol))
	data, err := s.client.CallAPI(context.Background(), r)
	defer func() { uploadMetrics(symbol.BaseAsset, "ListOpenOrdersOfSymbol", err, start) }()

	if err != nil {
		return nil, err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	size := len(j.MustArray())
	var orders = make([]*Order, size)
	for i := 0; i < size; i++ {
		item := j.GetIndex(i)
		orders[i] = convertToOrder(item)
	}
	return orders, nil
}

/**
  {
    "symbol": "LTCBTC",
    "orderId": 1,
    "orderListId": -1,
    "clientOrderId": "myOrder1",
    "price": "0.1",
    "origQty": "1.0",
    "executedQty": "0.0",
    "cummulativeQuoteQty": "0.0",
    "status": "NEW",
    "timeInForce": "GTC",
    "type": "LIMIT",
    "side": "BUY",
    "stopPrice": "0.0",
    "icebergQty": "0.0",
    "time": 1499827319559,
    "updateTime": 1499827319559,
    "isWorking": true,
    "origQuoteOrderQty": "0.000000"
  }

{
  "symbol": "LTCBTC", // 交易对
  "orderId": 1, // 系统的订单ID
  "orderListId": -1, // OCO订单的ID，不然就是-1
  "clientOrderId": "myOrder1", // 客户自己设置的ID
  "price": "0.1", // 订单价格
  "origQty": "1.0", // 用户设置的原始订单数量
  "executedQty": "0.0", // 交易的订单数量
  "cummulativeQuoteQty": "0.0", // 累计交易的金额
  "status": "NEW", // 订单状态
  "timeInForce": "GTC", // 订单的时效方式
  "type": "LIMIT", // 订单类型， 比如市价单，现价单等
  "side": "BUY", // 订单方向，买还是卖
  "stopPrice": "0.0", // 止损价格
  "icebergQty": "0.0", // 冰山数量
  "time": 1499827319559, // 订单时间
  "updateTime": 1499827319559, // 最后更新时间
  "isWorking": true, // 订单是否出现在orderbook中
  "origQuoteOrderQty": "0.000000" // 原始的交易金额
}
*/
func convertToOrder(item *simplejson.Json) *Order {
	return &Order{
		Symbol:                   item.Get("symbol").MustString(),                                             // "LTCBTC",
		OrderID:                  item.Get("orderId").MustString(),                                            // 1,
		ClientOrderID:            item.Get("clientOrderId").MustString(),                                      // "t7921223K12",
		Price:                    NewDecimalFromStringIgnoreErr(item.Get("price").MustString()),               // "0.1",
		OrigQuantity:             NewDecimalFromStringIgnoreErr(item.Get("origQty").MustString()),             // "1.0",
		ExecutedQuantity:         NewDecimalFromStringIgnoreErr(item.Get("executedQty").MustString()),         // "0.0",
		CummulativeQuoteQuantity: NewDecimalFromStringIgnoreErr(item.Get("cummulativeQuoteQty").MustString()), // "0.0",
		Status:                   OrderStatusType(item.Get("status").MustString()),                            // "NEW",
		TimeInForce:              TimeInForceType(item.Get("timeInForce").MustString()),                       // "GTC",
		Type:                     OrderType(item.Get("type").MustString()),                                    // "LIMIT",
		Side:                     SideType(item.Get("side").MustString()),                                     // "BUY",
		StopPrice:                NewDecimalFromStringIgnoreErr(item.Get("stopPrice").MustString()),           // "0.0",
		IcebergQuantity:          NewDecimalFromStringIgnoreErr(item.Get("icebergQty").MustString()),          // "0.0",
		Time:                     item.Get("time").MustInt64(),                                                // 1499827319559,
		UpdateTime:               item.Get("updateTime").MustInt64(),                                          // 1499827319559,
		IsWorking:                item.Get("isWorking").MustBool(),                                            // true
	}
}

func (s *OrderManager) ListAllOrders(symbol Symbol) ([]*Order, error) {
	start := time.Now()
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: allOrdersEndpoint,
		SecType:  SecTypeSigned,
	}
	r.SetParam("symbol", getSymbolAlias(symbol))
	data, err := s.client.CallAPI(context.Background(), r)
	defer func() { uploadMetrics(symbol.BaseAsset, "ListAllOrders", err, start) }()

	if err != nil {
		return nil, err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	size := len(j.MustArray())
	var orders = make([]*Order, size)
	for i := 0; i < size; i++ {
		item := j.GetIndex(i)
		orders[i] = convertToOrder(item)
	}
	return orders, nil
}

// CreateOrder
/**
symbol	string	是	交易对
side	ENUM	是	详见枚举定义：订单方向
type	ENUM	是	详见枚举定义：订单类型
quantity	decimal	否	委托数量
quoteOrderQty	decimal	否	委托总额
price	decimal	否	委托价格
newClientOrderId	string	否	客户自定义的唯一订单ID
recvWindow	long	否	赋值不能大于 60000
timestamp	long	是
*/
func (s *OrderManager) CreateOrder(plan OrderPlan) (*CreateOrderResponse, error) {
	//fmt.Println("try to create order", plan.ToString())
	//return nil, errors.New("NOT NOW")
	if !DefaultHealthChecker.IsAllFeaturesHealthy() {
		plg.Warn("unhealthy, skip create order in MEXC")
		return nil, errors.New("unhealthy right now, unable to create order")
	}

	start := time.Now()
	s.lg.Warn("createOrder start", "orderPlan", plan.ToString())
	endpoint := orderEndpoint
	r := &Request{
		Method:   http.MethodPost,
		Endpoint: endpoint,
		SecType:  SecTypeSigned,
	}
	symbol2USDT := getSymbolAlias(plan.Symbol)
	r.SetParam("symbol", symbol2USDT)
	r.SetParam("side", plan.Side)      //ENUM: BUY SELL
	r.SetParam("type", plan.OrderType) // ENUM: same as Binance, some of what is not supported right now. https://www.MEXC.me/docs/v1/intro#enum-definitions

	if plan.ClientOrderID != "" {
		r.SetParam("newClientOrderId", plan.ClientOrderID)
	}

	if plan.Quantity != nil {
		qty := plan.Quantity.String()
		r.SetParam("quantity", qty)
	} else {
		return nil, errors.New("quantity can't be null")
	}

	if plan.OrderType == OrderTypeLimit {
		if plan.Price != nil {
			r.SetParam("price", plan.Price.String())
		} else {
			return nil, errors.New("price can't be null")
		}
	} else if plan.OrderType == OrderTypeMarket {
		//r.SetParam()
	} else {
		// not supported.
	}

	// plan.QuoteOrderQty //quoteOrderQty param is not supported in MEXC

	data, err := s.client.CallAPI(context.Background(), r)
	defer func() { uploadMetrics(plan.Symbol.BaseAsset, "CreateOrder", err, start) }()

	if err != nil {
		s.lg.Error("createOrder failed", "clientOrderId", plan.ClientOrderID, "err", err)
		fmt.Println("ERROR---------", err, r)
		//alerting.Notify(err, "Create Order Failed in MEXC", "ClientOrderID", plan.ClientOrderID)
		return nil, err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}

	orderId := j.Get("orderId").MustString()
	s.lg.Warn("createOrder succeed", "ClientOrderID", plan.ClientOrderID, "orderId", orderId)
	return &CreateOrderResponse{
		OrderID:       orderId,
		ClientOrderID: plan.ClientOrderID,
	}, nil
}

func (s *OrderManager) GetOrder(symbol Symbol, orderId string, clientOrderId string) (*Order, error) {
	start := time.Now()
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: orderEndpoint,
		SecType:  SecTypeSigned,
	}
	r.SetParam("symbol", getSymbolAlias(symbol))
	if orderId != "" {
		r.SetParam("orderId", orderId)
	}
	if clientOrderId != "" {
		r.SetParam("origClientOrderId", clientOrderId)
	}
	data, err := s.client.CallAPI(context.Background(), r)
	defer func() { uploadMetrics(symbol.BaseAsset, "GetOrder", err, start) }()

	if err != nil {
		return nil, err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return convertToOrder(j), nil
}

func (s *OrderManager) CancelOrder(symbol Symbol, orderId string, clientOrderId string) (OrderStatusType, error) {
	start := time.Now()
	r := &Request{
		Method:   http.MethodDelete,
		Endpoint: orderEndpoint,
		SecType:  SecTypeSigned,
	}
	r.SetParam("symbol", getSymbolAlias(symbol))
	if orderId != "" {
		r.SetParam("orderId", orderId)
	}
	if clientOrderId != "" {
		r.SetParam("origClientOrderId", clientOrderId)
	}
	data, err := s.client.CallAPI(context.Background(), r)
	defer func() { uploadMetrics(symbol.BaseAsset, "CancelOrder", err, start) }()

	if err != nil {
		s.lg.Debug("cancelOrder failed", "orderId", orderId, "clientOrderId", clientOrderId, "err", err)
		return "", err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return "", err
	}
	status := j.Get("status").MustString()
	s.lg.Debug("cancelOrder succeed", "orderId", orderId, "clientOrderId", clientOrderId, "status", status)
	return OrderStatusType(status), nil
}

func uploadMetrics(asset Asset, typ string, err error, start time.Time) {
	go func() {
		metrics.M_Coin_Order_Total.WithLabelValues(
			string(MEXC), string(asset), typ, IsErrNil(err),
		).Inc()
		duration := time.Since(start).Microseconds()
		metrics.M_Coin_Order_Executeion_Time_Summary.WithLabelValues(
			string(MEXC), string(asset), typ, IsErrNil(err),
		).Observe(float64(duration))
		metrics.M_Coin_Order_Executeion_Time_Histogram.WithLabelValues(
			string(MEXC), string(asset), typ, IsErrNil(err),
		).Observe(float64(duration))
	}()
}
