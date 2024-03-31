package general

import (
	"github.com/shopspring/decimal"
	"strconv"
)

// Order define order info
type Order struct {
	Symbol                   string
	OrderID                  string
	OrderListId              int64
	ClientOrderID            string
	Price                    decimal.Decimal
	OrigQuantity             decimal.Decimal
	ExecutedQuantity         decimal.Decimal
	CummulativeQuoteQuantity decimal.Decimal
	Status                   OrderStatusType
	TimeInForce              TimeInForceType
	Type                     OrderType
	Side                     SideType
	StopPrice                decimal.Decimal
	IcebergQuantity          decimal.Decimal
	Time                     int64
	UpdateTime               int64
	IsWorking                bool
	IsIsolated               bool
	OrigQuoteOrderQuantity   decimal.Decimal
}

type CreateOrderResponse struct {
	OrderID       string
	ClientOrderID string
}

func NewCreateOrderResponse(orderID int64, ClientOrderID string) *CreateOrderResponse {
	return &CreateOrderResponse{
		OrderID:       strconv.FormatInt(orderID, 10),
		ClientOrderID: ClientOrderID,
	}
}
