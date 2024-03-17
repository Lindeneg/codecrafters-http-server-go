package main

import (
	"fmt"
	"net"
	"os"
)

const (
	ResponseOK = "HTTP/1.1 200 OK\r\n\r\n"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading request bytes: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Received: ", string(buf))
	_, err = conn.Write([]byte(ResponseOK))
	if err != nil {
		fmt.Println("Error writing response bytes: ", err.Error())
		os.Exit(1)
	}
}
