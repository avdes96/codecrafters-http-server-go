package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

type Request struct {
	requestLine *RequestLine
}

type RequestLine struct {
	method  string
	target  string
	version string
}

func getResponse(target string) []byte {
	if target == "/" {
		return get200Response("")
	}
	if strings.HasPrefix(target, "/echo/") {
		return get200Response(strings.TrimPrefix(target, "/echo/"))
	}
	return get404Response()
}

func getLineToCrlf(reader *bufio.Reader) ([]byte, error) {
	buf := []byte{}
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		buf = append(buf, line...)
		if len(buf) >= 2 && line[len(buf)-2] == '\r' {
			return buf[:len(buf)-2], nil
		}

	}
}

func parseRequest(reader *bufio.Reader) (*Request, error) {
	rl, err := parseRequestLine(reader)
	if err != nil {
		return nil, err
	}
	return &Request{requestLine: rl}, nil
}

func parseRequestLine(r *bufio.Reader) (*RequestLine, error) {
	requestLineStr, err := getLineToCrlf(r)
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(requestLineStr, []byte{' '})
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected request line to be 3 parts, got %d", len(parts))
	}
	return &RequestLine{
		method:  string(parts[0]),
		target:  string(parts[1]),
		version: string(parts[2]),
	}, nil
}

func getTemplate() string {
	return "HTTP/1.1 %d %s\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: %d\r\n\r\n%s"
}

func get200Response(body string) []byte {
	return []byte(fmt.Sprintf(getTemplate(), 200, "OK", len(body), body))
}

func get404Response() []byte {
	const msg = "Not Found"
	return []byte(fmt.Sprintf(getTemplate(), 404, msg, len(msg), msg))
}

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

	reader := bufio.NewReader(conn)
	request, err := parseRequest(reader)

	if err != nil {
		fmt.Println("Error getting request: ", err.Error())
		os.Exit(1)
	}
	resp := getResponse(request.requestLine.target)

	writer := bufio.NewWriter(conn)
	_, err = writer.Write(resp)
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}
	if err := writer.Flush(); err != nil {
		fmt.Println("Error flushing to connection: ", err.Error())
		os.Exit(1)
	}
}
