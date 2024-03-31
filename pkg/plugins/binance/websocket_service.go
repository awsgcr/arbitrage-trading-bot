package binance

import (
	stdjson "encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"strings"
)

// Endpoints
const (
	baseWsMainURL          = "wss://stream.binance.com:9443/ws"
	baseWsTestnetURL       = "wss://testnet.binance.vision/ws"
	baseCombinedMainURL    = "wss://stream.binance.com:9443/stream?streams="
	baseCombinedTestnetURL = "wss://testnet.binance.vision/stream?streams="
)

// getWsEndpoint return the base endpoint of the WS according the UseTestnet flag
func getWsEndpoint() string {
	if UseTestnet {
		return baseWsTestnetURL
	}
	return baseWsMainURL
}

// getCombinedEndpoint return the base endpoint of the combined stream according the UseTestnet flag
func getCombinedEndpoint() string {
	if UseTestnet {
		return baseCombinedTestnetURL
	}
	return baseCombinedMainURL
}

// WsPartialDepthEvent define websocket partial depth book event
type WsPartialDepthEvent struct {
	Symbol       string
	LastUpdateID int64 `json:"lastUpdateId"`
	Bids         []Bid `json:"bids"`
	Asks         []Ask `json:"asks"`
}

// WsPartialDepthHandler handle websocket partial depth event
type WsPartialDepthHandler func(event *WsPartialDepthEvent)

// WsPartialDepthServe serve websocket partial depth handler with a symbol, using 1sec updates
func WsPartialDepthServe(symbol string, levels string, handler WsPartialDepthHandler, errHandler ErrHandler) (wsServer *WsServe, err error) {
	endpoint := fmt.Sprintf("%s/%s@depth%s", getWsEndpoint(), strings.ToLower(symbol), levels)
	return wsPartialDepthServe(endpoint, symbol, handler, errHandler)
}

// WsPartialDepthServe100Ms serve websocket partial depth handler with a symbol, using 100msec updates
func WsPartialDepthServe100Ms(symbol string, levels string, handler WsPartialDepthHandler, errHandler ErrHandler) (wsServer *WsServe, err error) {
	endpoint := fmt.Sprintf("%s/%s@depth%s@100ms", getWsEndpoint(), strings.ToLower(symbol), levels)
	return wsPartialDepthServe(endpoint, symbol, handler, errHandler)
}

// WsPartialDepthServe serve websocket partial depth handler with a symbol
func wsPartialDepthServe(endpoint string, symbol string, handler WsPartialDepthHandler, errHandler ErrHandler) (wsServer *WsServe, err error) {
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}
		event := new(WsPartialDepthEvent)
		event.Symbol = symbol
		event.LastUpdateID = j.Get("lastUpdateId").MustInt64()
		bidsLen := len(j.Get("bids").MustArray())
		event.Bids = make([]Bid, bidsLen)
		for i := 0; i < bidsLen; i++ {
			item := j.Get("bids").GetIndex(i)
			event.Bids[i], _ = NewPriceLevelFromString(item.GetIndex(0).MustString(), item.GetIndex(1).MustString())
		}
		asksLen := len(j.Get("asks").MustArray())
		event.Asks = make([]Ask, asksLen)
		for i := 0; i < asksLen; i++ {
			item := j.Get("asks").GetIndex(i)
			event.Asks[i], _ = NewPriceLevelFromString(item.GetIndex(0).MustString(), item.GetIndex(1).MustString())
		}
		handler(event)
	}
	return NewWsServe(endpoint, wsHandler)
}
func WsCombinedPartialDepthServe100Ms(symbolLevels map[string]string, handler WsPartialDepthHandler, errHandler ErrHandler) (*WsServe, error) {
	endpoint := getCombinedEndpoint()
	for s, l := range symbolLevels {
		endpoint += fmt.Sprintf("%s@depth%s@100ms", strings.ToLower(s), l) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	return wsCombinedPartialDepthServe(endpoint, errHandler, handler)
}

// WsCombinedPartialDepthServe is similar to WsPartialDepthServe, but it for multiple symbols
func WsCombinedPartialDepthServe(symbolLevels map[string]string, handler WsPartialDepthHandler, errHandler ErrHandler) (*WsServe, error) {
	endpoint := getCombinedEndpoint()
	for s, l := range symbolLevels {
		endpoint += fmt.Sprintf("%s@depth%s", strings.ToLower(s), l) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	return wsCombinedPartialDepthServe(endpoint, errHandler, handler)
}

func wsCombinedPartialDepthServe(endpoint string, errHandler ErrHandler, handler WsPartialDepthHandler) (*WsServe, error) {
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}
		event := new(WsPartialDepthEvent)
		stream := j.Get("stream").MustString()
		symbol := strings.Split(stream, "@")[0]
		event.Symbol = strings.ToUpper(symbol)
		data := j.Get("data").MustMap()
		event.LastUpdateID, _ = data["lastUpdateId"].(stdjson.Number).Int64()
		bidsLen := len(data["bids"].([]interface{}))
		event.Bids = make([]Bid, bidsLen)
		for i := 0; i < bidsLen; i++ {
			item := data["bids"].([]interface{})[i].([]interface{})
			event.Bids[i], _ = NewPriceLevelFromString(item[0].(string), item[1].(string))
		}
		asksLen := len(data["asks"].([]interface{}))
		event.Asks = make([]Ask, asksLen)
		for i := 0; i < asksLen; i++ {

			item := data["asks"].([]interface{})[i].([]interface{})
			event.Asks[i], _ = NewPriceLevelFromString(item[0].(string), item[1].(string))
		}
		handler(event)
	}
	return NewWsServe(endpoint, wsHandler)
}

