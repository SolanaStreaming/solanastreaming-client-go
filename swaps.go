package solanastreaming

import (
	"context"
	"encoding/json"

	"github.com/gagliardetto/solana-go"
)

type SwapSubscribeParams struct {
	Include FilterFields `json:"include"`
}

type FilterFields struct {
	AmmAccount    []solana.PublicKey
	WalletAccount []solana.PublicKey
	BaseTokenMint []solana.PublicKey
	USDValue      *float64
}

type SwapsSubscription struct {
	sub subscription[SwapNotification]
}

func (c *Client) SubscribeSwaps(ctx context.Context, params *SwapSubscribeParams) (*SwapsSubscription, error) {

	var input *json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		input = (*json.RawMessage)(&data)
	}

	subscriptionID, receiver, err := c.subscribe(ctx, "swapSubscribe", input)
	if err != nil {
		return nil, err
	}

	return &SwapsSubscription{
		sub: subscription[SwapNotification]{
			ID:       subscriptionID,
			messages: receiver,
			client:   c,
		},
	}, nil
}

func (s *SwapsSubscription) Receive(ctx context.Context) (SwapNotification, error) {
	if s == nil {
		return SwapNotification{}, ErrNoSubscription
	}
	return receive[SwapNotification](ctx, s.sub)
}

func (s *SwapsSubscription) Unsubscribe(ctx context.Context) error {
	return unsubscribe[SwapNotification](ctx, s.sub, "swapUnsubscribe")
}

func (s *SwapsSubscription) UpdateParams(ctx context.Context, params *SwapSubscribeParams) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return updateParams[SwapNotification](ctx, s.sub, (*json.RawMessage)(&data))
}
