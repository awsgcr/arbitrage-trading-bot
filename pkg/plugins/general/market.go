package general

import (
	"context"
	"errors"
)

const DefaultLimit = 10

type GMarketManager struct {
	fetchDepthFn   func(symbol Symbol, limit int) *DepthInfo
	wsWatchDepthFn func(ctx context.Context, infoC chan *DepthInfo, limit int, symbols ...Symbol) error

	Watching           bool
	WatchingSymbols    []Symbol
	WatchingDepthLimit int
}

func InitGMarketManager(fetchDepthFn func(symbol Symbol, limit int) *DepthInfo,
	wsWatchDepthFn func(ctx context.Context, infoC chan *DepthInfo, limit int, symbols ...Symbol) error,
) GMarketManager {
	return GMarketManager{
		fetchDepthFn:       fetchDepthFn,
		wsWatchDepthFn:     wsWatchDepthFn,
		Watching:           false,
		WatchingDepthLimit: DefaultLimit,
	}
}

// Public Method

func (s *GMarketManager) IsWatching() bool {
	return s.Watching
}

// FetchDepth 同步获取DepthInfo数据，Via API.
func (s *GMarketManager) FetchDepth(symbol Symbol, limit int) (*DepthInfo, error) {
	depthInfo := s.fetchDepthFn(symbol, limit)
	if depthInfo.Err != nil {
		return nil, depthInfo.Err
	}
	return depthInfo, nil
}

func (s *GMarketManager) WsWatchMarketDepth(ctx context.Context, infoC chan *DepthInfo, symbols ...Symbol) error {
	if s.wsWatchDepthFn == nil {
		return errors.New("wsWatchDepthFn is not defined")
	}

	s.WatchingSymbols = symbols
	s.Watching = true
	defer func() {
		s.Watching = false
		s.WatchingSymbols = []Symbol{}
	}()

	return s.wsWatchDepthFn(ctx, infoC, s.WatchingDepthLimit, symbols...)
}

//func (s *GMarketManager) WatchDepth(ctx context.Context, symbols ...Symbol) error {
//	group, _ := errgroup.WithContext(ctx)
//	s.WatchingSymbols = symbols
//	s.Watching = true
//
//	ticker := time.NewTicker(time.Duration(150*len(symbols)) * time.Millisecond)
//	for {
//		select {
//		case <-ticker.C:
//			for _, symbol := range symbols {
//				var sb = symbol
//				group.Go(func() error {
//					err := s.fetchDepth(sb, s.WatchingDepthLimit)
//					if err != nil {
//						fmt.Println(err)
//					}
//					return err
//				})
//			}
//		case <-ctx.Done():
//			s.Watching = false
//			s.WatchingSymbols = []Symbol{}
//			return nil
//		}
//	}
//}
