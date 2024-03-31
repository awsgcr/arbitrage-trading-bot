package general

import (
	"strconv"
)

type PriceLevel struct {
	Price    float64
	Quantity float64
}

var Nil = PriceLevel{}

func NewFromString(price string, quantity string) (PriceLevel, error) {
	var (
		p   float64
		q   float64
		err error
	)
	p, err = strconv.ParseFloat(price, 64)
	if err != nil {
		return Nil, err
	}
	q, err = strconv.ParseFloat(quantity, 64)
	if err != nil {
		return Nil, err
	}
	return PriceLevel{Price: p, Quantity: q}, nil
}

func (p *PriceLevel) New(price float64, quantity float64) *PriceLevel {
	p.Price = price
	p.Quantity = quantity
	return p
}

// Ask is a type alias for PriceLevel.
type Ask = PriceLevel

// Bid is a type alias for PriceLevel.
type Bid = PriceLevel
