package headers

import (
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	headers := make(map[string]string)

	return Headers(headers)
}

const (
	crlf                 string = "\r\n"
	headerValidationRule string = `^ *[^\s]*: *[^\s]* *$`
)

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if !strings.Contains(string(data), crlf) {
		// More data required.
		return 0, false, nil
	}

	if strings.HasPrefix(string(data), crlf) {
		// We are done parsing the headers in this request.
		return 0, true, nil
	}

	parts := strings.SplitN(string(data), crlf, 2)
	if len(parts) != 2 {
		return 0, false, fmt.Errorf(
			"unexpected number of parts found after extracting the header from %q: want 2, got %d",
			string(data),
			len(parts),
		)
	}

	header := parts[0]

	if err := validateHeader(header); err != nil {
		return 0, false, fmt.Errorf("header validation error: %w", err)
	}

	key, value, err := extractHeader(header)
	if err != nil {
		return 0, false, fmt.Errorf(
			"header extraction error: %w",
			err,
		)
	}

	h[key] = value

	return len([]byte(header)) + len([]byte(crlf)), false, nil
}

func validateHeader(header string) error {
	pattern := regexp.MustCompile(headerValidationRule)

	if !pattern.MatchString(header) {
		return fmt.Errorf("invalid header: %s", header)
	}

	return nil
}

func extractHeader(header string) (key string, value string, err error) {
	parts := strings.SplitN(header, ":", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"unexpected number of parts found after extracting the header: want 2, got %d",
			len(parts),
		)
	}

	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])

	return key, value, err
}
