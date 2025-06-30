package response

import "errors"

const chunkedBodyDone string = "0\n\n"

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, errors.New("the response writer is not in the correct state to write the body")
	}

	data := string(p) + "\n"

	return w.writer.Write([]byte(data))
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writerStateBody {
		return 0, errors.New("the response writer is not in the correct state to write the body")
	}

	return w.writer.Write([]byte(chunkedBodyDone))
}
