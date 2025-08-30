package solanastreaming

import (
	"context"
)

type LatestBlockSubscription struct {
	sub subscription[LatestBlockNotification]
}

func (c *Client) SubscribeLatestBlock(ctx context.Context) (*LatestBlockSubscription, error) {
	subscriptionID, receiver, err := c.subscribe(ctx, "latestBlockSubscribe", nil)
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
	if s == nil {
		return LatestBlockNotification{}, ErrNoSubscription
	}
	return receive[LatestBlockNotification](ctx, s.sub)
}

// Unsubscribe from the latest block notifications. To prevent deadlocks, Avoid putting your Unsubscribe() call in your Receive() loop
func (s *LatestBlockSubscription) Unsubscribe(ctx context.Context) error {
	return unsubscribe[LatestBlockNotification](ctx, s.sub, "latestBlockUnsubscribe")
}
