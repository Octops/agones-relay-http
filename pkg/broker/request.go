package broker

import (
	"encoding/json"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/pkg/errors"
	"io"
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
		Payload: &Payload{Body: envelope},
	}
	request.Method = record.Method
	request.Endpoints = record.URL

	return request
}
