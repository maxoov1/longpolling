package longpolling

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	_defaultTimeout = 5 * time.Second
)

type (
	Option          func(client *Client)
	ResponseHandler func(response *http.Response, err error)
)

func WithTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		client.timeout = timeout
	}
}

func WithCustomClient(httpClient *http.Client) Option {
	return func(client *Client) {
		client.httpClient = httpClient
	}
}

func WithResponseHandler(responseHandler ResponseHandler) Option {
	return func(client *Client) {
		client.responseHandler = responseHandler
	}
}

func WithMethod(method string) Option {
	return func(client *Client) {
		client.method = method
	}
}

type Client struct {
	httpClient      *http.Client
	ticker          *time.Ticker
	timeout         time.Duration
	resource        string
	method          string
	stopChannel     chan bool
	responseHandler ResponseHandler
}

func NewClient(resource string, options ...Option) *Client {
	client := &Client{
		resource:    resource,
		stopChannel: make(chan bool, 1),
	}

	for _, option := range options {
		option(client)
	}

	if client.httpClient == nil {
		client.httpClient = http.DefaultClient
	}

	if client.timeout == 0 {
		client.timeout = _defaultTimeout
	}
	client.ticker = time.NewTicker(client.timeout)

	return client
}

func (c *Client) makeRequest(
	method string, resource string, body io.Reader,
) (*http.Response, error) {
	request, err := http.NewRequest(method, resource, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	return response, nil
}

func (c *Client) handleTick() {
	if c.responseHandler == nil {
		return
	}

	response, err := c.makeRequest(c.method, c.resource, nil)
	c.responseHandler(response, err)

	if err == nil {
		response.Body.Close()
	}
}

func (c *Client) startPolling() {
	for {
		select {
		case <-c.ticker.C:
			go c.handleTick()

		case <-c.stopChannel:
			return
		}
	}
}

func (c *Client) Start() {
	c.ticker.Reset(c.timeout)
	go c.startPolling()
}

func (c *Client) Stop() {
	c.stopChannel <- true
	close(c.stopChannel)

	c.ticker.Stop()
}
