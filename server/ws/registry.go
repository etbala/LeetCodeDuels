package ws

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	connMu       sync.RWMutex
	connChannels = make(map[string]chan<- []byte)
	ps           *redis.PubSub
)

func registerConn(connID string, ch chan<- []byte) {
	connMu.Lock()
	connChannels[connID] = ch
	connMu.Unlock()
}

func unregisterConn(connID string) {
	connMu.Lock()
	delete(connChannels, connID)
	connMu.Unlock()
}

// lookupSendCh returns the channel for a given connID, or false if not found.
func lookupSendCh(connID string) (chan<- []byte, bool) {
	connMu.RLock()
	ch, ok := connChannels[connID]
	connMu.RUnlock()
	return ch, ok
}

func InitPubSub() {
	ps = ConnManager.client.PSubscribe(context.Background(), "ws:instance:*", "disconnect")
	go pubSubRoutine(ps)
}

func pubSubRoutine(pubsub *redis.PubSub) {
	ch := pubsub.Channel()
	for msg := range ch {
		channel := msg.Channel
		switch {
		case channel == "disconnect":
			var dm disconnectMsg
			if err := json.Unmarshal([]byte(msg.Payload), &dm); err == nil {
				if sendCh, ok := lookupSendCh(dm.ConnID); ok {
					// tell the other instance to tear down
					// e.g. close the sendCh so WriteLoop ends
					close(sendCh)
				}
			}
		case strings.HasPrefix(channel, "ws:instance:"):
			connID := strings.TrimPrefix(channel, "ws:instance:")
			if sendCh, ok := lookupSendCh(connID); ok {
				// forward the raw payload bytes into the WriteLoop
				sendCh <- []byte(msg.Payload)
			}
		}
	}
}

func CleanupPubSub(ctx context.Context) error {
	if ps != nil {
		// Unsubscribe from patterns (PUnsubscribe)
		if err := ps.PUnsubscribe(ctx); err != nil {
			return err
		}
		// and from any normal channel subscriptions (Unsubscribe)
		if err := ps.Unsubscribe(ctx); err != nil {
			return err
		}
		// Only now close the PubSub, which will close the Channel() and make your for-range exit
		if err := ps.Close(); err != nil {
			return err
		}
	}

	// Clean up any in-flight per-connection channels
	connMu.Lock()
	for id, ch := range connChannels {
		close(ch)
		delete(connChannels, id)
	}
	connMu.Unlock()
	return nil
}
