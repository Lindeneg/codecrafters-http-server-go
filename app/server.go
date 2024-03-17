package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	ResponseOK       = "HTTP/1.1 200 OK"
	ResponseNotFound = "HTTP/1.1 404 Not Found"
	TypeTextPlain    = "text/plain"
	TypeOctetStream  = "application/octet-stream"
)

var (
	protocol  string
	host      string
	port      string
	directory string
)

func main() {
	parseEnv()
	l, err := net.Listen(protocol, fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		fmt.Println("Failed to bind to port ", port)
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

type headers map[string]string

type request struct {
	method  string
	path    string
	version string
	headers headers
}

func (r request) IsGet() bool {
	return r.method == "GET"
}

func (r request) IsPost() bool {
	return r.method == "POST"
}

type response struct {
	status  string
	headers headers
	content string
}

func (res response) WriteToConn(conn net.Conn) error {
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

func parseEnv() {
	flag.StringVar(&protocol, "protocol", "tcp", "protocol to use")
	flag.StringVar(&host, "host", "0.0.0.0", "host to use")
	flag.StringVar(&port, "port", "4221", "port to use")
	flag.StringVar(&directory, "directory", "", "dir with files to serve")
	flag.Parse()
	fmt.Printf("Listening at %s://%s:%s and serving directory %q\n", protocol, host, port, directory)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	req, err := connectionToRequest(conn)
	if err != nil {
		fmt.Println("Error parsing connection as request: ", err.Error())
		return
	}
	res := response{}
	switch {
	case req.IsGet():
		handleGetRequest(req, &res)
	default:
		res.status = ResponseNotFound
	}
	err = res.WriteToConn(conn)
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

func handleGetRequest(req request, res *response) {
	if req.path == "/" {
		res.status = ResponseOK
		return
	}
	if req.path == "/user-agent" {
		responseContent(res, req.headers["User-Agent"], TypeTextPlain)
		return
	}
	if p, ok := strings.CutPrefix(req.path, "/echo/"); ok {
		responseContent(res, p, TypeTextPlain)
		return
	}
	if p, ok := strings.CutPrefix(req.path, "/files/"); ok {
		bytes, err := os.ReadFile(fmt.Sprintf("%s/%s", directory, p))
		if err == nil {
			responseContent(res, string(bytes), TypeOctetStream)
			return
		}
	}
	res.status = ResponseNotFound
}

func responseContent(res *response, content string, contentType string) {
	if res.headers == nil {
		res.headers = make(headers, 2)
	}
	res.status = ResponseOK
	res.headers["Content-Type"] = contentType
	res.headers["Content-Length"] = fmt.Sprint(len(content))
	res.content = content
}
