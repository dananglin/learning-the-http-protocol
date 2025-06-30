package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"http-from-tcp/internal/server"
)

const port = 42069

const htmlTemplate = `<html>
  <head>
    <title>{{ .Title }}</title>
  </head>
  <body>
    <h1>{{ .Header }}</h1>
    <p>{{ .Paragraph }}</p>
  </body>
</html>`

func main() {
	loggingLevel := new(slog.LevelVar)
	slogOpts := slog.HandlerOptions{Level: loggingLevel}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slogOpts))

	slog.SetDefault(logger)
	loggingLevel.Set(slog.LevelInfo)

	if err := run(); err != nil {
		slog.Error(err.Error())

		os.Exit(1)
	}
}

func run() error {
	server, err := server.Serve(port, handler)
	if err != nil {
		return fmt.Errorf("error starting the server: %w", err)
	}
	defer server.Close()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()
	stop()

	slog.Info("Server gracefully stopped.")

	return nil
}

func handler(w *response.Writer, req *request.Request) {
	switch {
	case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/"):
		proxyHandler(w, req, "https://httpbin.org")
	default:
		serverHandler(w, req.RequestLine.RequestTarget)
	}
}

func serverHandler(w *response.Writer, path string) {
	var (
		statusCode    response.StatusCode
		status        string
		htmlHeader    string
		htmlParagraph string
	)

	switch path {
	case "/yourproblem":
		statusCode = response.StatusCodeBadRequest
		status = "Bad Request"
		htmlHeader = status
		htmlParagraph = "Your request honestly kinda sucked."
	case "/myproblem":
		statusCode = response.StatusCodeServerError
		status = "Internal Server Error"
		htmlHeader = status
		htmlParagraph = "Okay, you know what? This one is on me."
	default:
		statusCode = response.StatusCodeOK
		status = "OK"
		htmlHeader = "Success!"
		htmlParagraph = "Your request was an absolute banger."
	}

	if err := w.WriteStatusLine(statusCode); err != nil {
		slog.Error("error writing the status line", "error", err.Error())

		return
	}

	buf := new(bytes.Buffer)

	tmpl := template.Must(template.New("response").Parse(htmlTemplate))

	data := struct {
		Title     string
		Header    string
		Paragraph string
	}{
		Title:     fmt.Sprintf("%d %s", int(statusCode), status),
		Header:    htmlHeader,
		Paragraph: htmlParagraph,
	}

	if err := tmpl.Execute(buf, data); err != nil {
		slog.Error("error executing the HTML template", "error", err.Error())

		return
	}

	headers := response.GetDefaultHeaders(buf.Len())
	headers.Edit("Content-Type", "text/html")

	if err := w.WriteHeaders(headers); err != nil {
		slog.Error("error writing the headers", "error", err.Error())

		return
	}

	_, err := w.WriteBody(buf.Bytes())
	if err != nil {
		slog.Error("error writing the response body", "error", err.Error())

		return
	}
}

func proxyHandler(w *response.Writer, req *request.Request, baseURL string) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(60*time.Second),
	)
	defer cancel()

	proxyReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		baseURL+"/"+strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/"),
		nil,
	)
	if err != nil {
		slog.Error("error creating the proxy request", "error", err.Error())

		return
	}

	client := http.Client{}

	resp, err := client.Do(proxyReq)
	if err != nil {
		slog.Error("error getting the response from the server", "error", err.Error())

		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error(
			"unexpected status code returned from the server",
			"code",
			strconv.Itoa(resp.StatusCode),
			"status",
			resp.Status,
		)

		return
	}

	if err := w.WriteStatusLine(response.StatusCodeOK); err != nil {
		slog.Error("error writing the status line", "error", err.Error())

		return
	}

	headers := response.GetDefaultHeaders(0)
	headers.Delete(response.HeaderContentLength)
	headers.Delete(response.HeaderConnection)
	headers.Add(response.HeaderTransferEncoding, "chunked")

	if err := w.WriteHeaders(headers); err != nil {
		slog.Error("error writing the headers", "error", err.Error())

		return
	}

	buf := make([]byte, 1024, 1024)

ResponseReadLoop:
	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break ResponseReadLoop
			}

			slog.Error(
				"error reading the data from the response body",
				"error",
				err.Error(),
			)

			return
		}

		slog.Info("A chunk of data was read from the response", "size", n)

		chunkSize := strings.ToUpper(strconv.FormatInt(int64(n), 16))

		_, err = w.WriteChunkedBody([]byte(chunkSize))
		if err != nil {
			slog.Error(
				"error writing the chunk size",
				"error",
				err.Error(),
			)
		}

		_, err = w.WriteChunkedBody(buf[:n])
		if err != nil {
			slog.Error(
				"error writing a chunk of the body",
				"error",
				err.Error(),
			)
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		slog.Error(
			"error writing the end of the chunked body response body",
			"error",
			err.Error(),
		)

		return
	}
}
