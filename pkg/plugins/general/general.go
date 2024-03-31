package general

import (
	"github.com/shopspring/decimal"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/setting"
	"strings"
)

var glg = log.New("plugins.general")

type Exchange string
type Asset string // Like ETH, BTC, USDT, etc
type Symbol struct {
	BaseAsset  Asset
	QuoteAsset Asset
} // like ETHUSDT, BTCUSDT, etc, 每个所的命名方式可能会有区别

func NewSymbol(baseAsset Asset) Symbol {
	return Symbol{
		BaseAsset:  baseAsset,
		QuoteAsset: DefaultQuoteCoin,
	}
}

const (
	TradingSpot        = "spot"
	TradingDerivatives = "derivatives"

	Binance  Exchange = "binance"
	MEXC     Exchange = "MEXC"
	CoinEX   Exchange = "coinEx"
	OKX      Exchange = "okx"
	CoinBase Exchange = "coinBase"
)

const (
	USDT Asset = "USDT"

	NEO Asset = "NEO"
	OGN Asset = "OGN"

	ETH     Asset = "ETH"
	INJ     Asset = "INJ"
	WOO     Asset = "WOO"
	OG      Asset = "OG"
	AVAX    Asset = "AVAX"
	LEVER   Asset = "LEVER"
	BNB     Asset = "BNB"
	YFI     Asset = "YFI"
	SHIB    Asset = "SHIB"
	N_1INCH Asset = "1INCH"
	UNI     Asset = "UNI"
	AAVE    Asset = "AAVE"
	ALICE   Asset = "ALICE"
	AXS     Asset = "AXS"
	COMP    Asset = "COMP"
	ENJ     Asset = "ENJ"
	SAND    Asset = "SAND"
	OMG     Asset = "OMG"
	MANA    Asset = "MANA"
	LINK    Asset = "LINK"
	SNX     Asset = "SNX"

	UnKnown Asset = "UnKnown"

	DefaultQuoteCoin = USDT
)

func GetSecretsForExchanger(exchange Exchange) *setting.Secret {
	switch exchange {
	case Binance:
		return &setting.SecretsConf.Binance
	case MEXC:
		return &setting.SecretsConf.MEXC
	default:
		return nil
	}
}

// SymbolBasicInfo including asset precision, quota asset, quota asset precision, etc.
type SymbolBasicInfo struct {
	Symbol              string `json:"symbol"`
	BaseAsset           string `json:"baseAsset"`
	BaseAssetPrecision  int32  `json:"baseAssetPrecision"` // 似乎没啥用
	QuoteAsset          string `json:"quoteAsset"`
	QuoteAssetPrecision int32  `json:"quoteAssetPrecision"` // 似乎没啥用

	// Price
	MinPrice          decimal.Decimal
	MaxPrice          decimal.Decimal
	TickSize          decimal.Decimal // for price
	TickSizePrecision int32
	MinQuantity       decimal.Decimal
	MaxQuantity       decimal.Decimal
	StepSize          decimal.Decimal // for Quantity
	StepSizePrecision int32
	MinNotional       decimal.Decimal // Price * Quantity quoteAmountPrecision	string	最小下单金额
}

func ConvertPrecisionFromStringToInt(precision string) int32 {
	index := strings.Index(precision, ".")
	if index == -1 {
		return int32(0 - (len(precision) - 1))
	}
	oneIndex := strings.Index(precision, "1")
	return int32(oneIndex - index)
}

func ConvertPrecisionFromIntToDecimal(p int32) decimal.Decimal {
	return decimal.NewFromInt(1).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt32(p)))
}

func NewDecimalFromStringIgnoreErr(s string) decimal.Decimal {
	res, _ := decimal.NewFromString(s)
	return res
}

func IsErrNil(err error) string {
	if err == nil {
		return "ok"
	} else {
		return "err"
	}
}
