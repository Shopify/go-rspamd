# go-rspamd

## Introduction <br/>

go-rspamd is a client library written to help interact with a rspamd instance, via HTTP. go-rspamd facilitates
* Content scanning of emails
* Training rspamd's Bayesian classifier
* Updating rspamd's fuzzy storage by adding or removing messages 
* Analyzing content scanning results 

Refer to rspamd [documentation](https://rspamd.com/doc/) for help configuring and setting up rspamd.

## Usage 

The API is defined [here](https://pkg.go.dev/github.com/Shopify/go-rspamd/v3).

The client helps send emails to all POST endpoints on rspamd. Support for all GET endpoints does not currently exist as they can be accessed through rspamd's web interface, although support may exist in the future. A full list of all endpoints can be found [here](https://rspamd.com/doc/architecture/protocol.html). 

The client supports email formats that implement `io.Reader` or `io.WriterTo`.

### Examples

_Note:_ go-rspamd is geared towards clients that use [context](https://golang.org/pkg/context/). However if you don't, whenever `context.Context` is expected, you can use `context.Background()`.

Import go-rspamd:
```go
import "github.com/Shopify/go-rspamd/v3"
```

Instantiate with the url of your rspamd instance:
```go
 client := rspamd.New("https://my-rspamd.com")
```

Optionally pass in credentials:
```go
 client := rspamd.New("https://my-rspamd.com", rspamd.Credentials("username", "password"))
```

Or instantiate with your own [resty](https://github.com/go-resty/resty) client:
```go
 client := rspamd.NewFromClient(restyClient)
```

Ping your rspamd instance:
```go
pong, _ := client.Ping(ctx)
```

Scan an email from an io.Reader:
```go
eml, _ := os.Open("/path/to/email")
header := http.Header{}
rspamd.SetQueueID(header, "MyQueueId")
req := &rspamd.CheckRequest{
    Message: eml,
    Header:  header,
}
checkResp, _ ::= client.Check(ctx, req)
```

Scan an email from an `io.WriterTo`:
```go
eml := ... // implements io.WriterTo, like *gomail.Message for example
header := http.Header{}
rspamd.SetQueueID(header, "MyQueueId")
req := &rspamd.CheckRequest{
    Message: rspamd.ReaderFromWriterTo(eml),
    Header:  header,
}
checkResp, _ := .Check(ctx, req)
```

Add a message to fuzzy storage, attaching a flag and weight as per [docs](https://rspamd.com/doc/architecture/protocol.html#controller-http-endpoints):
```go
eml := ... // implements io.Reader
header := http.Header{}
rspamd.SetQueueID(header, "MyQueueId")
req := &rspamd.FuzzyRequest{
    Message: eml,
    Header:  header,
    Flag:    1,
    Weight:  2,
}
fuzzyResp, _ := client.FuzzyAdd(ctx, req)
```

## Semantics

### Contributing

We invite contributions to extend the API coverage.

Report a bug: Open an issue  
Fix a bug: Open a pull request

### Versioning

go-rspamd respects the [Semantic Versioning](https://semver.org/) for major/minor/patch versions. 

### License

MIT
