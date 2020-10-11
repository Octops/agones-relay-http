package broker

import (
	"encoding/json"
	"fmt"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/pkg/errors"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"reflect"
	"strings"
)

type RelayRequest struct {
	Method    string
	Endpoints []string
	Payload   *Payload
}

type Payload struct {
	Body *events.Envelope
}

type RequestQueue struct {
	Name  string
	Queue chan *RelayRequest
}

func (p *Payload) Read(b []byte) (n int, err error) {
	j, err := json.Marshal(p)
	if err != nil {
		return 0, errors.Wrap(io.ErrUnexpectedEOF, err.Error())
	}

	count := copy(b, j)
	return count, io.EOF
}

func createRequest(record *EventRelayRecord, envelope *events.Envelope) *RelayRequest {
	request := &RelayRequest{
		Method: record.Method,
	}

	switch record.Method {
	case http.MethodPost, http.MethodPut:
		request.Payload = &Payload{
			Body: envelope,
		}
		request.Endpoints = record.URL
	case http.MethodDelete:
		request.Endpoints = makeDeleteURLEndpoints(record.URL, envelope)
	}

	return request
}

func makeDeleteURLEndpoints(endpoints []string, envelope *events.Envelope) []string {
	deleteEndpoints := []string{}

	source := getEventSourceFromEnvelope(envelope)

	namespace, name := getResourceKeyFromMessage(envelope.Message.(events.Message))
	for _, ep := range endpoints {
		deleteEndpoints = append(deleteEndpoints, fmt.Sprintf("%s?source=%s&namespace=%s&name=%s", ep, strings.ToLower(source), namespace, name))
	}

	return deleteEndpoints
}

func getEventSourceFromEnvelope(envelope *events.Envelope) string {
	source := reflect.TypeOf(envelope.Message).Elem().String()
	parts := strings.Split(source, ".")
	if len(parts) > 1 {
		source = parts[1]
	}
	return source
}

func getResourceKeyFromMessage(msg events.Message) (string, string) {
	res := msg.Content().(v1.Object)

	namespace := res.GetNamespace()
	name := res.GetName()

	return namespace, name
}
