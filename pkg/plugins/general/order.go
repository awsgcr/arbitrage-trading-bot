package general

// OrderStatusType define order status type
type OrderStatusType string

// TimeInForceType define time in force type of order
type TimeInForceType string

// OrderType define order type
type OrderType string

// SideType define side type of order
type SideType string

// Global enums
const (
	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"

	OrderTypeLimit           OrderType = "LIMIT"
	OrderTypeMarket          OrderType = "MARKET"
	OrderTypeLimitMaker      OrderType = "LIMIT_MAKER"
	OrderTypeStopLoss        OrderType = "STOP_LOSS"
	OrderTypeStopLossLimit   OrderType = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit      OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitLimit OrderType = "TAKE_PROFIT_LIMIT"

	OrderStatusTypePreNew    OrderStatusType = "PRE_NEW"
	OrderStatusTypePreCancel OrderStatusType = "PRE_CANCEL"

	OrderStatusTypeNew             OrderStatusType = "NEW"
	OrderStatusTypePartiallyFilled OrderStatusType = "PARTIALLY_FILLED"
	OrderStatusTypeFilled          OrderStatusType = "FILLED"
	OrderStatusTypeCanceled        OrderStatusType = "CANCELED"
	OrderStatusTypePendingCancel   OrderStatusType = "PENDING_CANCEL"
	OrderStatusTypeRejected        OrderStatusType = "REJECTED"
	OrderStatusTypeExpired         OrderStatusType = "EXPIRED"

	// TimeInForceTypeGTC
	//GTC (Good-Till-Cancel): the order will last until it is completed or you cancel it.
	//IOC (Immediate-Or-Cancel): the order will attempt to execute all or part of it immediately at the price and quantity available, then cancel any remaining, unfilled part of the order. If no quantity is available at the chosen price when you place the order, it will be canceled immediately. Please note that Iceberg orders are not supported.
	//FOK (Fill-Or-Kill): the order is instructed to execute in full immediately (filled), otherwise it will be canceled (killed). Please note that Iceberg orders are not supported.
	//GTC (有效直到取消)：訂單將維持有效到成交或被您取消。
	//IOC (立即成交或取消)：以可用價格及數量立即嘗試成交全部或部分訂單，然後取消剩餘未成交的訂單部分。如果您下單時所選擇的價格沒有可供應的數量，訂單將會立即被取消。請注意，此訂單類型不支持冰山委託。
	//FOK (全部成交或取消)：訂單必須立即完全成交 (全部成交)，否則將被取消 (完全取消)。請注意，此訂單類型不支持冰山委託。
	TimeInForceTypeGTC TimeInForceType = "GTC"
	TimeInForceTypeIOC TimeInForceType = "IOC"
	TimeInForceTypeFOK TimeInForceType = "FOK"
)