// WsDepthHandler handle websocket depth event
type WsDepthHandler func(event *WsDepthEvent)

// WsDepthServe serve websocket depth handler with a symbol, using 1sec updates
func WsDepthServe(symbol string, handler WsDepthHandler, errHandler ErrHandler) (*WsServe, error) {
	endpoint := fmt.Sprintf("%s/%s@depth", getWsEndpoint(), strings.ToLower(symbol))
	return wsDepthServe(endpoint, handler, errHandler)
}

// WsDepthServe100Ms serve websocket depth handler with a symbol, using 100msec updates
func WsDepthServe100Ms(symbol string, handler WsDepthHandler, errHandler ErrHandler) (*WsServe, error) {
	endpoint := fmt.Sprintf("%s/%s@depth@100ms", getWsEndpoint(), strings.ToLower(symbol))
	return wsDepthServe(endpoint, handler, errHandler)
}

// WsDepthServe serve websocket depth handler with an arbitrary endpoint address
func wsDepthServe(endpoint string, handler WsDepthHandler, errHandler ErrHandler) (*WsServe, error) {
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}
		event := new(WsDepthEvent)
		event.Event = j.Get("e").MustString()
		event.Time = j.Get("E").MustInt64()
		event.Symbol = j.Get("s").MustString()
		event.LastUpdateID = j.Get("u").MustInt64()
		event.FirstUpdateID = j.Get("U").MustInt64()
		bidsLen := len(j.Get("b").MustArray())
		event.Bids = make([]Bid, bidsLen)
		for i := 0; i < bidsLen; i++ {
			item := j.Get("b").GetIndex(i)
			event.Bids[i], _ = NewPriceLevelFromString(item.GetIndex(0).MustString(), item.GetIndex(1).MustString())
		}
		asksLen := len(j.Get("a").MustArray())
		event.Asks = make([]Ask, asksLen)
		for i := 0; i < asksLen; i++ {
			item := j.Get("a").GetIndex(i)
			event.Asks[i], _ = NewPriceLevelFromString(item.GetIndex(0).MustString(), item.GetIndex(1).MustString())
		}
		handler(event)
	}
	return NewWsServe(endpoint, wsHandler)
}

// WsDepthEvent define websocket depth event
type WsDepthEvent struct {
	Event         string `json:"e"`
	Time          int64  `json:"E"`
	Symbol        string `json:"s"`
	LastUpdateID  int64  `json:"u"`
	FirstUpdateID int64  `json:"U"`
	Bids          []Bid  `json:"b"`
	Asks          []Ask  `json:"a"`
}

// WsCombinedDepthServe is similar to WsDepthServe, but it for multiple symbols
func WsCombinedDepthServe(symbols []string, handler WsDepthHandler, errHandler ErrHandler) (*WsServe, error) {
	endpoint := getCombinedEndpoint()
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@depth", strings.ToLower(s)) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	return wsCombinedDepthServe(endpoint, handler, errHandler)
}

func WsCombinedDepthServe100Ms(symbols []string, handler WsDepthHandler, errHandler ErrHandler) (*WsServe, error) {
	endpoint := getCombinedEndpoint()
	for _, s := range symbols {
		endpoint += fmt.Sprintf("%s@depth@100ms", strings.ToLower(s)) + "/"
	}
	endpoint = endpoint[:len(endpoint)-1]
	return wsCombinedDepthServe(endpoint, handler, errHandler)
}

func wsCombinedDepthServe(endpoint string, handler WsDepthHandler, errHandler ErrHandler) (*WsServe, error) {
	wsHandler := func(message []byte) {
		j, err := newJSON(message)
		if err != nil {
			errHandler(err)
			return
		}
		event := new(WsDepthEvent)
		stream := j.Get("stream").MustString()
		symbol := strings.Split(stream, "@")[0]
		event.Symbol = strings.ToUpper(symbol)
		data := j.Get("data").MustMap()
		event.Time, _ = data["E"].(stdjson.Number).Int64()
		event.LastUpdateID, _ = data["u"].(stdjson.Number).Int64()
		event.FirstUpdateID, _ = data["U"].(stdjson.Number).Int64()
		bidsLen := len(data["b"].([]interface{}))
		event.Bids = make([]Bid, bidsLen)
		for i := 0; i < bidsLen; i++ {
			item := data["b"].([]interface{})[i].([]interface{})
			event.Bids[i], _ = NewPriceLevelFromString(item[0].(string), item[1].(string))
		}
		asksLen := len(data["a"].([]interface{}))
		event.Asks = make([]Ask, asksLen)
		for i := 0; i < asksLen; i++ {

			item := data["a"].([]interface{})[i].([]interface{})
			event.Asks[i], _ = NewPriceLevelFromString(item[0].(string), item[1].(string))
		}
		handler(event)
	}
	return NewWsServe(endpoint, wsHandler)
}

func newJSON(data []byte) (j *simplejson.Json, err error) {
	j, err = simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	return j, nil
}
