package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	// DefaultEndpoint contains endpoint URL of FCM service.
	DefaultEndpoint = "https://fcm.googleapis.com/v1/projects/%s/messages:send"

	// DefaultTimeout duration in second
	DefaultTimeout time.Duration = 30 * time.Second
)

// Client abstracts the interaction between the application server and the
// FCM server via HTTP protocol. The developer must obtain an API key from the
// Google APIs Console page and pass it to the `Client` so that it can
// perform authorized requests on the application server's behalf.
// To send a message to one or more devices use the Client's Send.
//
// If the `HTTP` field is nil, a zeroed http.Client will be allocated and used
// to send messages.
type Client struct {
	client   *http.Client
	endpoint string
	timeout  time.Duration
}

// NewClient creates new Firebase Cloud Messaging Client based on API key and
// with default endpoint and http client.
func NewClient(projectId string, opts ...Option) (*Client, error) {
	c := &Client{
		endpoint: fmt.Sprintf(DefaultEndpoint, projectId),
		client:   &http.Client{},
		timeout:  DefaultTimeout,
	}
	for _, o := range opts {
		if err := o(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// SendWithContext sends a message to the FCM server without retrying in case of service
// unavailability. A non-nil error is returned if a non-recoverable error
// occurs (i.e. if the response status is not "200 OK").
// Behaves just like regular send, but uses external context.
func (c *Client) SendWithContext(ctx context.Context, accessToken string, msg *NewMessage) (*Response, error) {
	// validate
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// marshal message
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return c.send(ctx, accessToken, data)
}

// Send sends a message to the FCM server without retrying in case of service
// unavailability. A non-nil error is returned if a non-recoverable error
// occurs (i.e. if the response status is not "200 OK").
func (c *Client) Send(msg *NewMessage, accessToken string) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	return c.SendWithContext(ctx, accessToken, msg)
}

// SendWithRetry sends a message to the FCM server with defined number of
// retrying in case of temporary error.
func (c *Client) SendWithRetry(msg *NewMessage, accessToken string, retryAttempts int) (*Response, error) {
	return c.SendWithRetryWithContext(context.Background(), msg, accessToken, retryAttempts)
}

// SendWithRetryWithContext sends a message to the FCM server with defined number of
// retrying in case of temporary error.
// Behaves just like regular SendWithRetry, but uses external context.
func (c *Client) SendWithRetryWithContext(ctx context.Context, msg *NewMessage, accessToken string, retryAttempts int) (*Response, error) {
	// validate
	if err := msg.Validate(); err != nil {
		return nil, err
	}
	// marshal message
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	resp := new(Response)
	err = retry(func() error {
		ctx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		var er error
		resp, er = c.send(ctx, accessToken, data)
		return er
	}, retryAttempts)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// send sends a request.
func (c *Client) send(ctx context.Context, accessToken string, data []byte) (*Response, error) {
	// create request
	req, err := http.NewRequest("POST", c.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	// add headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Content-Type", "application/json")

	// execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, connectionError(err.Error())
	}
	defer resp.Body.Close()

	if err := handleFCMResponse(resp); err != nil {
		return nil, err
	}

	// build return
	response := new(Response)
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}

	return response, nil
}

// handleFCMResponse processes the HTTP response.
// For a 200 OK response, it does nothing.
// For error responses, it decodes the error JSON (which follows the google.rpc.Status format)
// and returns an error containing the status and message.
func handleFCMResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return serverError(fmt.Sprintf("%d error: %s", resp.StatusCode, resp.Status))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode error response: %w", err)
	}

	return response.Err()
}
