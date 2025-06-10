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
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("WARNING: error closing file: %v.\n", err)
		}
	}()

	currentLine := ""

	for !errors.Is(err, io.EOF) {
		data := make([]byte, 8)

		_, err = file.Read(data)

		parts := strings.Split(string(data), "\n")

		currentLine += parts[0]

		if len(parts) == 2 {
			fmt.Println("read:", currentLine)
			currentLine = parts[1]
		}
	}

	return nil
}
