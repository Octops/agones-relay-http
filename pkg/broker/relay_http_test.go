package broker

import (
	"context"
	"encoding/json"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/Octops/agones-relay-http/internal/runtime"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
)

func TestRelayHTTP_SendMessage(t *testing.T) {
	testCases := []struct {
		name          string
		endpointURL   string
		requestMethod string
		envelope      *events.Envelope
		wantErr       bool
	}{
		{
			name:          "it should send a message for OnAdd event",
			endpointURL:   "http://localhost:8090/add",
			requestMethod: http.MethodPost,
			envelope: &events.Envelope{
				Header: &events.Header{
					Headers: map[string]string{
						"event_source": events.EventSourceOnAdd.String(),
						"event_type":   events.GameServerEventAdded.String(),
					},
				},
				Message: "Payload OnAdd",
			},
			wantErr: false,
		},
		{
			name:          "it should send a message for OnUpdate event",
			endpointURL:   "http://localhost:8090/update",
			requestMethod: http.MethodPut,
			envelope: &events.Envelope{
				Header: &events.Header{
					Headers: map[string]string{
						"event_source": events.EventSourceOnUpdate.String(),
						"event_type":   events.GameServerEventUpdated.String(),
					},
				},
				Message: "Payload OnUpdate",
			},
			wantErr: false,
		},
		{
			name:          "it should send a message for OnDelete event",
			endpointURL:   "http://localhost:8090/delete",
			requestMethod: http.MethodDelete,
			envelope: &events.Envelope{
				Header: &events.Header{
					Headers: map[string]string{
						"event_source": events.EventSourceOnDelete.String(),
						"event_type":   events.GameServerEventDeleted.String(),
					},
				},
				Message: "Payload OnDelete",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := runtime.NewLogger(true)

			wg := sync.WaitGroup{}
			wg.Add(1)

			var response *http.Response
			client := func(req *http.Request) (*http.Response, error) {
				response = &http.Response{
					Status:  "200 OK",
					Body:    req.Body,
					Request: req,
				}
				wg.Done()
				return response, nil
			}

			relay, err := NewRelayHTTP(logger, RelayConfig{
				OnAddUrl:    "http://localhost:8090/add",
				OnUpdateUrl: "http://localhost:8090/update",
				OnDeleteUrl: "http://localhost:8090/delete",
			}, client)

			require.NoError(t, err)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go relay.Start(ctx)

			err = relay.SendMessage(tc.envelope)
			require.NoError(t, err)

			wg.Wait()
			body, err := ioutil.ReadAll(response.Body)
			require.NoError(t, err)

			var payload Payload
			err = json.Unmarshal(body, &payload)
			require.NoError(t, err)
			require.Equal(t, tc.envelope, payload.Body)

			require.Equal(t, tc.requestMethod, response.Request.Method)
			require.Equal(t, tc.endpointURL, response.Request.URL.String())
		})
	}
}
