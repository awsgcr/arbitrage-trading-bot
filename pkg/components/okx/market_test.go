package okx

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestWsOrderBook(t *testing.T) {
	ctx, shutdownF := context.WithCancel(context.Background())

	c := NewMarketClient("")
	err := c.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				ask, _ := c.TopAsk()
				fmt.Println(ask)
			case <-ctx.Done():
				fmt.Printf("finished.")
				return
			}
		}
	}()

	ticker := time.NewTicker(8 * time.Second)
	<-ticker.C
	shutdownF()

	ticker2 := time.NewTicker(5 * time.Second)
	<-ticker2.C
}
