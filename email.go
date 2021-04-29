package rspamd

import (
	"io"

	"github.com/pkg/errors"
)

type Email struct {
	message io.Reader
	queueID string
	options Options
}

type Options struct {
	flag   int
	weight float64
}

type SymbolData struct {
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	MetricScore float64 `json:"metric_score"`
	Description string  `json:"description"`
}

func NewEmailFromWriterTo(message io.WriterTo) *Email {
	return &Email{
		message: readerFromWriterTo(message),
		options: Options{},
	}
}

func NewEmailFromReader(message io.Reader) *Email {
	return &Email{
		message: message,
		options: Options{},
	}
}

func (e *Email) QueueID(queueID string) *Email {
	e.queueID = queueID
	return e
}

func (e *Email) Flag(flag int) *Email {
	e.options.flag = flag
	return e
}

func (e *Email) Weight(weight float64) *Email {
	e.options.weight = weight
	return e
}

func readerFromWriterTo(writerTo io.WriterTo) io.Reader {
	r, w := io.Pipe()

	go func() {
		if _, err := writerTo.WriteTo(w); err != nil {
			_ = w.CloseWithError(errors.Wrap(err, "writing to pipe"))
			return
		}

		_ = w.Close() // Always succeeds
	}()

	return r
}
