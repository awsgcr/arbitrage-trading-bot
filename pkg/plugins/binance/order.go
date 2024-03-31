package binance

import (
	"context"
	"errors"
	"github.com/adshao/go-binance/v2"
	"jasonzhu.com/coin_labor/core/components/alerting"
	"jasonzhu.com/coin_labor/core/components/metrics"
	"jasonzhu.com/coin_labor/core/setting"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"strconv"
	"time"
)

type OrderManager struct {
	secret *setting.Secret
	client *binance.Client
}

func NewOrderManager() OrderInterface {
	secret := GetSecretsForExchanger(Binance)
	return &OrderManager{
		secret: secret,
		client: getBinanceClient(secret),
	}
}

func (s *OrderManager) ListOpenOrders() (res []*Order, err error) {
	start := time.Now()
	openOrders, err := s.client.NewListOpenOrdersService().
		Do(context.Background())
	defer func() { uploadMetrics(UnKnown, "ListOpenOrders", err, start) }()

	if err != nil {
		return nil, err
	}
	var orders []*Order
	for _, o := range openOrders {
		orders = append(orders, convertToOrder(o))
	}
	return orders, nil
}

func (s *OrderManager) ListOpenOrdersOfSymbol(symbol Symbol) (res []*Order, err error) {
	start := time.Now()
	symbol2USDT := getSymbolAlias(symbol)
	openOrders, err := s.client.NewListOpenOrdersService().Symbol(symbol2USDT).
		Do(context.Background())
	defer func() { uploadMetrics(symbol.BaseAsset, "ListOpenOrdersOfSymbol", err, start) }()

	if err != nil {
		return nil, err
	}
	var orders []*Order
	for _, o := range openOrders {
		orders = append(orders, convertToOrder(o))
	}
	return orders, nil
}

func convertToOrder(o *binance.Order) *Order {
	return &Order{
		Symbol:                   o.Symbol,
		OrderID:                  strconv.FormatInt(o.OrderID, 10),
		OrderListId:              o.OrderListId,
		ClientOrderID:            o.ClientOrderID,
		Price:                    NewDecimalFromStringIgnoreErr(o.Price),
		OrigQuantity:             NewDecimalFromStringIgnoreErr(o.OrigQuantity),
		ExecutedQuantity:         NewDecimalFromStringIgnoreErr(o.ExecutedQuantity),
		CummulativeQuoteQuantity: NewDecimalFromStringIgnoreErr(o.CummulativeQuoteQuantity),
		Status:                   OrderStatusType(o.Status),
		TimeInForce:              TimeInForceType(o.TimeInForce),
		Type:                     OrderType(o.Type),
		Side:                     SideType(o.Side),
		StopPrice:                NewDecimalFromStringIgnoreErr(o.StopPrice),
		IcebergQuantity:          NewDecimalFromStringIgnoreErr(o.IcebergQuantity),
		Time:                     o.Time,
		UpdateTime:               o.UpdateTime,
		IsWorking:                o.IsWorking,
		IsIsolated:               o.IsIsolated,
		OrigQuoteOrderQuantity:   NewDecimalFromStringIgnoreErr(o.OrigQuoteOrderQuantity),
	}
}

func (s *OrderManager) ListAllOrders(symbol Symbol) ([]*Order, error) {
	start := time.Now()
	symbol2USDT := getSymbolAlias(symbol)
	res, err := s.client.NewListOrdersService().Symbol(symbol2USDT).
		Do(context.Background())

	defer func() { uploadMetrics(symbol.BaseAsset, "ListAll", err, start) }()

	if err != nil {
		return nil, err
	}
	var orders []*Order
	for _, o := range res {
		orders = append(orders, convertToOrder(o))
	}
	return orders, nil
}

