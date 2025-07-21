package solanastreaming

import (
	"context"
	"encoding/json"
)

type LatestBlockSubscription struct {
	sub subscription[LatestBlockNotification]
}

func (c *Client) SubscribeLatestBlock(ctx context.Context, params *NewPairSubscribeParams) (*LatestBlockSubscription, error) {

	var input *json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		input = (*json.RawMessage)(&data)
	}

	subscriptionID, receiver, err := c.subscribe(ctx, "latestBlockSubscribe", input)
	if err != nil {
		return nil, err
	}

	return &LatestBlockSubscription{
		sub: subscription[LatestBlockNotification]{
			ID:       subscriptionID,
			messages: receiver,
			client:   c,
		},
	}, nil
}

func (s *LatestBlockSubscription) Receive(ctx context.Context) (LatestBlockNotification, error) {
	return receive[LatestBlockNotification](ctx, s.sub)
}

func (s *LatestBlockSubscription) Unsubscribe(ctx context.Context) error {
	return unsubscribe[LatestBlockNotification](ctx, s.sub, "latestBlockUnsubscribe")
}
