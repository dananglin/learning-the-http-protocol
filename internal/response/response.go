package response

import (
	"errors"
	"fmt"
	"io"

	"http-from-tcp/internal/headers"
)

type StatusCode int

const (
	StatusCodeOK          StatusCode = 200
	StatusCodeBadRequest  StatusCode = 400
	StatusCodeServerError StatusCode = 500
)

type writerState int

const (
	writerStateInitialised = iota
	writerStateHeaders
	writerStateBody
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  writerStateInitialised,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateInitialised {
		return errors.New("the response writer is not in the correct state to write the status line")
	}

	statuses := map[StatusCode]string{
		StatusCodeOK:          "HTTP/1.1 200 OK",
		StatusCodeBadRequest:  "HTTP/1.1 400 Bad Request",
		StatusCodeServerError: "HTTP/1.1 500 Internal Server Error",
	}

	_, err := w.writer.Write([]byte(statuses[statusCode] + "\n"))
	if err != nil {
		return fmt.Errorf("error writing the status line: %w", err)
	}

	w.state = writerStateHeaders

	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateHeaders {
		return errors.New("the response writer is not in the correct state to write the headers")
	}

	for key, value := range headers {
		header := key + ": " + value + "\n"
		_, err := w.writer.Write([]byte(header))
		if err != nil {
			return fmt.Errorf(
				"error writing the header %q: %w",
				header,
				err,
			)
		}
	}

	_, err := w.writer.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf(
			"error writing the final CRLF: %w",
			err,
		)
	}

	w.state = writerStateBody

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, errors.New("the response writer is not in the correct state to write the body")
	}

	return w.writer.Write(p)
}
