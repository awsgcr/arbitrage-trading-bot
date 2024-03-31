package binance

import (
	"fmt"
	"github.com/adshao/go-binance/v2"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/setting"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"strings"
)

const exName = Binance

var plg = log.New(fmt.Sprintf("plugin.%s", exName))

var supportedAssets = []Asset{
	INJ,
	WOO,
	OG,
	NEO,
	AAVE,
	OGN,
}

func isAssetSupported(symbol Symbol) bool {
	for _, s := range supportedAssets {
		if symbol.BaseAsset == s {
			return true
		}
	}
	return false
}

func buildSymbolWithDefaultQuoteCoin(asset Asset) Symbol {
	return Symbol{
		BaseAsset:  asset,
		QuoteAsset: DefaultQuoteCoin,
	}
}
func getSymbolAliasWithDefaultQuoteCoin(asset Asset) string {
	return getSymbolAlias(Symbol{BaseAsset: asset, QuoteAsset: DefaultQuoteCoin})
}
func getSymbolAlias(symbol Symbol) string {
	return string(symbol.BaseAsset) + string(symbol.QuoteAsset)
}

// Convert from string, like ETHUSDT
func newSymbolFromString(symbol string) Symbol {
	if strings.HasSuffix(strings.ToUpper(symbol), string(DefaultQuoteCoin)) {
		baseCoin := symbol[0 : len(symbol)-len(DefaultQuoteCoin)]
		for _, s := range supportedAssets {
			if baseCoin == string(s) {
				return Symbol{
					BaseAsset:  ToAsset(baseCoin),
					QuoteAsset: DefaultQuoteCoin,
				}
			}
		}
	}
	return Symbol{
		BaseAsset:  UnKnown,
		QuoteAsset: DefaultQuoteCoin,
	}
}

var (
	apiKey    = ""
	secretKey = ""
)

func getBinanceClient(secret *setting.Secret) *binance.Client {
	if secret == nil {
		return binance.NewClient(apiKey, secretKey)
	}
	return binance.NewClient(secret.Key, secret.Secret)
}
