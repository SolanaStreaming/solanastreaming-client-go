package solanastreaming

import (
	"context"
	"encoding/json"
)

type NewPairSubscribeParams struct {
	IncludePumpfun          bool `json:"include_pumpfun"`
	IncludeRaydiumLaunchlab bool `json:"include_raydium_launchlab"`
}

type NewPairsSubscription struct {
	sub subscription[NewPairNotification]
}

func (c *Client) SubscribeNewPairs(ctx context.Context, params *NewPairSubscribeParams) (*NewPairsSubscription, error) {

	var input *json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		input = (*json.RawMessage)(&data)
	}

	subscriptionID, receiver, err := c.subscribe(ctx, "newPairSubscribe", input)
	if err != nil {
		return nil, err
	}

	return &NewPairsSubscription{
		sub: subscription[NewPairNotification]{
			ID:       subscriptionID,
			messages: receiver,
			client:   c,
		},
	}, nil
}

func (s *NewPairsSubscription) Receive(ctx context.Context) (NewPairNotification, error) {
	if s == nil {
		return NewPairNotification{}, ErrNoSubscription
	}
	return receive[NewPairNotification](ctx, s.sub)
}

func (s *NewPairsSubscription) Unsubscribe(ctx context.Context) error {
	return unsubscribe[NewPairNotification](ctx, s.sub, "newPairUnsubscribe")
}

func (s *NewPairsSubscription) UpdateParams(ctx context.Context, params *NewPairSubscribeParams) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return updateParams[NewPairNotification](ctx, s.sub, (*json.RawMessage)(&data))
}
