package main

import (
	"fmt"
	"github.com/GhostVox/httptcp/internal/request"
	"log"
	"net"
)

func main() {
	tcpNetwork, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("Failed to setup tcp on port :42069")
	}
	defer tcpNetwork.Close()

	for {
		conn, err := tcpNetwork.Accept()
		if err != nil {
			log.Fatal("Failed to accept connection")
		}
		log.Println("Tcp Connection has been accepted")
		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("Failed to read request")
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)

		fmt.Printf("Channel has been closed\n")
	}

}
