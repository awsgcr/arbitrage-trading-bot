package binance

import (
	"context"
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
)

type BaseInfoManager struct {
	ExchangeInfo *binance.ExchangeInfo
	SymbolsMap   map[string]*SymbolBasicInfo
	client       *binance.Client
}

func newBaseInfoManager() (BaseInterface, error) {
	s := &BaseInfoManager{
		SymbolsMap: make(map[string]*SymbolBasicInfo),
		client:     getBinanceClient(nil),
	}
	_, err := s.syncExchangeInfo()
	return s, err
}

func (s *BaseInfoManager) ServerTime() (serverTime int64, err error) {
	return s.client.NewSetServerTimeService().Do(context.Background())
}

// FetchExchangeInfo demo: https://api.binance.com/api/v3/exchangeInfo?symbol=BNBBTC
func (s *BaseInfoManager) syncExchangeInfo() (*binance.ExchangeInfo, error) {
	symbols := make([]string, len(supportedAssets))
	for i, asset := range supportedAssets {
		symbols[i] = getSymbolAlias(buildSymbolWithDefaultQuoteCoin(asset))
	}
	res, err := s.client.NewExchangeInfoService().Symbols(symbols...).Do(context.Background())
	if err != nil {
		return nil, err
	}

	s.ExchangeInfo = res
	for _, symbol := range res.Symbols {
		priceFilter := symbol.PriceFilter()
		lotSizeFilter := symbol.LotSizeFilter()
		minNotionalFilter := symbol.MinNotionalFilter()
		info := &SymbolBasicInfo{
			Symbol:              symbol.Symbol,
			BaseAsset:           symbol.BaseAsset,
			BaseAssetPrecision:  int32(symbol.BaseAssetPrecision),
			QuoteAsset:          symbol.QuoteAsset,
			QuoteAssetPrecision: int32(symbol.QuoteAssetPrecision),
		}
		if priceFilter != nil {
			info.MinPrice = NewDecimalFromStringIgnoreErr(priceFilter.MinPrice)
			info.MaxPrice = NewDecimalFromStringIgnoreErr(priceFilter.MaxPrice)
			info.TickSize = NewDecimalFromStringIgnoreErr(priceFilter.TickSize)
			info.TickSizePrecision = ConvertPrecisionFromStringToInt(priceFilter.TickSize)
		}
		if lotSizeFilter != nil {
			info.MinQuantity = NewDecimalFromStringIgnoreErr(lotSizeFilter.MinQuantity)
			info.MaxQuantity = NewDecimalFromStringIgnoreErr(lotSizeFilter.MaxQuantity)
			info.StepSize = NewDecimalFromStringIgnoreErr(lotSizeFilter.StepSize)
			info.StepSizePrecision = ConvertPrecisionFromStringToInt(lotSizeFilter.StepSize)
		}
		if minNotionalFilter != nil {
			info.MinNotional = NewDecimalFromStringIgnoreErr(minNotionalFilter.MinNotional)
		}
		s.SymbolsMap[symbol.Symbol] = info
	}
	return res, nil
}

func (s *BaseInfoManager) GetSymbolBasicInfo(symbol Symbol) (*SymbolBasicInfo, error) {
	symbol2USDT := getSymbolAlias(symbol)
	info := s.SymbolsMap[symbol2USDT]
	if info == nil {
		return nil, errors.New(fmt.Sprintf("symbol[%s] not supported", symbol2USDT))
	}
	return info, nil
}

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
