package rspamd

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func Test_Check(t *testing.T) {
	transport := httpmock.NewMockTransport()
	restyClient := resty.New()
	restyClient.SetTransport(transport)
	client := New("http://rspamdexample.com", Credentials("username", "password"))
	client.client = restyClient

	e1 := &CheckRequest{
		Message: open(t, "./testdata/test1.eml"),
		Header:  NewHeaderConfigurer(nil).QueueID("1").Header(),
	}
	e2 := &CheckRequest{
		Message: open(t, "./testdata/test1.eml"),
		Header:  NewHeaderConfigurer(nil).QueueID("2").Header(),
	}
	e3 := &CheckRequest{
		Message: open(t, "./testdata/test1.eml"),
		Header:  NewHeaderConfigurer(nil).QueueID("3").Header(),
	}

	t.Run("success request (check)", func(t *testing.T) {
		transport.Reset()
		transport.RegisterResponder(http.MethodPost, "/checkv2", func(req *http.Request) (*http.Response, error) {
			_, _ = ioutil.ReadAll(req.Body)
			return httpmock.NewJsonResponse(200, CheckResponse{Score: 1.5})
		})

		resp, err := client.Check(context.Background(), e1)

		require.Nil(t, err)
		require.Equal(t, float64(1.5), resp.Score)
	})

	t.Run("bad status code (check)", func(t *testing.T) {
		transport.Reset()
		transport.RegisterResponder(http.MethodPost, "/checkv2", func(req *http.Request) (*http.Response, error) {
			_, _ = ioutil.ReadAll(req.Body)
			return httpmock.NewJsonResponse(400, CheckResponse{Score: 1.5})
		})

		_, err := client.Check(context.Background(), e2)

		require.Error(t, err)
		require.EqualError(t, err, "Unexpected response code: 400")
	})

	t.Run("http error (check)", func(t *testing.T) {
		transport.Reset()
		transport.RegisterResponder(http.MethodPost, "/checkv2", func(req *http.Request) (*http.Response, error) {
			_, _ = ioutil.ReadAll(req.Body)
			return nil, fmt.Errorf("http error")
		})

		_, err := client.Check(context.Background(), e3)

		require.Error(t, err)
		require.Contains(t, err.Error(), "executing request")
	})
}

func Test_Fuzzy(t *testing.T) {
	transport := httpmock.NewMockTransport()
	restyClient := resty.New()
	restyClient.SetTransport(transport)
	client := New("http://rspamdexample.com", Credentials("username", "password"))
	client.client = restyClient

	e4 := &FuzzyRequest{
		Message: open(t, "./testdata/test1.eml"),
		Flag:    1,
		Weight:  19,
		Header:  NewHeaderConfigurer(nil).QueueID("4").Header(),
	}
	e5 := &FuzzyRequest{
		Message: open(t, "./testdata/test1.eml"),
		Flag:    1,
		Header:  NewHeaderConfigurer(nil).QueueID("5").Header(),
	}

	t.Run("success request (fuzzy del)", func(t *testing.T) {
		transport.Reset()
		transport.RegisterResponder(http.MethodPost, "/fuzzydel", func(req *http.Request) (*http.Response, error) {
			_, _ = ioutil.ReadAll(req.Body)
			return httpmock.NewJsonResponse(200, FuzzyResponse{Success: true})
		})

		resp, err := client.FuzzyDel(context.Background(), e4)

		require.Nil(t, err)
		require.Equal(t, true, resp.Success)
	})

	t.Run("bad status code (fuzzy add)", func(t *testing.T) {
		transport.Reset()
		transport.RegisterResponder(http.MethodPost, "/fuzzyadd", func(req *http.Request) (*http.Response, error) {
			_, _ = ioutil.ReadAll(req.Body)
			return httpmock.NewJsonResponse(400, FuzzyResponse{Success: false})
		})

		_, err := client.FuzzyAdd(context.Background(), e5)

		require.Error(t, err)
		require.EqualError(t, err, "Unexpected response code: 400")
	})
}

func Test_IsAlreadyLearnedError(t *testing.T) {
	transport := httpmock.NewMockTransport()
	restyClient := resty.New()
	restyClient.SetTransport(transport)
	client := New("http://rspamdexample.com", Credentials("username", "password"))
	client.client = restyClient

	e6 := &LearnRequest{
		Message: open(t, "./testdata/test1.eml"),
		Header:  NewHeaderConfigurer(nil).QueueID("6").Header(),
	}
	e7 := &LearnRequest{
		Message: open(t, "./testdata/test1.eml"),
		Header:  NewHeaderConfigurer(nil).QueueID("7").Header(),
	}

	t.Run("true if return status is 208", func(t *testing.T) {
		transport.Reset()
		transport.RegisterResponder(http.MethodPost, "/learnspam", func(req *http.Request) (*http.Response, error) {
			_, _ = ioutil.ReadAll(req.Body)
			return httpmock.NewJsonResponse(208, struct {
				ErrorField string `json:"error"`
			}{ErrorField: "<EmailId> has been already learned as spam, ignore it"})
		})

		resp, err := client.LearnSpam(context.Background(), e6)

		require.Equal(t, false, resp.Success)
		require.Equal(t, true, IsAlreadyLearnedError(err))
	})

	t.Run("false if return status is 400", func(t *testing.T) {
		transport.Reset()
		transport.RegisterResponder(http.MethodPost, "/learnspam", func(req *http.Request) (*http.Response, error) {
			_, _ = ioutil.ReadAll(req.Body)
			return httpmock.NewJsonResponse(400, struct {
				ErrorField string `json:"error"`
			}{ErrorField: "error"})
		})

		resp, err := client.LearnSpam(context.Background(), e7)

		require.Equal(t, false, resp.Success)
		require.Equal(t, false, IsAlreadyLearnedError(err))
	})
}

func open(t *testing.T, path string) io.Reader {
	f, err := os.Open(path)
	require.NoError(t, err)
	return f
}
