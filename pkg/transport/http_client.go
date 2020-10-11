package transport

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Client struct {
	logger *logrus.Entry
	client *http.Client
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
	}, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
