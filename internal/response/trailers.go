package response

import (
	"errors"
	"fmt"
	"strings"

	"http-from-tcp/internal/headers"
)

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writerStateTrailers {
		return errors.New("the response writer is not in the correct state to write the trailers")
	}

	// get trailers from Trailer
	trailers := strings.Split(h[HeaderTrailer], ", ")

	// for each trailer, write key and value
	for idx := range trailers {
		trailer := trailers[idx] + ": " + h[trailers[idx]] + "\r\n"
		_, err := w.writer.Write([]byte(trailer))
		if err != nil {
			return fmt.Errorf(
				"error writing the trailer %q: %w",
				trailer,
				err,
			)
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf(
			"error writing the final CRLF: %w",
			err,
		)
	}

	return nil
}
