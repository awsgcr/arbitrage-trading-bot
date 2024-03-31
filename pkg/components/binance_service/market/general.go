package market

import "github.com/adshao/go-binance/v2"

func GetBinanceClient() *binance.Client {
	var (
		apiKey    = "your api key"
		secretKey = "your secret key"
	)
	return binance.NewClient(apiKey, secretKey)
}
