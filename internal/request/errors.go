package request

import "fmt"

type requestPartsError struct {
	numParts int
}

func (e requestPartsError) Error() string {
	return fmt.Sprintf(
		"received an unexpected number of parts after splitting the REQUEST: want: 2, got: %d",
		e.numParts,
	)
}

type requestLinePartsError struct {
	numparts int
}

func (e requestLinePartsError) Error() string {
	return fmt.Sprintf(
		"received an unexpected number of parts after splitting the REQUEST LINE: want: 3, got: %d",
		e.numparts,
	)
}

type methodFormatError struct {
	method string
}

func (e methodFormatError) Error() string {
	return "the received HTTP method '" +
		e.method +
		"' is incorrectly formatted"
}

type unsupportedHTTPVersionError struct {
	supportedVersion string
	gotVersion       string
}

func (e unsupportedHTTPVersionError) Error() string {
	return "received an unsupported HTTP version in the request: want " +
		e.supportedVersion +
		", got " +
		e.gotVersion
}

type incompleteHeadersLineError struct{}

func (e incompleteHeadersLineError) Error() string {
	return "the HEADERS LINE appears to be incomplete or missing"
}

type incompleteRequestLineError struct{}

func (e incompleteRequestLineError) Error() string {
	return "the REQUEST LINE appears to be incomplete or missing"
}

type incompleteBodyError struct{}

func (e incompleteBodyError) Error() string {
	return "the BODY of the request appears to be incomplete or missing"
}
