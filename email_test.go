package rspamd

import (
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type faultyWriterTo struct {
	err error
}

func (m *faultyWriterTo) WriteTo(w io.Writer) (n int64, err error) {
	return 0, m.err
}

func Test_readerFromWriterTo(t *testing.T) {
	r := readerFromWriterTo(&faultyWriterTo{err: errors.New("foo")})
	_, err := r.Read([]byte{})
	require.EqualError(t, err, "writing to pipe: foo")
}
