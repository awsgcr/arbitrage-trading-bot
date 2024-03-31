package general

import (
	"fmt"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestNewPriceLevelFromString(t *testing.T) {
	pl, _ := NewPriceLevelFromString("0.003", "0.001")
	fmt.Println(pl.Price.Sub(pl.Quantity))

	p, _ := decimal.NewFromString("0.003")
	q, _ := decimal.NewFromString("0.001")
	fmt.Println(p.Add(q).Round(3).String())

	precision := "0.00001"
	index := strings.Index(precision, ".")
	fmt.Println(len(precision[index:]) - 1)

	fmt.Println(genClientOrderID())

	fmt.Println(ConvertPrecisionFromIntToDecimal(-1))
	fmt.Println(ConvertPrecisionFromStringToInt("10"))

	switch "test" {
	case "test":
		fmt.Println(1)
	case "a":
		fmt.Println(2)
	}

	testDefer()
}

func testDefer() (res string) {
	res = "test"
	defer fmt.Println(res)
	return
}

func TestGetSecretsForExchanger(t *testing.T) {
	secret := GetSecretsForExchanger(Binance)
	fmt.Println(secret)
}
