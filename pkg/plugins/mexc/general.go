package mexc

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"strings"
)

const (
	exName = MEXC

	baseAPIMainURL = "https://api.mexc.com"
	baseWSMainURL  = "wss://wbs.mexc.com/ws"

	// Header
	apiKeyHeader = "X-MEXC-APIKEY"

	// Base Info
	serverTimeEndpoint   = "/api/v3/time"
	exchangeInfoEndpoint = "/api/v3/exchangeInfo"

	// Market
	orderBookEndpoint = "/api/v3/depth"

	// Account
	accountInfoEndpoint = "/api/v3/account"

	// Order
	openOrdersEndpoint = "/api/v3/openOrders"  // https://mxcdevelop.github.io/apidocs/spot_v3_cn/#066ca582c9
	allOrdersEndpoint  = "/api/v3/allOrders"   // https://mxcdevelop.github.io/apidocs/spot_v3_cn/#90376e83a0
	orderEndpoint      = "/api/v3/order"       // POST create; GET query; DELETE cancel
	batchOrderEndpoint = "/api/v3/batchOrders" // POST create; GET query; DELETE cancel
	myTradesEndpoint   = "/api/v3/myTrades"

	// WS
	listenKeyEndpoint = "/api/v3/userDataStream" // POST create; PUT Keep-alive; DELETE close

)

var plg = log.New(fmt.Sprintf("plugin.%s", exName))

var supportedAssets = []Asset{
	OG, AVAX, AAVE, LEVER, WOO, INJ, NEO, OGN,
}

func isAssetSupported(asset Asset) bool {
	for _, s := range supportedAssets {
		if asset == s {
			return true
		}
	}
	return false
}

func getSymbolAlias(symbol Symbol) string {
	return string(symbol.BaseAsset) + string(symbol.QuoteAsset)
}

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

func httpGetData(endpoint string, params http.Params) (*simplejson.Json, error) {
	data, err := http.Get(fmt.Sprintf("%s%s", baseAPIMainURL, endpoint), params)
	if err != nil {
		return nil, err
	}

	j, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return j, nil
}
