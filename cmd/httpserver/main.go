package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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
	server, err := server.Serve(port, serverHandler)
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

func serverHandler(w *response.Writer, req *request.Request) {
	var (
		statusCode    response.StatusCode
		status        string
		htmlHeader    string
		htmlParagraph string
	)

	switch req.RequestLine.RequestTarget {
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
	headers.Override("Content-Type", "text/html")

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
