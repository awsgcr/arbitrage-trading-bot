package binance

import (
	"context"
	"fmt"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/components/simplejson"
	"jasonzhu.com/coin_labor/core/setting"
	. "jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"net/http"
)

type ConvertManager struct {
	lg     log.Logger
	secret *setting.Secret
	client *Client
}

func newConvertManager() *ConvertManager {
	secret := GetSecretsForExchanger(Binance)
	return &ConvertManager{
		lg:     log.New("binance.convert_manager"),
		secret: secret,
		client: NewHMACClient(secret, "https://api.binance.com", "X-MBX-APIKEY"),
	}
}

func (s *ConvertManager) GetQuote() error {
	r := &Request{
		Method:   http.MethodPost,
		Endpoint: "/sapi/v1/convert/getQuote",
		SecType:  SecTypeSigned,
	}

	//fromAsset	STRING	YES
	//toAsset	STRING	YES
	//fromAmount	DECIMAL	EITHER	这是成交后将被扣除的金额
	//toAmount	DECIMAL	EITHER	这是成交后将会获得的金额
	r.SetParam("fromAsset", INJ)
	r.SetParam("toAsset", USDT)
	//r.SetParam("fromAmount", 10)
	r.SetParam("toAmount", 3)
	data, err := s.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	fmt.Println(j)
	return nil
}
