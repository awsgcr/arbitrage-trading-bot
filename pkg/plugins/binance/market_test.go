package binance

import (
	"fmt"
	"jasonzhu.com/coin_labor/pkg/plugins/general"
	"testing"
	"time"
)

var assets = []string{
	//"LTC", "INJ", "WOO",
	"LTC", "INJ", "WOO", "UNI", "MANA", "USDC", "EGLD", "HIFI", "HBAR", "DASH", "BOND", "ONE", "OGN", "RNDR", "ETC", "MAGIC", "WBTC", "ALPINE", "ACM", "SSV", "BICO", "IDEX", "ZRX", "YFI", "FTT", "ANT", "CTK", "CELR", "IQ", "MKR", "KSM", "IOTX", "APE", "IOST", "NEO", "SXP", "BAR", "AVA", "UNFI", "KAVA", "AVA", "QNT", "ENS", "PHB", "GLMR", "GMX",
}
var dot = ","

func TestBinancePlugin_GetMarketInfoManager(t *testing.T) {
	manager := newMarketInfoManager()

	for _, asset := range assets {
		depth, err := manager.FetchDepth(general.NewSymbol(general.Asset(asset)), 5)
		if err != nil {
			fmt.Println(fillWithBlank(asset, 6), dot, err)
			continue
		}
		ask, err := depth.TopAsk()
		if err != nil {
			fmt.Println(fillWithBlank(asset, 6), dot, err)
			continue
		}
		fmt.Println(fillWithBlank(asset, 6), dot, dot, ask.Price, dot, depth.Asks[1].Price, dot, depth.Asks[2].Price)
		time.Sleep(100 * time.Millisecond)
	}
}

func fillWithBlank(s string, size int) string {
	if len(s) > size {
		return s[0:size]
	} else {
		num := size - len(s)
		blank := ""
		for i := 0; i < num; i++ {
			blank += " "
		}
		return s + blank
	}
}
