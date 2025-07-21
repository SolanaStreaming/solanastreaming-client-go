package solanastreaming

import (
	"context"
	"encoding/json"
	"fmt"
)

type subscription[T any] struct {
	ID       uint
	messages chan *wireMessage
	err      chan error
	client   *Client
}

func receive[T any](ctx context.Context, sub subscription[T]) (T, error) {
	var value T
	if sub.client.generalErr != nil {
		return value, sub.client.generalErr
	}
	select {
	case <-ctx.Done():
		return value, ctx.Err()
	case err := <-sub.err:
		return value, err
	case v := <-sub.messages:
		if v.Params == nil {
			return value, fmt.Errorf("received nil params in message: %v", v)
		}
		err := json.Unmarshal(*v.Params, &value)
		if err != nil {
			return value, fmt.Errorf("unmarshal error: %w", err)
		}
		return value, nil
	}
}

func unsubscribe[T any](ctx context.Context, sub subscription[T], method string) error {
	unsubscribeParams := []byte(fmt.Sprintf(`{"subscription_id":%d}`, sub.ID))
	response, err := sub.client.sendSyncMessage(ctx, wireMessage{
		Method: method,
		Params: (*json.RawMessage)(&unsubscribeParams),
	})
	if err != nil {
		return err
	}

	// could not subscribe
	if response.Error != nil && response.Error.Code != 0 {
		// o.log.Errorf("solana wss error: %d %s", val.Error.Code, val.Error.Message)
		return fmt.Errorf("solana wss error: %d %s", response.Error.Code, response.Error.Message)
	}

	close(sub.messages)
	close(sub.err)

	return nil
}

func updateParams[T any](ctx context.Context, sub subscription[T], params *json.RawMessage) error {

	updateParams := struct {
		SubscriptionID uint             `json:"subscription_id"`
		Params         *json.RawMessage `json:"params"`
	}{
		SubscriptionID: sub.ID,
		Params:         params,
	}
	data, err := json.Marshal(updateParams)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	response, err := sub.client.sendSyncMessage(ctx, wireMessage{
		Method: "updateSubscriptionParams",
		Params: (*json.RawMessage)(&data),
	})
	if err != nil {
		return err
	}

	// could not subscribe
	if response.Error != nil && response.Error.Code != 0 {
		// o.log.Errorf("solana wss error: %d %s", val.Error.Code, val.Error.Message)
		return fmt.Errorf("solana wss error: %d %s", response.Error.Code, response.Error.Message)
	}

	close(sub.messages)
	close(sub.err)

	return nil
}

func (o *Client) subscribe(ctx context.Context, method string, params *json.RawMessage) (uint, chan *wireMessage, error) {
	if o.generalErr != nil {
		return 0, nil, o.generalErr
	}
	// subscribe to pairs and wait to see if subscription is successful
	response, err := o.sendSyncMessage(ctx, wireMessage{
		Method: method,
		Params: params,
	})
	if err != nil {
		return 0, nil, err
	}

	// could not subscribe
	if response.Error != nil && response.Error.Code != 0 {
		// o.log.Errorf("solana wss error: %d %s", val.Error.Code, val.Error.Message)
		return 0, nil, fmt.Errorf("solana wss error: %d %s", response.Error.Code, response.Error.Message)
	}

	subscribeResponse := struct {
		Message        string `json:"message"`
		SubscriptionID uint   `json:"subscription_id"`
	}{}
	err = json.Unmarshal(response.Result, &subscribeResponse)
	if err != nil {
		return 0, nil, err
	}

	// success: get subscription id from response and retup receiver
	// todo: potential issue here is we setup receive after sending message we could miss a few messages?
	receiverChan := make(chan *wireMessage, 1)
	o.lock.Lock()
	receiverKey := receiver{
		Type:  receiverTypeBySubscriptionID,
		Value: int(subscribeResponse.SubscriptionID),
	}
	o.receivers[receiverKey] = receiverChan
	o.lock.Unlock()

	return subscribeResponse.SubscriptionID, receiverChan, nil
}
