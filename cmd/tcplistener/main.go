package main

import (
	"fmt"
	"net"
	"os"

	"http-from-tcp/internal/request"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR: %v.\n", err)

		os.Exit(1)
	}
}

func run() error {
	address := "localhost:42069"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("error creating the listener: %w", err)
	}
	defer listener.Close()

	fmt.Println("Server is listening on:", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("CONNECTION ERROR: %v.\n", err)

			break
		}

		fmt.Println("Connection accepted.")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Printf("ERROR: error reading the request from the connection: %v", err)

			break
		}

		result := fmt.Sprintf(
			"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			req.RequestLine.Method,
			req.RequestLine.RequestTarget,
			req.RequestLine.HTTPVersion,
		)

		if len(req.Headers) > 0 {
			result += "Headers:\n"

			for key, value := range req.Headers {
				result += fmt.Sprintf("- %s: %s\n", key, value)
			}
		}

		result += fmt.Sprintf("Body:\n%s", string(req.Body))

		fmt.Println(result)

		fmt.Println("Connection closed.")
	}

	return nil
}
