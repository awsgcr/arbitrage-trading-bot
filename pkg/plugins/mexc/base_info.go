package mexc

import (
	"errors"
	"fmt"
	"jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
)

type BaseInfoManager struct {
	SymbolsMap map[string]*SymbolBasicInfo
}

func newBaseInfoManager() (BaseInterface, error) {
	s := &BaseInfoManager{
		SymbolsMap: make(map[string]*SymbolBasicInfo),
	}
	err := s.syncExchangeInfo()
	return s, err
}

func (s *BaseInfoManager) ServerTime() (serverTime int64, err error) {
	params := http.Params{}
	data, err := httpGetData(serverTimeEndpoint, params)
	if err != nil {
		return 0, err
	}
	return data.Get("serverTime").MustInt64(), nil
}

func (s *BaseInfoManager) syncExchangeInfo() error {
	j, err := httpGetData(exchangeInfoEndpoint, http.Params{})
	if err != nil {
		return err
	}

	symbolsLen := len(j.Get("symbols").MustArray())
	for i := 0; i < symbolsLen; i++ {
		item := j.Get("symbols").GetIndex(i)

		symbol := item.Get("symbol").MustString()
		baseAssetPrecision := int32(item.Get("baseAssetPrecision").MustInt())
		quoteAssetPrecision := int32(item.Get("quoteAssetPrecision").MustInt())
		s.SymbolsMap[symbol] = &SymbolBasicInfo{
			Symbol:              symbol,
			BaseAsset:           item.Get("baseAsset").MustString(),
			BaseAssetPrecision:  baseAssetPrecision,
			QuoteAsset:          item.Get("quoteAsset").MustString(),
			QuoteAssetPrecision: quoteAssetPrecision,

			TickSize:          ConvertPrecisionFromIntToDecimal(quoteAssetPrecision),
			TickSizePrecision: quoteAssetPrecision,
			MinQuantity:       NewDecimalFromStringIgnoreErr(item.Get("baseSizePrecision").MustString()),
			StepSize:          ConvertPrecisionFromIntToDecimal(baseAssetPrecision),
			StepSizePrecision: baseAssetPrecision,
			MinNotional:       NewDecimalFromStringIgnoreErr(item.Get("quoteAmountPrecision").MustString()), //quoteAmountPrecision	string	最小下单金额
		}
	}

	return nil
}

// GetSymbolBasicInfo TODO: Get from server
func (s *BaseInfoManager) GetSymbolBasicInfo(symbol Symbol) (*SymbolBasicInfo, error) {
	symbol2USDT := getSymbolAlias(symbol)
	info := s.SymbolsMap[symbol2USDT]
	if info == nil {
		return nil, errors.New(fmt.Sprintf("symbol[%s] not supported", symbol2USDT))
	}
	return info, nil
}

// GetSymbolsBasicInfo TODO: Get from server
func (s *BaseInfoManager) GetSymbolsBasicInfo() map[Symbol]*SymbolBasicInfo {
	var res = make(map[Symbol]*SymbolBasicInfo)
	for _, asset := range supportedAssets {
		symbol := Symbol{
			BaseAsset:  asset,
			QuoteAsset: DefaultQuoteCoin,
		}
		symbol2USDT := getSymbolAlias(symbol)
		info := s.SymbolsMap[symbol2USDT]
		if info != nil {
			res[symbol] = info
		}
	}
	return res
}
