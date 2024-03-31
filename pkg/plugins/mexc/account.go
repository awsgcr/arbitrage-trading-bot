package mexc

import (
	"context"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"jasonzhu.com/coin_labor/core/components/alerting"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/components/metrics"
	"jasonzhu.com/coin_labor/core/setting"
	. "jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"net/http"
	"time"
)

type AccountManager struct {
	lg                 log.Logger
	secret             *setting.Secret
	client             *Client
	balancesMapAtStart map[Asset]Balance

	userStreamListenKey string
}

func newAccountManager() AccountInterface {
	secret := GetSecretsForExchanger(MEXC)
	return &AccountManager{
		lg:     plg.New("s", "account"),
		secret: secret,
		client: NewHMACClient(secret, baseAPIMainURL, apiKeyHeader),
	}
}

// GetAccountInfo https://mxcdevelop.github.io/apidocs/spot_v3_cn/#bd9157656f
/**
Response Example
{
    "makerCommission": 20,
    "takerCommission": 20,
    "buyerCommission": 0,
    "sellerCommission": 0,
    "canTrade": true,
    "canWithdraw": true,
    "canDeposit": true,
    "updateTime": null,
    "accountType": "SPOT",
    "balances": [
        {
            "asset": "MX",
            "free": "3",
            "locked": "0"
        },
        {
            "asset": "BTC",
            "free": "0.0003",
            "locked": "0"
        }
    ],
    "permissions": [
        "SPOT"
    ]
}
*/
func (s *AccountManager) GetAccountInfo() (*Account, error) {
	r := &Request{
		Method:   http.MethodGet,
		Endpoint: accountInfoEndpoint,
		SecType:  SecTypeSigned,
	}
	data, err := s.client.CallAPI(context.Background(), r)
	if err != nil {
		return nil, err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return nil, err
	}
	size := len(j.Get("balances").MustArray())
	balances := make([]Balance, size)
	for i := 0; i < size; i++ {
		item := j.Get("balances").GetIndex(i)
		free := NewDecimalFromStringIgnoreErr(item.Get("free").MustString())
		locked := NewDecimalFromStringIgnoreErr(item.Get("locked").MustString())
		if free.GreaterThan(decimal.Zero) || locked.GreaterThan(decimal.Zero) {
			asset := item.Get("asset").MustString()
			balances[i] = Balance{
				Asset:  ToAsset(asset),
				Free:   free,
				Locked: locked,
			}
		}
	}
	if s.balancesMapAtStart == nil {
		s.balancesMapAtStart = make(map[Asset]Balance)
		for _, balance := range balances {
			s.balancesMapAtStart[balance.Asset] = balance
		}
	}
	a := &Account{
		UpdateTime: uint64(time.Now().UnixMilli()),
	}
	a.InitBalances(balances)
	return a, nil
}

func (s *AccountManager) GetBalanceAtStart(asset Asset) *Balance {
	if b, ok := s.balancesMapAtStart[asset]; ok {
		return &b
	}
	return nil
}

func (s *AccountManager) WsWatchUserDataChanges(ctx context.Context, eventC chan *UserDataEvent) error {
	s.lg.Warn("MEXC Account订阅开启")
	var err error
	err = s.createListenKey()
	if err != nil {
		return err
	}
	s.keepaliveListenKey(ctx)

	wsHandler := func(event *UserDataEvent) {
		eventC <- event
		go func() {
			metrics.M_Coin_UserDataWatch_Latency_Summary.WithLabelValues(string(MEXC)).Observe(float64(event.Time - uint64(time.Now().Unix())))
			metrics.M_Coin_UserDataWatch_Latency_Histogram.WithLabelValues(string(MEXC)).Observe(float64(event.Time - uint64(time.Now().Unix())))
		}()
	}
	errHandler := func(err error) {
		s.lg.Error("failed to fetch account changing messages from websocket.", "err", err)
		alerting.NotifyRightNow(err, "error occurred when fetching UserData from MEXC websocket.")
		DefaultHealthChecker.Declare(MEXCUserDataWatchFeature, HealthStateUnhealthy)
	}
	wsServe, err := WsUserDataServe(s.userStreamListenKey, wsHandler, errHandler)
	if err != nil {
		return nil
	}
	DefaultHealthChecker.Declare(MEXCUserDataWatchFeature, HealthStateHealthy)
	<-wsServe.DoneC()
	DefaultHealthChecker.Declare(MEXCUserDataWatchFeature, HealthStateUnhealthy)
	s.lg.Warn("MEXC Account订阅关闭")
	return nil
}

func (s *AccountManager) createListenKey() error {
	r := &Request{
		Method:   http.MethodPost,
		Endpoint: listenKeyEndpoint,
		SecType:  SecTypeSigned,
	}
	data, err := s.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}
	j, err := simplejson.NewJson(data)
	if err != nil {
		return err
	}
	s.userStreamListenKey = j.Get("listenKey").MustString()
	return nil
}

