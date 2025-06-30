package response

import (
	"strconv"

	"http-from-tcp/internal/headers"
)

const (
	HeaderContentLength    = "Content-Length"
	HeaderContentType      = "Content-Type"
	HeaderConnection       = "Connection"
	HeaderTransferEncoding = "Transfer-Encoding"
	HeaderTrailer          = "Trailer"
)

// GetDefaultHeaders returns the default response headers.
func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers[HeaderContentLength] = strconv.Itoa(contentLen)
	headers[HeaderContentType] = "text/plain"
	headers[HeaderConnection] = "close"

	return headers
}
