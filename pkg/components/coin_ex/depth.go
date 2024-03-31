package coin_ex

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"jasonzhu.com/coin_labor/core/components/log"
	"jasonzhu.com/coin_labor/core/util/http"
	. "jasonzhu.com/coin_labor/pkg/components/general"
	"strconv"
	"time"
)

//https://api.coinex.com/v1/market/depth?market=ethusdt&merge=0.001&limit=10

type DepthService struct {
	lg      log.Logger
	stopped bool

	ErrorCnt     int
	LastUpdateID int64
	Bids         []*Bid
	Asks         []*Ask
}

func (s *DepthService) Init() error {
	s.lg = log.New("coin_ex.depth")
	s.ErrorCnt = 1
	return nil
}

func (s *DepthService) Run(ctx context.Context) error {
	group, _ := errgroup.WithContext(ctx)
	s.stopped = false

	group.Go(func() error {
		sleep := false
		for {
			if sleep {
				time.Sleep(50 * time.Millisecond)
			}
			if s.stopped {
				return nil
			}
			start := time.Now().UnixMilli()
			err := s.Depth(defaultSymbol, 10)
			end := time.Now().UnixMilli()
			//fmt.Println("Time spend", "ms", end-start)
			if end-start > 50 {
				sleep = false
			} else {
				sleep = true
			}
			if err != nil {
				s.lg.Info("Failed to get depth info", "err", err)
				continue
			}
		}
	})
	<-ctx.Done()
	s.stopped = true
	return nil
}

func (s *DepthService) TopAsk() (*Ask, error) {
	if len(s.Asks) > 0 {
		return s.Asks[0], nil
	}
	return nil, errors.New("no asks")
}
func (s *DepthService) TopBid() (*Bid, error) {
	if len(s.Bids) > 0 {
		return s.Bids[0], nil
	}
	return nil, errors.New("no bids")
}
func (s *DepthService) Top() (*Ask, *Bid, error) {
	ask, err := s.TopAsk()
	if err != nil {
		return nil, nil, err
	}
	bid, err := s.TopBid()
	if err != nil {
		return nil, nil, err
	}
	return ask, bid, nil
}
func (s *DepthService) OK() bool {
	return s.ErrorCnt == 0
}

func (s *DepthService) Depth(symbol string, limit int) error {
	params := http.Params{
		"market": symbol,
		"merge":  "0.001",
		"limit":  strconv.Itoa(limit),
	}
	j, err := httpGetData(orderBookEndpoint, params)
	if err != nil {
		s.ErrorCnt++
		return err
	}

	s.LastUpdateID = j.Get("time").MustInt64()
	asksLen := len(j.Get("asks").MustArray())
	asks := make([]*Ask, asksLen)
	for i := 0; i < asksLen; i++ {
		item := j.Get("asks").GetIndex(i)
		ask, err := NewFromString(item.GetIndex(0).MustString(), item.GetIndex(1).MustString())
		if err != nil {
			continue
		}
		asks[i] = &ask
	}
	bidsLen := len(j.Get("bids").MustArray())
	bids := make([]*Bid, bidsLen)
	for i := 0; i < bidsLen; i++ {
		item := j.Get("bids").GetIndex(i)
		bid, err := NewFromString(item.GetIndex(0).MustString(), item.GetIndex(1).MustString())
		if err != nil {
			continue
		}
		bids[i] = &bid
	}

	s.Asks = asks
	s.Bids = bids
	s.ErrorCnt = 0
	return nil
}
