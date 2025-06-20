package response

import (
	"fmt"
	"io"
	"strconv"

	"http-from-tcp/internal/headers"
)

type StatusCode int

const (
	StatusCodeOK          StatusCode = 200
	StatusCodeBadRequest  StatusCode = 400
	StatusCodeServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statuses := map[StatusCode]string{
		StatusCodeOK:          "HTTP/1.1 200 OK",
		StatusCodeBadRequest:  "HTTP/1.1 400 Bad Request",
		StatusCodeServerError: "HTTP/1.1 500 Internal Server Error",
	}

	_, err := w.Write([]byte(statuses[statusCode] + "\n"))
	if err != nil {
		return fmt.Errorf("error writing the status line: %w", err)
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["Content-Length"] = strconv.Itoa(contentLen)
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		header := key + ": " + value + "\n"
		_, err := w.Write([]byte(header))
		if err != nil {
			return fmt.Errorf(
				"error writing the header %q: %w",
				header,
				err,
			)
		}
	}

	w.Write([]byte("\n"))

	return nil
}
