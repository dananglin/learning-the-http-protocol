package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"http-from-tcp/internal/headers"
)

const (
	supportedHttpVersion string = "HTTP/1.1"
	crlf                 string = "\r\n"
	endOfHeaders         string = crlf + crlf
	bufferSize           int    = 8
)

type requestState int

const (
	requestStateInitialiased = iota
	requestStateParsingHeaders
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0

	request := Request{
		state: requestStateInitialiased,
	}

ProcessRequest:
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
				switch request.state {
				case requestStateInitialiased:
					return nil, incompleteRequestLineError{}
				case requestStateParsingHeaders:
					return nil, incompleteHeadersLineError{}
				default:
					break ProcessRequest
				}
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
	switch r.state {
	case requestStateInitialiased:
		parsed, sizeOfParsed, err := parseRequestLine(string(data))
		if err != nil {
			return 0, fmt.Errorf(
				"error parsing the request line from the request: %w",
				err,
			)
		}

		// More data is needed from the requester.
		if sizeOfParsed == 0 {
			return 0, nil
		}

		// Add the parsed request line to r and
		// update the state to indicate that the next thing to parse are
		// the headers.
		r.RequestLine = parsed
		r.state = requestStateParsingHeaders

		// Return the size (in bytes) of the original request line that was parsed.
		return sizeOfParsed, nil
	case requestStateParsingHeaders:
		headers, sizeOfParsed, err := parseHeaders(data)
		if err != nil {
			return 0, fmt.Errorf(
				"error parsing the headers from the request: %w",
				err,
			)
		}

		// More data is needed from the requester.
		if sizeOfParsed == 0 {
			return 0, nil
		}

		// Add the parsed headers to r and
		// update the state to indicate that we are done parsing (NOTE: for now)
		r.Headers = headers
		r.state = requestStateDone

		// Return the size (in bytes) of the original headers line that was parsed.
		return sizeOfParsed, nil
	case requestStateDone:
		return 0, errors.New("request parsing error: attempt to read data in a done state")
	default:
		return 0, errors.New("request parsing error: unknown state")
	}
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
	if !strings.Contains(req, crlf) {
		return RequestLine{}, 0, nil
	}

	parts := strings.SplitN(req, crlf, 2)
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
		len([]byte(reqLine)) + len([]byte(crlf)),
		nil
}

func parseHeaders(data []byte) (headers.Headers, int, error) {
	if !strings.Contains(string(data), endOfHeaders) {
		// More data required.
		return headers.Headers{}, 0, nil
	}

	var (
		reqHeaders        = headers.NewHeaders()
		totalSizeOfParsed = 0
	)

	for {
		sizeOfParsed, done, err := reqHeaders.Parse(data[totalSizeOfParsed:])
		if err != nil {
			return headers.Headers{}, 0, fmt.Errorf("header parsing error: %w", err)
		}

		if done {
			// No headers were parsed if this is set to true,
			// so just break out of the loop.
			break
		}

		totalSizeOfParsed += sizeOfParsed
	}

	return reqHeaders, totalSizeOfParsed + len([]byte(crlf)), nil
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
