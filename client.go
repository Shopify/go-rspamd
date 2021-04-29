package rspamd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	checkV2Endpoint   = "checkv2"
	fuzzyAddEndpoint  = "fuzzyadd"
	fuzzyDelEndpoint  = "fuzzydel"
	learnSpamEndpoint = "learnspam"
	learnHamEndpoint  = "learnham"
	pingEndpoint      = "ping"
)

type Client interface {
	Check(context.Context, *Email) (*CheckResponse, error)
	LearnSpam(context.Context, *Email) (*LearnResponse, error)
	LearnHam(context.Context, *Email) (*LearnResponse, error)
	FuzzyAdd(context.Context, *Email) (*LearnResponse, error)
	FuzzyDel(context.Context, *Email) (*LearnResponse, error)
	Ping(context.Context) (PingResponse, error)
}

type client struct {
	client *resty.Client
}

var _ Client = &client{}

type CheckResponse struct {
	Score     float64               `json:"score"`
	MessageID string                `json:"message-id"`
	Symbols   map[string]SymbolData `json:"symbols"`
}

type LearnResponse struct {
	Success bool `json:"success"`
}

type PingResponse string

// Option is a function that configures the rspamd client.
type Option func(*client) error

type errUnexpectedResponse struct {
	Status int
}

func New(url string, options ...Option) *client {
	client := &client{
		client: resty.New().SetHostURL(url),
	}

	for _, option := range options {
		err := option(client)
		if err != nil {
			log.Fatal("failed to configure client")
		}
	}

	return client
}

func (c *client) Check(ctx context.Context, e *Email) (*CheckResponse, error) {
	result := &CheckResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, checkV2Endpoint)
	return result, err
}

func (c *client) LearnSpam(ctx context.Context, e *Email) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, learnSpamEndpoint)
	return result, err
}

func (c *client) LearnHam(ctx context.Context, e *Email) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, learnHamEndpoint)
	return result, err
}

func (c *client) FuzzyAdd(ctx context.Context, e *Email) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, fuzzyAddEndpoint)
	return result, err
}

func (c *client) FuzzyDel(ctx context.Context, e *Email) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.makeEmailRequest(ctx, e).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, fuzzyDelEndpoint)
	return result, err
}

func (c *client) Ping(ctx context.Context) (PingResponse, error) {
	var result PingResponse
	_, err := c.sendRequest(c.client.R().SetContext(ctx).SetResult(result), resty.MethodGet, pingEndpoint)
	return result, err
}

func (c *client) makeEmailRequest(ctx context.Context, e *Email) *resty.Request {
	headers := map[string]string{}
	if e.queueID != "" {
		headers["Queue-ID"] = e.queueID
	}
	if e.options.flag != 0 {
		headers["Flag"] = fmt.Sprintf("%d", e.options.flag)
	}
	if e.options.weight != 0.0 {
		headers["Weight"] = fmt.Sprintf("%f", e.options.weight)
	}
	return c.client.R().
		SetContext(ctx).
		SetHeaders(headers).
		SetBody(e.message)
}

func (c *client) sendRequest(req *resty.Request, method, url string) (*resty.Response, error) {
	res, err := req.Execute(method, url)

	if err != nil {
		return nil, fmt.Errorf("executing request: %q", err)
	}
	if res.StatusCode() != http.StatusOK {
		return nil, &errUnexpectedResponse{Status: res.StatusCode()}
	}

	return res, nil
}

// Credentials sets the credentials passed in parameters.
func Credentials(username string, password string) Option {
	return func(c *client) error {
		c.client.SetBasicAuth(username, password).SetHeader("User", username)
		return nil
	}
}

func (e *errUnexpectedResponse) Error() string {
	return fmt.Sprintf("Unexpected response code: %d", e.Status)
}

func IsNotFound(err error) bool {
	var errResp *errUnexpectedResponse
	return errors.As(err, &errResp) && errResp.Status == http.StatusNotFound
}
