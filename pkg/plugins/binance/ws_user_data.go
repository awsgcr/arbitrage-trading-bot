package binance

import (
	"encoding/json"
	"fmt"

	. "jasonzhu.com/coin_labor/pkg/plugins/general"
)

// UserDataEventType define spot user data event type
type BUserDataEventType string

const (
	BUserDataEventTypeOutboundAccountPosition BUserDataEventType = "outboundAccountPosition"
	BUserDataEventTypeBalanceUpdate           BUserDataEventType = "balanceUpdate"
	BUserDataEventTypeExecutionReport         BUserDataEventType = "executionReport"
	BUserDataEventTypeListStatus              BUserDataEventType = "ListStatus"
)

// BWsUserDataEvent define user data event
type BWsUserDataEvent struct {
	Event             BUserDataEventType `json:"e"`
	Time              int64              `json:"E"`
	TransactionTime   int64              `json:"T"`
	AccountUpdateTime int64              `json:"u"`
	AccountUpdate     BWsAccountUpdateList
	BalanceUpdate     BWsBalanceUpdate
	OrderUpdate       BWsOrderUpdate
	OCOUpdate         BWsOCOUpdate
}

type BWsAccountUpdateList struct {
	WsAccountUpdates []BWsAccountUpdate `json:"B"`
}

// BWsAccountUpdate define account update
type BWsAccountUpdate struct {
	Asset  string `json:"a"`
	Free   string `json:"f"`
	Locked string `json:"l"`
}

type BWsBalanceUpdate struct {
	Asset  string `json:"a"`
	Change string `json:"d"`
}

type BWsOrderUpdate struct {
	Symbol                  string          `json:"s"`
	ClientOrderId           string          `json:"c"`
	Side                    string          `json:"S"`
	Type                    string          `json:"o"`
	TimeInForce             TimeInForceType `json:"f"`
	Volume                  string          `json:"q"`
	Price                   string          `json:"p"`
	StopPrice               string          `json:"P"`
	TrailingDelta           int64           `json:"d"` // Trailing Delta
	IceBergVolume           string          `json:"F"`
	OrderListId             int64           `json:"g"` // for OCO
	OrigCustomOrderId       string          `json:"C"` // customized order ID for the original order
	ExecutionType           string          `json:"x"` // execution type for this event NEW/TRADE...
	Status                  string          `json:"X"` // order status
	RejectReason            string          `json:"r"`
	Id                      int64           `json:"i"` // order id
	LatestVolume            string          `json:"l"` // quantity for the latest trade
	FilledVolume            string          `json:"z"`
	LatestPrice             string          `json:"L"` // price for the latest trade
	FeeAsset                string          `json:"N"`
	FeeCost                 string          `json:"n"`
	TransactionTime         int64           `json:"T"`
	TradeId                 int64           `json:"t"`
	IsInOrderBook           bool            `json:"w"` // is the order in the order book?
	IsMaker                 bool            `json:"m"` // is this order maker?
	CreateTime              int64           `json:"O"`
	FilledQuoteVolume       string          `json:"Z"` // the quote volume that already filled
	LatestQuoteVolume       string          `json:"Y"` // the quote volume for the latest trade
	QuoteVolume             string          `json:"Q"`
	TrailingTime            int64           `json:"D"` // Trailing Time
	StrategyId              int64           `json:"j"` // Strategy ID
	StrategyType            int64           `json:"J"` // Strategy Type
	WorkingTime             int64           `json:"W"` // Working Time
	SelfTradePreventionMode string          `json:"V"`
}

type BWsOCOUpdate struct {
	Symbol          string `json:"s"`
	OrderListId     int64  `json:"g"`
	ContingencyType string `json:"c"`
	ListStatusType  string `json:"l"`
	ListOrderStatus string `json:"L"`
	RejectReason    string `json:"r"`
	ClientOrderId   string `json:"C"` // List Client Order ID
	Orders          BWsOCOOrderList
}

type BWsOCOOrderList struct {
	WsOCOOrders []BWsOCOOrder `json:"O"`
}

type BWsOCOOrder struct {
	Symbol        string `json:"s"`
	OrderId       int64  `json:"i"`
	ClientOrderId string `json:"c"`
}

// WsUserDataHandler handle WsUserDataEvent
type WsUserDataHandler func(event *BWsUserDataEvent)

// WsUserDataServe serve user data handler with listen key
func WsUserDataServe(listenKey string, handler WsUserDataHandler, errHandler ErrHandler) (*WsServe, error) {
	endpoint := fmt.Sprintf("%s/%s", getWsEndpoint(), listenKey)
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}

		event := new(BWsUserDataEvent)

		err = json.Unmarshal(message, event)
		if err != nil {
			errHandler(err)
			return
		}

		switch BUserDataEventType(j.Get("e").MustString()) {
		case BUserDataEventTypeOutboundAccountPosition:
			err = json.Unmarshal(message, &event.AccountUpdate)
			if err != nil {
				errHandler(err)
				return
			}
		case BUserDataEventTypeBalanceUpdate:
			err = json.Unmarshal(message, &event.BalanceUpdate)
			if err != nil {
				errHandler(err)
				return
			}
		case BUserDataEventTypeExecutionReport:
			err = json.Unmarshal(message, &event.OrderUpdate)
			if err != nil {
				errHandler(err)
				return
			}
			// Unmarshal has case sensitive problem
			event.TransactionTime = j.Get("T").MustInt64()
			event.OrderUpdate.TransactionTime = j.Get("T").MustInt64()
			event.OrderUpdate.Id = j.Get("i").MustInt64()
			event.OrderUpdate.TradeId = j.Get("t").MustInt64()
			event.OrderUpdate.FeeAsset = j.Get("N").MustString()
		case BUserDataEventTypeListStatus:
			err = json.Unmarshal(message, &event.OCOUpdate)
			if err != nil {
				errHandler(err)
				return
			}
		}

		handler(event)
	}
	return NewWsServe(endpoint, wsHandler)
}
