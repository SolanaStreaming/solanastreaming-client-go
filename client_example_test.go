package solanastreaming

import (
	"context"
	"fmt"
	"log"
)

func ExampleMain() {
	ctx := context.Background()
	cli := New("ae2b1fca515949e5d54fb22b8ed95575")

	err := cli.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	pairsSub, err := cli.SubscribeNewPairs(ctx, &NewPairSubscribeParams{IncludeLaunchpadTokens: true})
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		ev, err := pairsSub.Receive(ctx)
		if err != nil {
			log.Fatal(err)
			return
		}

		fmt.Printf("%#v\n", ev)
	}
}