// CreateOrder Create order
func (s *OrderManager) CreateOrder(plan OrderPlan) (*CreateOrderResponse, error) {
	//if true {
	//	fmt.Println("try to create order", plan.ToString())
	//	return nil, errors.New("NOT NOW")
	//}
	if !DefaultHealthChecker.IsAllFeaturesHealthy() {
		plg.Warn("unhealthy, skip create order in Binance")
		return nil, errors.New("unhealthy right now, unable to create order")
	}

	start := time.Now()
	plg.Warn("Create Order Start", "plan", plan.ToString())
	symbol2USDT := getSymbolAlias(plan.Symbol)
	service := s.client.NewCreateOrderService().Symbol(symbol2USDT).
		Side(binance.SideType(plan.Side)).
		Type(binance.OrderType(plan.OrderType))

	if plan.ClientOrderID != "" {
		service.NewClientOrderID(plan.ClientOrderID)
	}
	if plan.OrderType == OrderTypeLimit {
		service.
			TimeInForce(binance.TimeInForceType(plan.TimeInForce)).
			Quantity(plan.Quantity.String()).
			Price(plan.Price.String())
	} else if plan.OrderType == OrderTypeMarket {
		//service.Price(plan.Price.String()) Error Code: -1106 Parameter 'price' sent when not required.
		if plan.Quantity != nil {
			service.Quantity(plan.Quantity.String())
		} else if plan.QuoteOrderQty != nil {
			service.QuoteOrderQty(plan.QuoteOrderQty.String())
		}
	} else {
		return nil, errors.New("not supported orderType")
	}
	order, err := service.Do(context.Background())

	defer func() { uploadMetrics(plan.Symbol.BaseAsset, "CreateOrder", err, start) }()

	if err != nil {
		plg.Error("Create Order Failed", "ClientOrderID", plan.ClientOrderID, "err", err)
		alerting.Notify(err, "Create Order Failed in binance", "ClientOrderID", plan.ClientOrderID)
		return nil, err
	}
	plg.Warn("Create Order Succeed", "ClientOrderID", plan.ClientOrderID, "OrderID", order.OrderID)
	//alerting.Info("Create Order Succeed in Binance", "type", plan.OrderType, "clientOrderID", plan.ClientOrderID)
	return NewCreateOrderResponse(order.OrderID, order.ClientOrderID), nil
}

func (s *OrderManager) GetOrder(symbol Symbol, orderId string, clientOrderId string) (*Order, error) {
	start := time.Now()
	symbol2USDT := getSymbolAlias(symbol)
	service := s.client.NewGetOrderService().Symbol(symbol2USDT)
	if clientOrderId != "" {
		service.OrigClientOrderID(clientOrderId)
	}
	if orderId != "" {
		oId, err := strconv.ParseInt(orderId, 10, 64)
		if err == nil {
			service.OrderID(oId)
		}
	}
	order, err := service.Do(context.Background())
	defer func() { uploadMetrics(symbol.BaseAsset, "GetOrder", err, start) }()

	if err != nil {
		return nil, err
	}
	return convertToOrder(order), nil
}

func (s *OrderManager) CancelOrder(symbol Symbol, orderId string, clientOrderId string) (OrderStatusType, error) {
	start := time.Now()
	symbol2USDT := getSymbolAlias(symbol)
	service := s.client.NewCancelOrderService().Symbol(symbol2USDT)
	if clientOrderId != "" {
		service.OrigClientOrderID(clientOrderId)
	}
	if orderId != "" {
		oId, err := strconv.ParseInt(orderId, 10, 64)
		if err == nil {
			service.OrderID(oId)
		}
	}
	res, err := service.Do(context.Background())
	defer func() { uploadMetrics(symbol.BaseAsset, "CancelOrder", err, start) }()

	if err != nil {
		return "", err
	}
	return OrderStatusType(res.Status), nil
}

func uploadMetrics(asset Asset, typ string, err error, start time.Time) {
	go func() {
		metrics.M_Coin_Order_Total.WithLabelValues(
			string(Binance), string(asset), typ, IsErrNil(err),
		).Inc()
		duration := time.Since(start).Microseconds()
		metrics.M_Coin_Order_Executeion_Time_Summary.WithLabelValues(
			string(Binance), string(asset), typ, IsErrNil(err),
		).Observe(float64(duration))
		metrics.M_Coin_Order_Executeion_Time_Histogram.WithLabelValues(
			string(Binance), string(asset), typ, IsErrNil(err),
		).Observe(float64(duration))
	}()
}
