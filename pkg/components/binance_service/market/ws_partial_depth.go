package market

import (
	"errors"
	"github.com/adshao/go-binance/v2"
	"jasonzhu.com/coin_labor/core/components/log"
	. "jasonzhu.com/coin_labor/pkg/components/general"
)

const (
	defaultLevels = 5
)

type DepthTopService struct {
	lg     log.Logger
	client *binance.Client
	Symbol string
	Levels int

	ErrorCnt     int
	LastUpdateID int64
	Bids         []*Bid
	Asks         []*Ask
}

func NewDepthTopService(symbol string) DepthTopService {
	return NewDepthTopServiceWithLevels(symbol, defaultLevels)
}

func NewDepthTopServiceWithLevels(symbol string, levels int) DepthTopService {
	s := DepthTopService{
		lg:       log.New("binance.depth_top"),
		client:   GetBinanceClient(),
		Symbol:   symbol,
		Levels:   levels,
		ErrorCnt: 1,
	}
	return s
}

func (s *DepthTopService) TopAsk() (*Ask, error) {
	if len(s.Asks) > 0 {
		return s.Asks[0], nil
	}
	return nil, errors.New("no asks")
}
func (s *DepthTopService) TopBid() (*Bid, error) {
	if len(s.Bids) > 0 {
		return s.Bids[0], nil
	}
	return nil, errors.New("no bids")
}
func (s *DepthTopService) Top() (*Ask, *Bid, error) {
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
func (s *DepthTopService) OK() bool {
	return s.ErrorCnt == 0
}

func (s *DepthTopService) Watch() (chan struct{}, chan struct{}, error) {
	wsDepthHandler := func(event *binance.WsPartialDepthEvent) {
		var bids []*Bid
		var asks []*Ask
		for _, item := range event.Bids {
			bid, err := NewFromString(item.Price, item.Quantity)
			if err != nil {
				s.lg.Error("failed to parse string of bid to float", "err", err)
			} else {
				bids = append(bids, &bid)
			}
		}
		for _, item := range event.Asks {
			ask, err := NewFromString(item.Price, item.Quantity)
			if err != nil {
				s.lg.Error("failed to parse string of ask to float", "err", err)
			} else {
				asks = append(asks, &ask)
			}
		}
		s.LastUpdateID = event.LastUpdateID
		s.Bids = bids
		s.Asks = asks
		s.ErrorCnt = 0
		s.lg.Debug("watch top depth with updating", "LastUpdateID", s.LastUpdateID, "len(bid)", len(bids), "len(ask)", len(asks))
	}
	errHandler := func(err error) {
		s.ErrorCnt++
		s.lg.Error("failed to fetch new message from binance websocket.", "err", err)
	}
	doneC, stopC, err := binance.WsPartialDepthServe100Ms(s.Symbol, "5", wsDepthHandler, errHandler)
	if err != nil {
		return nil, nil, err
	}
	return doneC, stopC, nil
}
