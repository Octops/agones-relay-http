package transport

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Client struct {
	logger   *logrus.Entry
	client   *http.Client
	retries  int
	interval time.Duration
}

func NewClient(logger *logrus.Entry, timeout string) (*Client, error) {
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, err
	}

	return &Client{
		logger: logger,
		client: &http.Client{
			Timeout: duration,
		},
		retries:  5,
		interval: 5 * time.Second,
	}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	response := &http.Response{}
	fn := func() error {
		resp, err := c.client.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		response = resp
		return nil
	}

	err := withRetry(c.retries, c.interval, fn)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func withRetry(retries int, interval time.Duration, fn func() error) error {
	var err error
	for i := 0; i < retries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		time.Sleep(interval)
	}

	return err
}
