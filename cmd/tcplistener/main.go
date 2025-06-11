package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// const messagesFilepath string = "messages.txt"

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

		linesChan := getLinesChannel(conn)

	ReadLines:
		for {
			line, ok := <-linesChan
			if !ok {
				break ReadLines
			}

			fmt.Println(line)
		}

		fmt.Println("Connection closed.")
	}

	return nil
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	linesChan := make(chan string)

	go func() {
		defer close(linesChan)
		defer f.Close()

		var err error
		currentLine := ""
		var data []byte

		for {
			data = make([]byte, 8)
			_, err = f.Read(data)
			if err != nil {
				if errors.Is(err, io.EOF) {
					if currentLine != "" {
						linesChan <- currentLine
					}

					break
				} else {
					linesChan <- fmt.Sprintf("ERROR: %v", err)

					break
				}
			}

			parts := strings.Split(string(data), "\n")

			currentLine += parts[0]

			if len(parts) == 2 {
				linesChan <- currentLine
				currentLine = parts[1]
			}
		}
	}()

	return linesChan
}
