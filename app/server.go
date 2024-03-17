package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

type headers map[string]string

type request struct {
	method  string
	path    string
	version string
	headers headers
}

const (
	ResponseOK       = "HTTP/1.1 200 OK\r\n\r\n"
	ResponseNotFound = "HTTP/1.1 404 Not Found\r\n\r\n"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	defer conn.Close()

	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	req, err := connectionToRequest(conn)
	if err != nil {
		fmt.Println("Error parsing connection as request: ", err.Error())
		os.Exit(1)
	}

	err = handleResponse(conn, req)
	if err != nil {
		fmt.Println("Error responding to request: ", err.Error())
		os.Exit(1)
	}
}

func connectionToRequest(conn net.Conn) (req request, err error) {
	buf := make([]byte, 32<<8)
	_, err = conn.Read(buf)
	if err != nil {
		return req, err
	}
	parts := strings.Split(string(buf), "\r\n")
	if len(parts) == 0 {
		return req, errors.New("HTTP startline missing")
	}
	startLine := parts[0]
	headerLines := parts[1:]
	err = parseStartline(startLine, &req)
	if err != nil {
		return req, err
	}
	parseHeaderLines(headerLines, &req)
	return req, nil
}

func parseStartline(startLine string, req *request) error {
	startLines := strings.Split(startLine, " ")
	if len(startLines) != 3 {
		return errors.New("HTTP startline should contain METHOD PATH VERSION")
	}
	req.method = startLines[0]
	req.path = startLines[1]
	req.version = startLines[2]
	return nil
}

func parseHeaderLines(headerLines []string, req *request) {
	if req.headers == nil {
		req.headers = make(headers, len(headerLines))
	}
	for _, line := range headerLines {
		splittedLine := strings.Split(line, ": ")
		if len(splittedLine) == 2 {
			req.headers[splittedLine[0]] = splittedLine[1]
		}
	}
}

func handleResponse(conn net.Conn, req request) error {
	var err error
	switch req.path {
	case "/":
		_, err = conn.Write([]byte(ResponseOK))
		break
	default:
		_, err = conn.Write([]byte(ResponseNotFound))
		break
	}
	return err
}
