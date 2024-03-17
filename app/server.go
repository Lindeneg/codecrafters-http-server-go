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

type response struct {
	status  string
	headers headers
	content string
}

const (
	ResponseOK       = "HTTP/1.1 200 OK"
	ResponseNotFound = "HTTP/1.1 404 Not Found"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	req, err := connectionToRequest(conn)
	if err != nil {
		fmt.Println("Error parsing connection as request: ", err.Error())
		return
	}
	err = handleResponse(conn, req)
	if err != nil {
		fmt.Println("Error responding to request: ", err.Error())
		return
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
	res := response{}
	if req.path == "/" {
		res.status = ResponseOK
	} else if req.path == "/user-agent" {
		responseContent(&res, req.headers["User-Agent"])
	} else if p, ok := strings.CutPrefix(req.path, "/echo/"); ok {
		responseContent(&res, p)
	} else {
		res.status = ResponseNotFound
	}
	return writeResponse(conn, res)
}

func responseContent(res *response, content string) {
	res.status = ResponseOK
	res.headers = make(headers, 2)
	res.headers["Content-Type"] = "text/plain"
	res.headers["Content-Length"] = fmt.Sprint(len(content))
	res.content = content
}

func writeResponse(conn net.Conn, res response) error {
	_, err := conn.Write([]byte(fmt.Sprintf("%s\r\n", res.status)))
	if err != nil {
		return err
	}
	for k, v := range res.headers {
		_, err := conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		if err != nil {
			return err
		}
	}
	_, err = conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	if len(res.content) > 0 {
		_, err := conn.Write([]byte(fmt.Sprintf("%s\r\n", res.content)))
		if err != nil {
			return err
		}
	}
	return nil
}
