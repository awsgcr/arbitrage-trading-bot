package general

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"strings"
)

// OrderPlan https://github.com/binance/binance-spot-api-docs/blob/master/web-socket-api_CN.md
type OrderPlan struct {
	Symbol        Symbol
	Side          SideType
	ClientOrderID string

	// LIMIT: timeInForce, price, quantity
	// MARKET: quantity 或者 quoteOrderQty
	OrderType     OrderType
	TimeInForce   TimeInForceType
	Price         *decimal.Decimal
	Quantity      *decimal.Decimal
	QuoteOrderQty *decimal.Decimal

	Res *CreateOrderResponse
}

func NewLimitOrder(symbol Symbol, side SideType, timeInForce TimeInForceType, price decimal.Decimal, quantity decimal.Decimal) *OrderPlan {
	o := &OrderPlan{
		Symbol:        symbol,
		Side:          side,
		OrderType:     OrderTypeLimit,
		ClientOrderID: genClientOrderID(),
		TimeInForce:   timeInForce,
		Price:         &price,
		Quantity:      &quantity,
	}
	glg.Debug("new limit order", "d", o.ToString())
	return o
}

func genClientOrderID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func NewMarketOrder(symbol Symbol, side SideType, price, quantity decimal.Decimal) *OrderPlan {
	o := &OrderPlan{
		Symbol:        symbol,
		Side:          side,
		OrderType:     OrderTypeMarket,
		ClientOrderID: genClientOrderID(),
		Price:         &price,
		Quantity:      &quantity,
	}
	glg.Debug("new market order", "d", o.ToString())
	return o
}

// NewMarketOrderWithQuoteQty 使用 quoteOrderQty 的 MARKET 订单 明确的是通过买入(或卖出)想要花费(或获取)的 quote asset 数量。 基础资产的实际执行数量将取决于可用的市场流动性。
func NewMarketOrderWithQuoteQty(symbol Symbol, side SideType, price, quoteOrderQty decimal.Decimal) *OrderPlan {
	o := &OrderPlan{
		Symbol:        symbol,
		Side:          side,
		OrderType:     OrderTypeMarket,
		ClientOrderID: genClientOrderID(),
		Price:         &price,
		QuoteOrderQty: &quoteOrderQty,
	}
	glg.Debug("new market order", "d", o.ToString())
	return o
}

func (o *OrderPlan) Amount() decimal.Decimal {
	return o.Price.Mul(*o.Quantity)
}

func (o *OrderPlan) SetCreateOrderResponse(response *CreateOrderResponse) {
	o.Res = response
}

func (o *OrderPlan) ToString() string {
	return fmt.Sprintf("Symbol: %s, ClientOrderID: %s, type: %s, side: %s, price: %s, quantity: %s, amount: %s", o.Symbol.BaseAsset, o.ClientOrderID, o.OrderType, o.Side, o.Price, o.Quantity, o.Amount())
}
