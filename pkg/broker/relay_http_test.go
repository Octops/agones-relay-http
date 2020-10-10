package broker

import (
	"context"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/Octops/agones-relay-http/internal/runtime"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRelayHTTP_SendMessage(t *testing.T) {
	t.Run("it should send a message", func(t *testing.T) {
		logger := runtime.NewLogger(true)

		relay, err := NewRelayHTTP(logger, RelayConfig{
			OnAddUrl:    "http://localhost:8090/add",
			OnUpdateUrl: "http://localhost:8090/update",
			OnDelete:    "http://localhost:8090/delete",
		}, FakeClient)

		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())

		go relay.Start(ctx)

		envelope := &events.Envelope{
			Header: &events.Header{
				Headers: map[string]string{
					"event_source": events.EventSourceOnAdd.String(),
					"event_type":   events.GameServerEventAdded.String(),
				},
			},
		}

		err = relay.SendMessage(envelope)
		require.NoError(t, err)

		envelope = &events.Envelope{
			Header: &events.Header{
				Headers: map[string]string{
					"event_source": events.EventSourceOnUpdate.String(),
					"event_type":   events.GameServerEventUpdated.String(),
				},
			},
		}

		err = relay.SendMessage(envelope)
		require.NoError(t, err)

		time.Sleep(time.Second * 5)
		cancel()
	})
}