func (s *AccountManager) keepaliveListenKey(ctx context.Context) {
	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error {
		ticker := time.NewTicker(25 * time.Minute)
		for range ticker.C {
			r := &Request{
				Method:   http.MethodPut,
				Endpoint: listenKeyEndpoint,
				SecType:  SecTypeSigned,
			}
			r.SetParam("listenKey", s.userStreamListenKey)
			_, err := s.client.CallAPI(context.Background(), r)
			if err != nil {
				s.lg.Error("failed to keep alive listen key for account update", "err", err)
			}
		}
		return nil
	})
}

// WsUserDataHandler handle WsUserDataEvent
type WsUserDataHandler func(event *UserDataEvent)

// WsUserDataServe serve user data handler with listen key
// 如：wss://wbs.mexc.com/ws?listenKey=pqia91ma19a5s61cv6a81va65sd099v8a65a1a5s61cv6a81va65sdf19v8a65a1
/**
现货账户信息(实时)
在订阅成功后，每当账户余额发生变动或可用余额发生变动时，服务器将推送账户资产的更新。
request:
{
    "method": "SUBSCRIPTION",
    "params": [
    "spot@private.account.v3.api"
    ]
}

response:
{
    "c": "spot@private.account.v3.api",
    "d": {
        "a": "USDT",
        "c": 1681574587222,
        "f": "256",
        "fd": "-200",
        "l": "200",
        "ld": "200",
        "o": "ENTRUST_PLACE"
    },
    "t": 1681574587227
}


---------------------------------------------------------------------------
现货账户订单(实时)
request:
{
  "method": "SUBSCRIPTION",
  "params": [
      "spot@private.orders.v3.api"
  ]
}

a.限价/市价订单 (实时)
response:
{
    "c": "spot@private.orders.v3.api",
    "d": {
        "i": "bd95402a04ff4f068028ab1954931060",
        "c": "",
        "o": 1,
        "p": 200.00,
        "v": 1.00000,
        "S": 1,
        "a": 200,
        "m": 0,
        "A": 200,
        "V": 1.00000,
        "s": 1,
        "O": 1681574587218,
        "ap": 0,
        "cv": 0.00000,
        "ca": 0
    },
    "s": "ETHUSDT",
    "t": 1681574587230
}
*/
func WsUserDataServe(listenKey string, handler WsUserDataHandler, errHandler ErrHandler) (wsServe *WsServe, err error) {
	endpoint := fmt.Sprintf("%s?listenKey=%s", baseWSMainURL, listenKey)
	wsHandler := func(message []byte) {
		j, err := simplejson.NewJson(message)
		if err != nil {
			plg.Error("websocket error: line 248", "err", err)
			errHandler(err)
			return
		}

		//plg.Info("print userDataEvent of MEXC", string(message))
		eventType := j.Get("c").MustString()
		if eventType == "" {
			return
		}
		timeIntValue := j.Get("t").MustUint64()

		var event *UserDataEvent
		switch UserDataEventType(eventType) {
		case spotAccountMsg: //https://www.MEXC.me/docs/v1/websocket/payload/updateAccount
			/**
			d	json	账户信息
			> a	string	资产名称
			> c	long	结算时间
			> f	string	可用余额
			> fd	string	可用变动金额
			> l	string	冻结余额
			> ld	string	冻结变动金额
			> o	string	变动类型
			t	long	事件时间
			*/
			var balances = make([]WsAccountUpdate, 1)
			item := j.Get("d")
			balances[0] = WsAccountUpdate{
				Asset:  ToAsset(item.Get("a").MustString()),
				Free:   NewDecimalFromStringIgnoreErr(item.Get("f").MustString()),
				Locked: NewDecimalFromStringIgnoreErr(item.Get("l").MustString()),
			}
			event = &UserDataEvent{
				Event: UserDataEventTypeOutboundAccountPosition, //             UserDataEventType `json:"e"`
				Time:  timeIntValue,                             //              int64             `json:"E"`
				AccountUpdate: WsAccountUpdateList{
					WsAccountUpdates: balances,
				},
			}
		case spotOrdersMsg: //https://www.MEXC.me/docs/v1/websocket/payload/updateOrder

			//print OrderStatusTypeFilled &{
			//map[c:spot@private.orders.v3.api
			//	d:map[
			//		A:0 O:1685375848665 S:1 V:0.00
			//		a:10.269 下单总金额
			//		ap:7.335 平均成交价
			//		c:c8347b6237794768b7ee3c42ddaaa144
			//		ca:10.269 累计成交金额
			//		cv:1.40 累计成交数量
			//		i:6d54587710b441839a1171ff9df05972
			//		lv:1.40
			//		m:1 o:1
			//		p:7.335 下单价格
			//		s:2
			//		v:1.40 下单数量
			//	] s:INJUSDT t:1685375849373]}
			/**
			d	json	账户订单信息
			> A	bigDecimal	实际剩余金额: remainAmount
			> O	long	订单创建时间
			> S	int	交易类型 1:买 2:卖
			> V	bigDecimal	实际剩余数量: remainQuantity
			> a	bigDecimal	下单总金额
			> c	string	用户自定义订单id: clientOrderId
			> i	string	订单id
			> m	int	是否是挂单: isMaker
			> o	int	订单类型LIMIT_ORDER(1),POST_ONLY(2),IMMEDIATE_OR_CANCEL(3),
			FILL_OR_KILL(4),MARKET_ORDER(5); 止盈止损（100）
			> p	bigDecimal	下单价格
			> s	int	订单状态 1:未成交 2:已成交 3:部分成交 4:已撤单 5:部分撤单
			> v	bigDecimal	下单数量
			> ap	bigDecimal	平均成交价
			> cv	bigDecimal	累计成交数量
			> ca	bigDecimal	累计成交金额
			t	long	事件时间
			s	string	交易对
			*/
			item := j.Get("d")
			var side SideType
			switch item.Get("S").MustInt() { //S	int	交易类型 1:买 2:卖
			case 1:
				side = SideTypeBuy
			case 2:
				side = SideTypeSell
			}

			var orderType OrderType
			switch item.Get("o").MustInt() { //	订单类型LIMIT_ORDER(1),POST_ONLY(2),IMMEDIATE_OR_CANCEL(3), FILL_OR_KILL(4),MARKET_ORDER(5); 止盈止损（100）
			case 1:
				orderType = OrderTypeLimit
			case 5:
				orderType = OrderTypeMarket
			}

			var status OrderStatusType
			switch item.Get("s").MustInt() { //s	int	订单状态 1:未成交 2:已成交 3:部分成交 4:已撤单 5:部分撤单
			case 1:
				status = OrderStatusTypeNew
			case 2:
				status = OrderStatusTypeFilled
				go func() {
					fmt.Println("print OrderStatusTypeFilled", j)
				}()

			case 3:
				status = OrderStatusTypePartiallyFilled
				go func() {
					fmt.Println("print OrderStatusTypePartiallyFilled", j)
				}()

			case 4:
				status = OrderStatusTypeCanceled
			case 5:
				status = OrderStatusTypeCanceled
			}
			symbol := newSymbolFromString(j.Get("s").MustString())

			// Test
			//fmt.Println("print Original", string(message))
			//cvStr, err := item.Get("cv").Float64()
			//fmt.Println("print CV", item.Get("cv"), cvStr, err)
			//apStr, err := item.Get("ap").Float64()
			//fmt.Println("print AP", item.Get("ap"), apStr, err)

			event = &UserDataEvent{
				Event:           UserDataEventTypeExecutionReport,
				Time:            timeIntValue,
				TransactionTime: j.Get("t").MustInt64(),
				OrderUpdate: WsOrderUpdate{
					Symbol:        symbol,
					ClientOrderId: item.Get("c").MustString(),
					Side:          side,
					Type:          orderType,

					/** 			//		a:10.269 下单总金额
					//		ap:7.335 平均成交价
					//		c:c8347b6237794768b7ee3c42ddaaa144
					//		ca:10.269 累计成交金额
					//		cv:1.40 累计成交数量
					//		i:6d54587710b441839a1171ff9df05972
					//		lv:1.40
					//		m:1 o:1
					//		p:7.335 下单价格
					//		s:2
					//		v:1.40 下单数量
					*/
					//TimeInForce:             TimeInForceType(j.Get("f").MustString()),
					Volume:      decimal.NewFromFloat(item.Get("cv").MustFloat64()),
					Price:       decimal.NewFromFloat(item.Get("ap").MustFloat64()),
					LatestPrice: decimal.NewFromFloat(item.Get("ap").MustFloat64()),
					Status:      status, // order status
					//RejectReason:      j.Get("r").MustString(),
					Id:           item.Get("i").MustInt64(), // order id
					FilledVolume: decimal.NewFromFloat(item.Get("cv").MustFloat64()),
					//TransactionTime:   j.Get("T").MustInt64(),
					IsMaker:    item.Get("m").MustBool(), // is this order maker?
					CreateTime: item.Get("O").MustInt64(),
					//FilledQuoteVolume: NewDecimalFromStringIgnoreErr(item.Get("A").MustString()), // the quote volume that already filled
				},
			}
		}

		handler(event)
	}
	wsServe, err = NewWsServe(endpoint, wsHandler)
	if err != nil {
		return nil, err
	}

	go func() {
		wsServe.Write(SubEvent{
			Method: "SUBSCRIPTION",
			Params: []string{spotAccountMsg, spotOrdersMsg},
		})
	}()
	return wsServe, nil
}

const (
	spotAccountMsg = "spot@private.account.v3.api"
	spotOrdersMsg  = "spot@private.orders.v3.api"
	spotDealsMsg   = "spot@private.deals.v3.api"
)

type SubEvent struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
}
