[![Go Reference](https://pkg.go.dev/badge/github.com/solanastreaming/solanastreaming-client-go.svg)](https://pkg.go.dev/github.com/solanastreaming/solanastreaming-client-go)
[![Go Report Card](https://goreportcard.com/badge/solanastreaming/solanastreaming-client-go)](https://goreportcard.com/report/solanastreaming/solanastreaming-client-go) 
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/solanastreaming/solanastreaming-client-go/blob/master/LICENSE)

SolanaStreaming client library for the websocket api.

This client library provides functions for subscribing to swaps and pairs. 

Example usage:
```
go get github.com/solanastreaming/solanastreaming-client-go
```

Example main.go
```golang
package main

import (
    solanastreaming "github.com/solanastreaming/solanastreaming-client-go"
)

func main() {
    ctx := context.Background()
    cli := solanastreaming.New("ae2b1fca515949e5d54fb22b8ed95575")

    err := cli.Connect(ctx)
    if err != nil {
        log.Fatal(err)
        return
    }

    pairsSub, err := cli.SubscribeNewPairs(ctx, &solanastreaming.NewPairSubscribeParams{
        IncludePumpfun: true,
    })
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
```