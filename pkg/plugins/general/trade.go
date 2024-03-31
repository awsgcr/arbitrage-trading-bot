package general

import "github.com/shopspring/decimal"

type Trade struct {
	Symbol          string          `json:"symbol"`
	Id              string          `json:"id"`
	OrderId         string          `json:"orderId"`
	OrderListId     int64           `json:"orderListId"`
	Price           decimal.Decimal `json:"price"`
	Qty             decimal.Decimal `json:"qty"`
	QuoteQty        decimal.Decimal `json:"quoteQty"`
	Commission      string          `json:"commission"`
	CommissionAsset string          `json:"commissionAsset"`
	Time            int64           `json:"time"`
	IsBuyer         bool            `json:"isBuyer"`
	IsMaker         bool            `json:"isMaker"`
	IsBestMatch     bool            `json:"isBestMatch"`
	ClientOrderId   string          `json:"clientOrderId"`
}
