package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	netAddress, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
		return
	}
	netConn, err := net.DialUDP("udp", nil, netAddress)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer netConn.Close()

	buffer := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		text, err := buffer.ReadString('\n')
		if err != nil {
			log.Fatal(err)
			return
		}
		data := []byte(text)
		_, err = netConn.Write(data)
		if err != nil {
			log.Fatal(err)
			return
		}

	}

}
