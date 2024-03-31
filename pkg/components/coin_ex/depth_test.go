package coin_ex

import (
	"context"
	"testing"
	"time"
)

var depthService = DepthService{}

func TestDepthService_Depth(t *testing.T) {
	depthService.Init()

	ctx, shutdownFn := context.WithCancel(context.Background())
	_ = depthService.Run(ctx)

	go func() {
		time.Sleep(5 * time.Second)
		shutdownFn()
	}()
	<-ctx.Done()
}
