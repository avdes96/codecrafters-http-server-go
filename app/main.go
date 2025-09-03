package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

type Headers map[string]string

type Request struct {
	requestLine *RequestLine
	headers     Headers
}

type RequestLine struct {
	method  string
	target  string
	version string
}

func getResponse(request *Request) []byte {
	target := request.requestLine.target
	if target == "/user-agent" {
		return get200Response(request.headers["user-agent"])
	}
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
	headers, err := parseHeaders(reader)
	if err != nil {
		return nil, err
	}
	return &Request{requestLine: rl, headers: headers}, nil
}

func parseRequestLine(reader *bufio.Reader) (*RequestLine, error) {
	requestLineStr, err := getLineToCrlf(reader)
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

func parseHeaders(reader *bufio.Reader) (Headers, error) {
	headers := make(Headers)
	for {
		line, err := getLineToCrlf(reader)
		if err != nil {
			return nil, err
		}
		if len(line) == 0 {
			break
		}
		parts := bytes.SplitN(line, []byte(": "), 2)
		key := strings.ToLower(string(parts[0]))
		val := string(parts[1])
		headers[key] = val
	}
	return headers, nil
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
	resp := getResponse(request)

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
