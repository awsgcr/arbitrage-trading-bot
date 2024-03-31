package binance

import (
	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"jasonzhu.com/coin_labor/core/components/alerting"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/components/metrics"
	"jasonzhu.com/coin_labor/core/setting"
	. "jasonzhu.com/coin_labor/pkg/plugins/general"
	"time"
)

type AccountManager struct {
	lg                 log.Logger
	secret             *setting.Secret
	client             *binance.Client
	balancesMapAtStart map[Asset]Balance

	userStreamListenKey string //现货账户
}

func newAccountManager() AccountInterface {
	secret := GetSecretsForExchanger(Binance)
	return &AccountManager{
		lg:     log.New("binance.account_manager"),
		secret: secret,
		client: getBinanceClient(secret),
	}
}

func (s *AccountManager) GetAccountInfo() (*Account, error) {
	res, err := s.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	var balances []Balance
	for _, balance := range res.Balances {
		free, _ := decimal.NewFromString(balance.Free)
		locked, _ := decimal.NewFromString(balance.Locked)
		if free.GreaterThan(decimal.Zero) || locked.GreaterThan(decimal.Zero) {
			balances = append(balances, Balance{
				Asset:  ToAsset(balance.Asset),
				Free:   free,
				Locked: locked,
			})
		}
	}
	if s.balancesMapAtStart == nil {
		s.balancesMapAtStart = make(map[Asset]Balance)
		for _, balance := range balances {
			s.balancesMapAtStart[balance.Asset] = balance
		}
	}
	a := &Account{
		MakerCommission:  res.MakerCommission,
		TakerCommission:  res.TakerCommission,
		BuyerCommission:  res.BuyerCommission,
		SellerCommission: res.SellerCommission,
		CanTrade:         res.CanTrade,
		CanWithdraw:      res.CanWithdraw,
		CanDeposit:       res.CanDeposit,
		UpdateTime:       res.UpdateTime,
		AccountType:      res.AccountType,
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
	var err error
	s.userStreamListenKey, err = s.client.NewStartUserStreamService().Do(context.Background())
	if err != nil {
		return err
	}

	s.keepaliveListenKey(ctx)

	wsHandler := func(bEvent *BWsUserDataEvent) {
		//Payload: 账户更新 - outboundAccountPosition
		//每当帐户余额发生更改时，包含可能由生成余额变动的事件而变动的资产。

		//Payload: 余额更新 - balanceUpdate
		//账户发生充值或提取
		//交易账户之间发生划转(例如 现货向杠杆账户划转)

		//Payload: 订单更新 - executionReport
		//https://binance-docs.github.io/apidocs/spot/cn/#payload-3
		//fmt.Println(bEvent)
		event := convertToUserDataEvent(bEvent)
		go func() {
			metrics.M_Coin_UserDataWatch_Latency_Summary.WithLabelValues(string(Binance)).Observe(float64(event.Time - uint64(time.Now().UnixMilli())))
			metrics.M_Coin_UserDataWatch_Latency_Histogram.WithLabelValues(string(Binance)).Observe(float64(event.Time - uint64(time.Now().UnixMilli())))
		}()
		eventC <- event
	}
	errHandler := func(err error) {
		s.lg.Error("failed to fetch account changing messages from binance websocket.", "err", err)
		alerting.NotifyRightNow(err, "error occurred when fetching UserData from binance websocket.")
		DefaultHealthChecker.Declare(BinanceUserDataWatchFeature, HealthStateUnhealthy)
	}
	wsServe, err := WsUserDataServe(s.userStreamListenKey, wsHandler, errHandler)
	if err != nil {
		return nil
	}
	DefaultHealthChecker.Declare(BinanceUserDataWatchFeature, HealthStateHealthy)
	<-wsServe.DoneC()
	DefaultHealthChecker.Declare(BinanceUserDataWatchFeature, HealthStateUnhealthy)
	return nil
}

func (s *AccountManager) keepaliveListenKey(ctx context.Context) {
	group, _ := errgroup.WithContext(ctx)
	group.Go(func() error {
		ticker := time.NewTicker(25 * time.Minute)
		for range ticker.C {
			err := s.client.NewKeepaliveUserStreamService().ListenKey(s.userStreamListenKey).Do(context.Background())
			if err != nil {
				s.lg.Error("failed to keepalive listen key for user stream, will retry 3 minutes later")
				time.Sleep(3 * time.Minute)
				err = s.client.NewKeepaliveUserStreamService().ListenKey(s.userStreamListenKey).Do(context.Background())
				s.lg.Warn("retry keepalive listen key", "err", err)
			}
		}
		return nil
	})
}
