package solanastreaming

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	cli := New("1e41e2d28d4adec68accd1b120d3eec7")
	cli.Connect(ctx)
	cli.SetLogLevel(logrus.DebugLevel)

	sub, err := cli.SubscribeNewPairs(ctx, &NewPairSubscribeParams{IncludePumpfun: true})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}
	for {
		ev, err := sub.Receive(ctx)
		if err != nil {
			t.Logf("receive error: %v", err)
			return
		}

		fmt.Printf("New Pair Received: %s", ev.Pair.BaseToken.Account)
	}
}
