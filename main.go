package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const messagesFilepath string = "messages.txt"

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR: %v.\n", err)

		os.Exit(1)
	}
}

func run() error {
	var err error

	file, err := os.Open(messagesFilepath)
	if err != nil {
		return fmt.Errorf(
			"error opening %s: %w",
			messagesFilepath,
			err,
		)
	}

	linesChan := getLinesChannel(file)

Outer:
	for {
		line, ok := <-linesChan
		if !ok {
			break Outer
		}

		fmt.Println("read:", line)
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
