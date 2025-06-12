package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

const supportedHttpVersion string = "HTTP/1.1"

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HTTPVersion   string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf(
			"error reading the data: %w",
			err,
		)
	}

	reqLine, err := parseRequestLine(string(req))
	if err != nil {
		return nil, fmt.Errorf(
			"error parsing the request line: %w",
			err,
		)
	}

	return &Request{
		RequestLine: reqLine,
	}, nil
}

func parseRequestLine(req string) (RequestLine, error) {
	parts := strings.SplitN(req, "\r\n", 2)
	if len(parts) != 2 {
		return RequestLine{}, fmt.Errorf("received an unexpected number of parts after splitting the REQUEST: want: 2, got: %d", len(parts))
	}

	reqLine := parts[0]

	parts = strings.Split(reqLine, " ")
	if len(parts) != 3 {
		return RequestLine{}, fmt.Errorf("received an unexpected number of parts after splitting the REQUEST LINE: want: 3, got: %d", len(parts))
	}

	method, requestTarget, httpVersion := parts[0], parts[1], parts[2]

	// Verify that the method is all caps
	for _, letter := range method {
		if !unicode.IsUpper(letter) {
			return RequestLine{}, fmt.Errorf("the received HTTP method %q is incorrectly formatted", method)
		}
	}

	// Verify that the HTTP Version is literally HTTP/1.1
	if httpVersion != supportedHttpVersion {
		return RequestLine{}, fmt.Errorf("received an unsupported HTTP version in the request: want: %q, got %q", supportedHttpVersion, httpVersion)
	}

	return RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HTTPVersion:   "1.1",
	}, nil
}
