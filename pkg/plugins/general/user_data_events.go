package general

import "github.com/shopspring/decimal"

// UserDataEventType define spot user data event type
type UserDataEventType string

const (
	UserDataEventTypeOutboundAccountPosition UserDataEventType = "outboundAccountPosition"
	UserDataEventTypeBalanceUpdate           UserDataEventType = "balanceUpdate"
	UserDataEventTypeExecutionReport         UserDataEventType = "executionReport"
	UserDataEventTypeListStatus              UserDataEventType = "ListStatus"
)

// UserDataEvent define user data event
type UserDataEvent struct {
	Event             UserDataEventType `json:"e"`
	Time              uint64            `json:"E"`
	TransactionTime   int64             `json:"T"`
	AccountUpdateTime int64             `json:"u"`
	AccountUpdate     WsAccountUpdateList
	BalanceUpdate     WsBalanceUpdate
	OrderUpdate       WsOrderUpdate
	OCOUpdate         WsOCOUpdate
}

type WsAccountUpdateList struct {
	WsAccountUpdates []WsAccountUpdate `json:"B"`
}

// WsAccountUpdate define account update
type WsAccountUpdate struct {
	Asset  Asset           `json:"a"`
	Free   decimal.Decimal `json:"f"`
	Locked decimal.Decimal `json:"l"`
}

type WsBalanceUpdate struct {
	Asset  Asset           `json:"a"`
	Change decimal.Decimal `json:"d"` //Balance Delta 余额增量
}

type WsOrderUpdate struct {
	Symbol        Symbol          `json:"s"` //  "s": "ETHBTC",                 // 交易对
	ClientOrderId string          `json:"c"` //  "c": "mUvoqJxFIILMdfAW5iGSOW", // clientOrderId
	Side          SideType        `json:"S"` //  "S": "BUY",                    // 订单方向
	Type          OrderType       `json:"o"` //  "o": "LIMIT",                  // 订单类型
	TimeInForce   TimeInForceType `json:"f"` //  "f": "GTC",                    // 有效方式
	Volume        decimal.Decimal `json:"q"` //  "q": "1.00000000",             // 订单原始数量
	Price         decimal.Decimal `json:"p"` //  "p": "0.10264410",             // 订单原始价格
	//StopPrice               decimal.Decimal `json:"P"` //  "P": "0.00000000",             // 止盈止损单触发价格
	//TrailingDelta           int64           `json:"d"` // Trailing Delta
	//IceBergVolume           string          `json:"F"` //  "F": "0.00000000",             // 冰山订单数量
	//OrderListId             int64           `json:"g"` //  "g": -1,                       // OCO订单 OrderListId                               // for OCO
	//OrigCustomOrderId       string          `json:"C"` //  "C": "",                       // 原始订单自定义ID(原始订单，指撤单操作的对象。撤单本身被视为另一个订单)     // customized order ID for the original order
	//ExecutionType           string          `json:"x"` //  "x": "NEW",                    // 本次事件的具体执行类型                                   // execution type for this event NEW/TRADE...
	Status       OrderStatusType `json:"X"` //  "X": "NEW",                    // 订单的当前状态                                              // order status
	RejectReason string          `json:"r"` //  "r": "NONE",                   // 订单被拒绝的原因
	Id           int64           `json:"i"` //  "i": 4293153,                  // orderId                                                  // order id
	//LatestVolume            string          `json:"l"` //  "l": "0.00000000",             // 订单末次成交量                                        // quantity for the latest trade
	FilledVolume decimal.Decimal `json:"z"` //  "z": "0.00000000",             // 订单累计已成交量
	LatestPrice  decimal.Decimal `json:"L"` //  "L": "0.00000000",             // 订单末次成交价格                                        // price for the latest trade
	//FeeAsset                string          `json:"N"` //  "N": null,                     // 手续费资产类别
	//FeeCost                 decimal.Decimal `json:"n"` //  "n": "0",                      // 手续费数量
	TransactionTime int64 `json:"T"` //  "T": 1499405658657,            // 成交时间
	//TradeId           int64  `json:"t"`
	//IsInOrderBook     bool   `json:"w"` //  "w": true,                     // 订单是否在订单簿上？                                    // is the order in the order book?
	IsMaker           bool            `json:"m"` //  "m": false,                    // 该成交是作为挂单成交吗？                                        // is this order maker?
	CreateTime        int64           `json:"O"` //  "O": 1499405658657,            // 订单创建时间
	FilledQuoteVolume decimal.Decimal `json:"Z"` //  "Z": "0.00000000",             // 订单累计已成交金额                                 // the quote volume that already filled
	//LatestQuoteVolume string `json:"Y"` //  "Y": "0.00000000",             // 订单末次成交金额                                  // the quote volume for the latest trade
	//QuoteVolume string `json:"Q"` //  "Q": "0.00000000",             // Quote Order Quantity
	//TrailingTime            int64  `json:"D"` //  "D": 1668680518494,            // 追踪时间; 这仅在追踪止损订单已被激活时可见                         // Trailing Time
	//StrategyId              int64  `json:"j"` // Strategy ID
	//StrategyType            int64  `json:"J"` // Strategy Type
	//WorkingTime int64 `json:"W"` //  "W": 1499405658657,            // Working Time; 订单被添加到 order book 的时间//
	//SelfTradePreventionMode string `json:"V"` //  "V": "NONE"                    // SelfTradePreventionMode//
}

type WsOCOUpdate struct {
	Symbol          string `json:"s"`
	OrderListId     int64  `json:"g"`
	ContingencyType string `json:"c"`
	ListStatusType  string `json:"l"`
	ListOrderStatus string `json:"L"`
	RejectReason    string `json:"r"`
	ClientOrderId   string `json:"C"` // List Client Order ID
	Orders          WsOCOOrderList
}

type WsOCOOrderList struct {
	WsOCOOrders []WsOCOOrder `json:"O"`
}

type WsOCOOrder struct {
	Symbol        string `json:"s"`
	OrderId       int64  `json:"i"`
	ClientOrderId string `json:"c"`
}
