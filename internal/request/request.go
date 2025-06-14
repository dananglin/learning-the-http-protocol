package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const (
	supportedHttpVersion string = "HTTP/1.1"
	newLineChars         string = "\r\n"
	bufferSize           int    = 8
)

type requestState int

const (
	requestStateInitialiased = iota
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	state       requestState
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0

	request := Request{
		state: requestStateInitialiased,
	}

	for request.state != requestStateDone {
		// Increase the size of the buffer if it is full.
		if readToIndex >= cap(buf) {
			buf = increaseBufferSize(buf)
		}

		// Read the data from the reader and note the size of the data that
		// has been read.
		sizeOfRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				// request.state = requestStateDone

				break
			}

			return nil, fmt.Errorf("error reading the data: %w", err)
		}

		// Update the readToIndex
		readToIndex = readToIndex + sizeOfRead

		// Now try and parse the data and note the size of the data that has been parsed (if parsed).
		sizeOfParsed, err := request.parse(buf)
		if err != nil {
			return nil, fmt.Errorf("error parsing the data: %w", err)
		}

		// If the data has not been parsed, re-loop so that we can read more data into the buffer.
		if sizeOfParsed == 0 {
			continue
		}

		buf = clearParsedData(buf, sizeOfParsed)
		readToIndex = readToIndex - sizeOfParsed
	}

	return &request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == requestStateInitialiased {
		parsed, sizeOfParsed, err := parseRequestLine(string(data))
		if err != nil {
			return 0, fmt.Errorf("request parsing error: %w", err)
		}

		// More data is needed from the requester.
		if sizeOfParsed == 0 {
			return 0, nil
		}

		// Update the state and add the parsed request line to r.
		r.RequestLine = parsed
		r.state = requestStateDone

		// Return the size (in bytes) of the original request line that was parsed.
		return sizeOfParsed, nil
	}

	if r.state == requestStateDone {
		return 0, fmt.Errorf("request parsing error: attempt to read data in a done state")
	}

	return 0, fmt.Errorf("request parsing error: unknown state")
}

type RequestLine struct {
	Method        string
	RequestTarget string
	HTTPVersion   string
}

// parseRequestLine parses the request line of the request. If successful, parseRequestLine
// returns the parsed request line and the size (in bytes) of the original request line that
// was parsed.
func parseRequestLine(req string) (RequestLine, int, error) {
	if !strings.Contains(req, newLineChars) {
		return RequestLine{}, 0, nil
	}

	parts := strings.SplitN(req, newLineChars, 2)
	if len(parts) != 2 {
		return RequestLine{}, 0, requestPartsError{len(parts)}
	}

	reqLine := parts[0]

	parts = strings.Split(reqLine, " ")
	if len(parts) != 3 {
		return RequestLine{}, 0, requestLinePartsError{len(parts)}
	}

	method, requestTarget, httpVersion := parts[0], parts[1], parts[2]

	// Verify that the method is all caps
	for _, letter := range method {
		if !unicode.IsUpper(letter) {
			return RequestLine{}, 0, methodFormatError{method}
		}
	}

	// Verify that the HTTP Version is literally HTTP/1.1
	if httpVersion != supportedHttpVersion {
		return RequestLine{}, 0, unsupportedHTTPVersionError{
			supportedVersion: supportedHttpVersion,
			gotVersion:       httpVersion,
		}
	}

	return RequestLine{
			Method:        method,
			RequestTarget: requestTarget,
			HTTPVersion:   "1.1",
		},
		len([]byte(reqLine)) + len([]byte(newLineChars)),
		nil
}

// increaseBufferSize returns a buffer that is double the capacity of the
// input buffer with the data of the input buffer copied over to the
// output buffer.
func increaseBufferSize(buf []byte) []byte {
	newBufferSize := cap(buf) + bufferSize
	output := make([]byte, newBufferSize, newBufferSize)
	copy(output, buf)

	return output
}

func clearParsedData(buf []byte, sizeOfParsed int) []byte {
	output := make([]byte, cap(buf), cap(buf))
	copy(output, buf[sizeOfParsed:])

	return output
}
