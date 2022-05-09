package rspamd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

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

// Client is a rspamd HTTP client.
type Client interface {
	Check(context.Context, *CheckRequest) (*CheckResponse, error)
	LearnSpam(context.Context, *LearnRequest) (*LearnResponse, error)
	LearnHam(context.Context, *LearnRequest) (*LearnResponse, error)
	FuzzyAdd(context.Context, *FuzzyRequest) (*FuzzyResponse, error)
	FuzzyDel(context.Context, *FuzzyRequest) (*FuzzyResponse, error)
	Ping(context.Context) (PingResponse, error)
}

type client struct {
	client *resty.Client
}

var _ Client = &client{}

type Headers map[string]string

func (h Headers) QueueId(queueId string) Headers {
	h["Queue-Id"] = queueId
	return h
}

func (h Headers) flag(flag int) Headers {
	h["Flag"] = strconv.Itoa(flag)
	return h
}

func (h Headers) weight(weight int) Headers {
	h["Weight"] = strconv.Itoa(weight)
	return h
}

// CheckRequest encapsulates the request of Check.
type CheckRequest struct {
	Message io.Reader
	Headers Headers
}

// SymbolData encapsulates the data returned for each symbol from Check.
type SymbolData struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	MetricScore float64 `json:"metric_score"`
	Description string  `json:"description"`
}

// CheckResponse encapsulates the response of Check.
type CheckResponse struct {
	Score     float64               `json:"score"`
	Action    string                `json:"action"`
	MessageID string                `json:"message-id"`
	Symbols   map[string]SymbolData `json:"symbols"`
}

// LearnRequest encapsulates the request of LearnSpam, LearnHam.
type LearnRequest struct {
	Message io.Reader
	Headers Headers
}

// LearnResponse encapsulates the response of LearnSpam, LearnHam.
type LearnResponse struct {
	Success bool `json:"success"`
}

// FuzzyRequest encapsulates the request of FuzzyAdd, FuzzyDel.
type FuzzyRequest struct {
	Message io.Reader
	Flag    int
	Weight  int
	Headers Headers
}

// FuzzyResponse encapsulates the response of FuzzyAdd, FuzzyDel.
type FuzzyResponse struct {
	Success bool     `json:"success"`
	Hashes  []string `json:"hashes"`
}

// PingResponse encapsulates the response of Ping.
type PingResponse string

// Option is a function that configures the rspamd client.
type Option func(*client) error

type UnexpectedResponseError struct {
	Status int
}

// New returns a client.
// It takes the url of a rspamd instance, and configures the client with Options which are closures.
func New(url string, options ...Option) *client {
	cl := NewFromClient(resty.New().SetHostURL(url))

	for _, option := range options {
		err := option(cl)
		if err != nil {
			log.Fatal("failed to configure cl")
		}
	}

	return cl
}

// NewFromClient returns a client.
// It takes an instance of resty.Client.
func NewFromClient(restyClient *resty.Client) *client {
	cl := &client{
		client: restyClient,
	}
	return cl
}

// Check scans an email, returning a spam score and list of symbols.
func (c *client) Check(ctx context.Context, cr *CheckRequest) (*CheckResponse, error) {
	result := &CheckResponse{}
	req := c.buildRequest(ctx, cr.Message, cr.Headers).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, checkV2Endpoint)
	return result, err
}

// LearnSpam trains rspamd's Bayesian classifier by marking an email as spam.
func (c *client) LearnSpam(ctx context.Context, lr *LearnRequest) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.buildRequest(ctx, lr.Message, lr.Headers).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, learnSpamEndpoint)
	return result, err
}

// LearnHam trains rspamd's Bayesian classifier by marking an email as ham.
func (c *client) LearnHam(ctx context.Context, lr *LearnRequest) (*LearnResponse, error) {
	result := &LearnResponse{}
	req := c.buildRequest(ctx, lr.Message, lr.Headers).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, learnHamEndpoint)
	return result, err
}

// FuzzyAdd adds an email to fuzzy storage.
func (c *client) FuzzyAdd(ctx context.Context, fr *FuzzyRequest) (*FuzzyResponse, error) {
	result := &FuzzyResponse{}
	if fr.Headers == nil {
		fr.Headers = Headers{}
	}
	req := c.buildRequest(ctx, fr.Message, fr.Headers.flag(fr.Flag).weight(fr.Weight)).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, fuzzyAddEndpoint)
	return result, err
}

// FuzzyDel removes an email from fuzzy storage.
func (c *client) FuzzyDel(ctx context.Context, fr *FuzzyRequest) (*FuzzyResponse, error) {
	result := &FuzzyResponse{}
	if fr.Headers == nil {
		fr.Headers = Headers{}
	}
	req := c.buildRequest(ctx, fr.Message, fr.Headers.flag(fr.Flag)).SetResult(result)
	_, err := c.sendRequest(req, resty.MethodPost, fuzzyDelEndpoint)
	return result, err
}

// Ping pings the client's rspamd instance.
func (c *client) Ping(ctx context.Context) (PingResponse, error) {
	var result PingResponse
	_, err := c.sendRequest(c.client.R().SetContext(ctx).SetResult(result), resty.MethodGet, pingEndpoint)
	return result, err
}

func (c *client) buildRequest(ctx context.Context, message io.Reader, headers Headers) *resty.Request {
	return c.client.R().
		SetContext(ctx).
		SetHeaders(headers).
		SetBody(message)
}

func (c *client) sendRequest(req *resty.Request, method, endpoint string) (*resty.Response, error) {
	res, err := req.Execute(method, endpoint)

	if err != nil {
		return nil, fmt.Errorf("executing request: %q", err)
	}
	if res.StatusCode() != http.StatusOK {
		return nil, &UnexpectedResponseError{Status: res.StatusCode()}
	}

	return res, nil
}

// Credentials sets the credentials passed in parameters.
// It returns an Option which is used to configure the client.
func Credentials(username string, password string) Option {
	return func(c *client) error {
		c.client.SetBasicAuth(username, password).SetHeader("User", username)
		return nil
	}
}

func (e *UnexpectedResponseError) Error() string {
	return fmt.Sprintf("Unexpected response code: %d", e.Status)
}

// IsNotFound returns true if a request returned a 404. This helps discern a known issue with rspamd's /checkv2 endpoint.
func IsNotFound(err error) bool {
	var errResp *UnexpectedResponseError
	return errors.As(err, &errResp) && errResp.Status == http.StatusNotFound
}

// IsAlreadyLearnedError returns true if a request returns 208, which can happen if rspamd detects a message has already been learned as SPAM/HAM.
// This can allow clients to gracefully handle this use case.
func IsAlreadyLearnedError(err error) bool {
	var errResp *UnexpectedResponseError
	return errors.As(err, &errResp) && errResp.Status == http.StatusAlreadyReported
}

func ReaderFromWriterTo(writerTo io.WriterTo) io.Reader {
	r, w := io.Pipe()

	go func() {
		if _, err := writerTo.WriteTo(w); err != nil {
			_ = w.CloseWithError(fmt.Errorf("writing to pipe: %q", err))
			return
		}

		_ = w.Close() // Always succeeds
	}()

	return r
}
