package response

import (
	"errors"
	"fmt"
)

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, errors.New("the response writer is not in the correct state to write the chunked body")
	}

	data := string(p) + "\r\n"

	return w.writer.Write([]byte(data))
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writerStateBody {
		return 0, errors.New("the response writer is not in the correct state to write the chunked body")
	}

	const chunkedBodyDone string = "0\r\n"

	n, err := w.writer.Write([]byte(chunkedBodyDone))
	if err != nil {
		return 0, fmt.Errorf("error writing the end of the chunked body: %w", err)
	}

	w.state = writerStateTrailers

	return n, nil
}
