package general

import "errors"

type DepthInfo struct {
	Symbol          Symbol
	Time            int64
	TransactionTime int64
	LastUpdateID    int64
	Bids            []*Bid
	Asks            []*Ask
	Err             error
}

func NewDepthInfoWithErr(symbol Symbol, err error) *DepthInfo {
	return &DepthInfo{
		Symbol: symbol,
		Err:    err,
	}
}

func (s *DepthInfo) TopAsk() (*Ask, error) {
	if len(s.Asks) > 0 {
		return s.Asks[0], nil
	}
	return nil, errors.New("no asks")
}
func (s *DepthInfo) TopBid() (*Bid, error) {
	if len(s.Bids) > 0 {
		return s.Bids[0], nil
	}
	return nil, errors.New("no bids")
}
func (s *DepthInfo) Top() (*Ask, *Bid, error) {
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
func (s *DepthInfo) TopN(depth int) (*Ask, *Bid, error) {
	if len(s.Asks) >= depth && len(s.Bids) >= depth {
		askN := s.Asks[depth-1]
		bidN := s.Bids[depth-1]
		return askN, bidN, nil
	}
	return nil, nil, errors.New("no depth info")
}
