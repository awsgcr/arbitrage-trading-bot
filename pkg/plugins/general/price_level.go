package general

import (
	"github.com/shopspring/decimal"
)

type PriceLevel struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
}

var Nil = PriceLevel{}

func NewPriceLevelFromString(price string, quantity string) (PriceLevel, error) {
	var (
		p   decimal.Decimal
		q   decimal.Decimal
		err error
	)
	p, err = decimal.NewFromString(price)
	if err != nil {
		return Nil, err
	}
	q, err = decimal.NewFromString(quantity)
	if err != nil {
		return Nil, err
	}
	return PriceLevel{Price: p, Quantity: q}, nil
}

func (p *PriceLevel) NewPriceLevel(price decimal.Decimal, quantity decimal.Decimal) *PriceLevel {
	p.Price = price
	p.Quantity = quantity
	return p
}

// Ask is a type alias for PriceLevel.
type Ask = PriceLevel

// Bid is a type alias for PriceLevel.
type Bid = PriceLevel
