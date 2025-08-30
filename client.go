// Package solanastreaming provides a client library for connecting to the SolanaStreaming websocket api.
// It includes functions for subscribing to Swaps, New Pairs and the latest blocks processed
//
// Go to https://solanastreaming.com for an api key and https://solanastreaming.com/docs for more documentation.
package solanastreaming

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Client struct {
	apiKey     string
	host       string
	log        *logrus.Logger
	conn       *websocket.Conn
	generalErr error
	lock       sync.Mutex // for writing to the same connection
	receivers  map[receiver]chan *wireMessage
}

// New creates a new client instance.
func New(apiKey string) *Client {
	logger := logrus.New()
	logger.SetLevel(logrus.PanicLevel) // default to panic level, can be changed later
	return &Client{
		apiKey:    apiKey,
		host:      "wss://api.solanastreaming.com",
		log:       logger,
		receivers: make(map[receiver]chan *wireMessage),
	}
}

func (o *Client) SetLogLevel(level logrus.Level) {
	o.log.SetLevel(level)
}
func (o *Client) SetHost(host string) {
	o.host = host
}
func (o *Client) SetLogger(logger *logrus.Logger) {
	o.log = logger
}

func (o *Client) Close() error {
	if o.conn != nil {
		return o.conn.Close()
	}
	return nil
}

// Connect establishes a WebSocket connection to the Solana Streaming API and should always be called before any other methods.
func (o *Client) Connect(ctx context.Context) error {
	o.generalErr = nil
	conn, resp, err := websocket.DefaultDialer.Dial(o.host, http.Header{
		"X-API-KEY":  []string{o.apiKey},
		"User-Agent": []string{"solanastreaming-client-go"},
	})
	if err != nil {
		var reason []byte
		if resp != nil {
			reason, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusTooManyRequests {
				err = errors.Wrapf(ErrRateLimitExceeded, err.Error())
			}
		}
		o.log.Errorf("wss dial: %s %s", err.Error(), string(reason))
		return errors.Wrap(err, string(reason))
	}
	o.conn = conn

	go o.receiveMessages()
	return nil
}

func (o *Client) receiveMessages() {
	for {
		_, message, err := o.conn.ReadMessage()
		if err != nil {
			// cant receive so reconnect (can be triggered by set read deadline)
			o.log.Errorf("wss read: %s", err.Error())
			// set generat eerr is not already set (to be returned to receivers or subscribe calls)
			if o.generalErr == nil {
				o.generalErr = err
			}
			return
		}
		o.log.Debugf("WSS_RECEIVE: %s", string(message))

		var event wireMessage
		err = json.Unmarshal(message, &event)
		if err != nil {
			o.log.Errorf("wss unmarshal: %s", err.Error())
			continue
		}

		// response to jsonID or subscription_id
		o.lock.Lock()
		for v := range o.receivers {
			ch := o.receivers[v]
			// this is an error message, send to firt receiver and exit
			if event.ID == 0 && event.Error != nil && event.Error.Message != "" && ch != nil {
				ch <- &event
				o.lock.Unlock()
				return
			}
			if v.Type == receiverTypeBySubscriptionID && event.SubscriptionID == uint(v.Value) {
				ch <- &event
				break
			}
			if v.Type == receiverTypeByRequestID && event.ID == v.Value {
				ch <- &event
				break
			}
		}
		o.lock.Unlock()
	}
}

// send a message over the wire without waiting for a response
func (o *Client) sendMessage(msg wireMessage) error {
	o.lock.Lock()
	defer o.lock.Unlock()
	if o.conn == nil {
		return ErrConnectFirst
	}
	d, err := json.Marshal(msg)
	if err != nil {
		o.log.Errorf("wss marshal: %s", err.Error())
		return err
	}
	o.log.Debugf("WSS_SEND: %s", string(d))
	err = o.conn.WriteMessage(websocket.TextMessage, d)
	if err != nil {
		o.log.Errorf("wss write: %s", err.Error())
		return err
	}
	return nil
}

// send a message over the wire and wait for a response
func (o *Client) sendSyncMessage(ctx context.Context, msg wireMessage) (*wireMessage, error) {
	requestID := randRequestID()
	msg.ID = requestID

	// register response receiver
	response := make(chan *wireMessage)
	o.lock.Lock()
	receiverKey := receiver{
		Type:  receiverTypeByRequestID,
		Value: requestID,
	}
	o.receivers[receiverKey] = response
	o.lock.Unlock()

	// remove after to reduce chance of memory leak
	defer func() {
		o.lock.Lock()
		delete(o.receivers, receiverKey)
		o.lock.Unlock()
	}()

	err := o.sendMessage(msg)
	if err != nil {
		return nil, err
	}

	// ensure timeout if we havent received a response
	timeout := time.After(5 * time.Second)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timeout:
		o.log.Errorf("wss timeout: %d", requestID)
		return nil, errors.New("timeout")
	case val := <-response:
		return val, nil
	}
}

// func (c *Client) decodeLoop(rawCh chan json.RawMessage, method string, outCh interface{}) {
// 	for raw := range rawCh {
// 		var wrapper struct {
// 			Method string          `json:"method"`
// 			Params json.RawMessage `json:"params"`
// 		}
// 		if json.Unmarshal(raw, &wrapper) != nil || wrapper.Method != method {
// 			continue
// 		}
// 		var notif struct {
// 			Params json.RawMessage `json:"params"`
// 		}
// 		json.Unmarshal(wrapper.Params, &notif)

// 		targetCh := outCh.(chan<- any)
// 		var evt any
// 		json.Unmarshal(notif.Params, &evt)
// 		targetCh <- evt
// 	}
// 	close(outCh.(chan any))
// }

func randRequestID() int {
	requestID, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return int(requestID.Int64()) + 1
}
