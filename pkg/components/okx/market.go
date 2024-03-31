package okx

import (
	"context"
	"errors"
	"fmt"
	"github.com/amir-the-h/okex"
	"github.com/amir-the-h/okex/api"
	"github.com/amir-the-h/okex/events"
	"github.com/amir-the-h/okex/events/public"
	ws_public_requests "github.com/amir-the-h/okex/requests/ws/public"
	"golang.org/x/sync/errgroup"
	"jasonzhu.com/coin_labor/core/components/log"
	. "jasonzhu.com/coin_labor/pkg/components/general"
	"time"
)

type MarketClient struct {
	lg     log.Logger
	client *api.Client
	group  *errgroup.Group

	Symbol       string
	ErrorCnt     int
	LastUpdateID int64
	Bids         []*Bid
	Asks         []*Ask
}

func NewMarketClient(symbol string) MarketClient {
	c := MarketClient{
		lg:     log.New("okx.market"),
		Symbol: symbol,
	}
	return c
}

func (c *MarketClient) TopAsk() (*Ask, error) {
	if len(c.Asks) > 0 {
		return c.Asks[0], nil
	}
	return nil, errors.New("no asks")
}
func (c *MarketClient) TopBid() (*Bid, error) {
	if len(c.Bids) > 0 {
		return c.Bids[0], nil
	}
	return nil, errors.New("no bids")
}
func (c *MarketClient) Top() (*Ask, *Bid, error) {
	ask, err := c.TopAsk()
	if err != nil {
		return nil, nil, err
	}
	bid, err := c.TopBid()
	if err != nil {
		return nil, nil, err
	}
	return ask, bid, nil
}
func (c *MarketClient) OK() bool {
	return c.ErrorCnt == 0
}

func (c *MarketClient) Run(ctx context.Context) (err error) {
	apiKey := "YOUR-API-KEY"
	secretKey := "YOUR-SECRET-KEY"
	passphrase := "YOUR-PASS-PHRASE"
	dest := okex.NormalServer // The main API server
	//dest := okex.AwsServer // The main API server
	c.client, err = api.NewClient(ctx, apiKey, secretKey, passphrase, dest)
	if err != nil {
		return err
	}

	c.group, _ = errgroup.WithContext(ctx)
	c.initChannels()
	err = c.subscribeOrderBook()
	if err != nil {
		return err
	}
	return nil
}

func (c *MarketClient) initChannels() {
	c.lg.Info("Starting")
	errChan := make(chan *events.Error)
	subChan := make(chan *events.Subscribe)
	uSubChan := make(chan *events.Unsubscribe)
	logChan := make(chan *events.Login)
	sucChan := make(chan *events.Success)
	c.client.Ws.SetChannels(errChan, subChan, uSubChan, logChan, sucChan)

	c.group.Go(func() error {
		for {
			select {
			case <-logChan:
				c.lg.Info("[Authorized]")
			case success := <-sucChan:
				c.lg.Info("[SUCCESS]", "info", success)
			case sub := <-subChan:
				channel, _ := sub.Arg.Get("channel")
				c.lg.Info("[Subscribed]", "channel", channel)
			case uSub := <-uSubChan:
				channel, _ := uSub.Arg.Get("channel")
				c.lg.Info("[Unsubscribed]", "channel", channel)
			case err := <-c.client.Ws.ErrChan:
				c.ErrorCnt++
				c.lg.Error("[Error]", "err", err)
				for _, datum := range err.Data {
					c.lg.Error("[Error]", "datum", datum)
				}
			case b := <-c.client.Ws.DoneChan:
				c.ErrorCnt = 10
				c.lg.Info("[End]", "info", b)
				return nil
			}
		}
	})
}

func (c *MarketClient) subscribeOrderBook() error {
	obCh := make(chan *public.OrderBook)
	err := c.client.Ws.Public.OrderBook(ws_public_requests.OrderBook{
		InstID:  "ETH-USDT",
		Channel: "books5",
	}, obCh)
	if err != nil {
		return err
	}

	c.group.Go(func() error {
		for {
			select {
			case i := <-obCh:
				//ch, _ := i.Arg.Get("channel")
				//fmt.Printf("[Event]\t%s", ch)

				var bids []*Bid
				var asks []*Ask
				for _, p := range i.Books {
					//for i := len(p.Asks) - 1; i >= 0; i-- {
					//	fmt.Printf("\t\tAsk\t%+v", p.Asks[i])
					//}
					for _, bid := range p.Bids {
						//fmt.Printf("\t\tBid\t%+v", bid)
						b := Bid{Price: bid.DepthPrice, Quantity: bid.Size}
						bids = append(bids, &b)
					}
					for _, ask := range p.Asks {
						//fmt.Printf("\t\tBid\t%+v", bid)
						a := Ask{Price: ask.DepthPrice, Quantity: ask.Size}
						asks = append(bids, &a)
					}
					//fmt.Printf("ts", p.TS)
					c.LastUpdateID = (time.Time)(p.TS).UnixMilli()
				}
				c.Bids = bids
				c.Asks = asks
				c.ErrorCnt = 0
				//c.LastUpdateID =
			case b := <-c.client.Ws.DoneChan:
				fmt.Printf("[End]:\t%v", b)
				return nil
			}
		}
	})
	return nil
}

func (c *MarketClient) subscribeTicker() error {
	tickersChan := make(chan *public.Tickers)
	err := c.client.Ws.Public.Tickers(ws_public_requests.Tickers{
		InstID: "ETH-USDT",
	}, tickersChan)
	if err != nil {
		return err
	}

	c.group.Go(func() error {
		for {
			select {
			case t := <-tickersChan:
				ch, _ := t.Arg.Get("channel")
				fmt.Printf("[Event]\t%s", ch)
				for _, p := range t.Tickers {
					fmt.Printf("\t\tTicker\t%+v", p)
					fmt.Printf("%.3f, %.3f, %.3f, %.2f/k", p.BidPx, p.AskPx, p.AskPx-p.BidPx, (p.AskPx-p.BidPx)/p.BidPx*1000)
				}
			case b := <-c.client.Ws.DoneChan:
				fmt.Printf("[End]:\t%v", b)
				return nil
			}
		}
	})
	return nil
}
