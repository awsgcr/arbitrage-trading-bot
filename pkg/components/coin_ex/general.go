package coin_ex

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"jasonzhu.com/coin_labor/core/util/http"
)

const baseAPIMainURL = "https://api.coinex.com/v1" // 公有接口每 IP 每秒 20次
const (
	serverTimeEndpoint = "/openapi/v1/time"
	orderBookEndpoint  = "/market/depth"
)

func httpGetData(endpoint string, params http.Params) (*simplejson.Json, error) {
	data, err := http.Get(fmt.Sprintf("%s%s", baseAPIMainURL, endpoint), params)
	if err != nil {
		return nil, err
	}

	j, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return j.Get("data"), nil
}
