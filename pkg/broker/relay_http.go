package broker

import (
	"context"
	"encoding/json"
	"github.com/Octops/agones-event-broadcaster/pkg/events"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RelayConfig struct {
	OnAddUrl       string
	OnUpdateUrl    string
	OnDelete       string
	WorkerReplicas int
}

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

type RelayHTTP struct {
	logger         *logrus.Entry
	wg             *sync.WaitGroup
	Client         Client
	OnAddURL       []string
	OnUpdateURL    []string
	OnDeleteURL    []string
	onAddQueue     *RequestQueue
	onUpdateQueue  *RequestQueue
	onDeleteQueue  *RequestQueue
	workerReplicas int
}

type Client func(req *http.Request) (*http.Response, error)

// TODO: Implement auth mechanism: BasicAuth
func NewRelayHTTP(logger *logrus.Entry, config RelayConfig, client Client) (*RelayHTTP, error) {
	applyConfigDefaults(&config)

	relay := &RelayHTTP{
		logger: logger,
		wg:     &sync.WaitGroup{},
		Client: client,
		onAddQueue: &RequestQueue{
			Name:  "OnAdd",
			Queue: make(chan *RelayRequest, 1024),
		},
		onUpdateQueue: &RequestQueue{
			Name:  "OnUpdate",
			Queue: make(chan *RelayRequest, 1024),
		},
		onDeleteQueue: &RequestQueue{
			Name:  "OnDelete",
			Queue: make(chan *RelayRequest, 1024),
		},
		workerReplicas: config.WorkerReplicas,
	}

	// TODO: Validate if urls are valid http endpoints.
	relay.OnAddURL = strings.Split(config.OnAddUrl, ",")
	relay.OnUpdateURL = strings.Split(config.OnUpdateUrl, ",")
	relay.OnDeleteURL = strings.Split(config.OnDelete, ",")

	return relay, nil
}

func (r *RelayHTTP) Start(ctx context.Context) error {
	r.StartWorkers(ctx, r.onAddQueue, r.Client)
	r.StartWorkers(ctx, r.onUpdateQueue, r.Client)
	r.StartWorkers(ctx, r.onDeleteQueue, r.Client)

	<-ctx.Done()
	r.logger.Info("stopping Relay HTTP broker")
	r.wg.Wait()

	return nil
}

func (r *RelayHTTP) StartWorkers(ctx context.Context, queue *RequestQueue, client Client) {
	for i := 0; i < r.workerReplicas; i++ {
		r.wg.Add(1)
		id := i + 1
		w := NewWorker(queue.Name+strconv.Itoa(id), queue, client)

		go func() {
			defer r.wg.Done()

			if err := w.Start(ctx); err != nil {
				r.logger.Fatal(errors.Wrap(err, "error starting worker"))
			}
		}()
	}
}

// Called by the Broadcaster and builds the envelope that will be send as argument to the SendMessage function
func (r *RelayHTTP) BuildEnvelope(event events.Event) (*events.Envelope, error) {
	envelope := &events.Envelope{}

	envelope.AddHeader("event_source", event.EventSource().String())
	envelope.AddHeader("event_type", event.EventType().String())
	envelope.Message = event.(events.Message)

	return envelope, nil
}

// Called by the Broadcaster when a new event happens
func (r *RelayHTTP) SendMessage(envelope *events.Envelope) error {
	eventSource, err := getEventSourceHeader(envelope)
	if err != nil {
		return err
	}

	request := &RelayRequest{
		Payload: &Payload{Body: envelope},
	}

	switch events.EventSource(eventSource) {
	case events.EventSourceOnAdd:
		request.Method = http.MethodPost
		request.Endpoints = r.OnAddURL
		return r.EnqueueRequest(r.onAddQueue.Queue, request)
	case events.EventSourceOnUpdate:
		request.Method = http.MethodPut
		request.Endpoints = r.OnUpdateURL
		return r.EnqueueRequest(r.onUpdateQueue.Queue, request)
	case events.EventSourceOnDelete:
		request.Method = http.MethodDelete
		request.Endpoints = r.OnDeleteURL
		return r.EnqueueRequest(r.onDeleteQueue.Queue, request)
	}

	return nil
}

func (r *RelayHTTP) EnqueueRequest(queue chan *RelayRequest, request *RelayRequest) error {
	select {
	case queue <- request:
	case <-time.After(5 * time.Second):
		return errors.New("request could not be enqueued due to timeout")
	}

	return nil
}

func (p *Payload) Read(b []byte) (n int, err error) {
	j, err := json.Marshal(p)
	if err != nil {
		return 0, errors.Wrap(io.ErrUnexpectedEOF, err.Error())
	}

	count := copy(b, j)
	return count, io.EOF
}

func applyConfigDefaults(config *RelayConfig) {
	if config.WorkerReplicas <= 0 {
		config.WorkerReplicas = 1
	}
}

func getEventSourceHeader(envelope *events.Envelope) (string, error) {
	if _, ok := envelope.Header.Headers["event_source"]; !ok {
		return "", errors.New("envelope header does not contain a valid event_source")
	}

	eventSource := envelope.Header.Headers["event_source"]
	return eventSource, nil
}
