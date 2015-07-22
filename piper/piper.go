package piper

import (
	"bytes"
	"io"
	"sync"
)

// PipeStatus holds the satus of a buffered pipe.
// It tells how many bytes have been read from the source,
// wrote to the destination and were buffered.
type PipeStatus struct {
	Read,
	Wrote,
	Buffered int

	ReadError,
	WriteError,
	BufferError error
}

// PipeOutput links the rc to writer w we pass and also writes to buf if it is not null
func PipeOutput(wg *sync.WaitGroup, rc io.ReadCloser, w io.Writer, buf *bytes.Buffer) (s PipeStatus) {
	defer wg.Done()
	// if we have no rc we cannot do anything, because
	// that's where the data come from
	if rc == nil {
		return
	}
	defer rc.Close()

	tmp := make([]byte, 1024)

	// to count how many bytes we read/write/buffer on
	// every loop
	var cR, cW, cB int

	for s.ReadError == nil && (s.WriteError == nil || s.BufferError == nil) {
		cR, s.ReadError = rc.Read(tmp)
		s.Read += cR

		if cR == 0 {
			continue
		}

		if buf != nil && s.BufferError == nil {
			cB, s.BufferError = buf.Write(tmp[0:cR])
			s.Buffered += cB
		}

		if w != nil && s.WriteError == nil {
			cW, s.WriteError = w.Write(tmp[0:cR])
			s.Wrote += cW
		}
	}
	return
}
