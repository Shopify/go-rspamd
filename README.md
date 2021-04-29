# rspamd-client-go

## Introduction <br/>

rspamd-client-go is a client library written to help interact with a rspamd instance, via HTTP. rspamd-client-go facilitates
* Content scanning of emails
* Training rspamd's Bayesian classifier
* Updating rspamd's fuzzy storage by adding or removing messages 
* Analyzing content scanning results 

Refer to rspamd [documentation](https://rspamd.com/doc/) for help configuring and setting up rspamd.

## Usage 

The API is defined [here](https://pkg.go.dev/gopkg.in/rspamd-client-go.v1). The examples below are just that - examples. For the full piccture, reference the documentation.

Generally the library can be thought of as partitioned into two parts, the client, and analysis. 

The client helps send emails to all POST endpoints on rspamd. Support for all GET endpoints does not currently exist as they can be accessed through rspamd's web interface, although support may exist in the future. A full list of all endpoints can be found [here](https://rspamd.com/doc/architecture/protocol.html). 

The client supports email formats that implement `io.Reader` or `io.WriteTo`. This means that clients can pass in both `gomail.Message` objects, which implement `io.WriteTo` or simply the contents of an `.eml` file, which implement `io.Reader`. gomail can be found [here](https://github.com/go-gomail/gomail).

The analysis currently supports casting meaning to different tiers of spam score. In the future, support may involve integrating and analyzing based off rspamd symbols.

### Examples

_Note:_ rspamd-client-go is geared towards clients that use [context](https://golang.org/pkg/context/). However if you don't, whenever `context.Context` is expected, you can use `context.Background()`.

Import rspamdclient:
```go
import "github.com/Shopify/rspamd-client-go"
```
Instantiate the client with the url of your rspamd instance, and optional credentials:
```go
 client := rspamdclient.New("https://contentscanner.com")
 // optionally pass in rspamdclient.Credentials("username", "password") as second argument
```
Ping your rspamd instance:
```go
pong, err := client.Ping(ctx)
```
Scan an email from a gomail `Message` (or anything that implements `io.WriteTo`):
```go
// let email be of type *gomail.Message
// attach a Queue-Id to rspamdclient.Email instance
checkRes, err := client.Check(ctx, rspamdclient.NewEmailFromWriter(email).QueueId(1))
```
Scan an email from an `.eml` file (or anything that implements `io.Reader`):
```go
// attach a Queue-Id to rspamdclient.Email instance
email := rspamdclient.NewEmailFromReader(rspamdclient.MustOpen("/path/to/email")).QueueId(2))
checkRes, err := client.Check(ctx, email)
```
Add a message to fuzzy storage, attaching a flag and weight as per [docs](https://rspamd.com/doc/architecture/protocol.html#controller-http-endpoints):
```go
// let email be of type *gomail.Message
learnRes, err := client.FuzzyAdd(ctx, rspamdclient.NewEmailFromWriter(email).QueueId(2).Flag(1).Weight(19))
```

## Semantics

### Contributing

### Versioning

### License
