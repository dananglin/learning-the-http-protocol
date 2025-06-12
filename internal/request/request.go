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
		return RequestLine{}, requestPartsError{len(parts)}
	}

	reqLine := parts[0]

	parts = strings.Split(reqLine, " ")
	if len(parts) != 3 {
		return RequestLine{}, requestLinePartsError{len(parts)}
	}

	method, requestTarget, httpVersion := parts[0], parts[1], parts[2]

	// Verify that the method is all caps
	for _, letter := range method {
		if !unicode.IsUpper(letter) {
			return RequestLine{}, methodFormatError{method}
		}
	}

	// Verify that the HTTP Version is literally HTTP/1.1
	if httpVersion != supportedHttpVersion {
		return RequestLine{}, unsupportedHTTPVersionError{
			supportedVersion: supportedHttpVersion,
			gotVersion:       httpVersion,
		}
	}

	return RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HTTPVersion:   "1.1",
	}, nil
}
