package binance

import (
	"fmt"
	"testing"
)

func TestConvertManager_GetQuote(t *testing.T) {
	manager := newConvertManager()
	err := manager.GetQuote()
	fmt.Println(err)
}
