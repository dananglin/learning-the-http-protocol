package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	address := "localhost:42069"

	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		printErrorAndExit(fmt.Errorf("error resolving the UDP endpoint: %w", err))
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		printErrorAndExit(fmt.Errorf("error creating the UDP connection: %w", err))
	}
	defer udpConn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("READ ERROR: %v.\n", err)

			continue
		}

		_, err = udpConn.Write([]byte(userInput))
		if err != nil {
			fmt.Printf("WRTIE ERROR: %v.\n", err)

			continue
		}
	}
}

func printErrorAndExit(err error) {
	fmt.Printf("ERROR: %v.\n", err)

	os.Exit(1)
}
