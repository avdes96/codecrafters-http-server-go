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

func getCrlfScanner(conn *net.Conn) *bufio.Scanner {
	scanner := bufio.NewScanner(*conn)
	cap := 64 * 1024
	buf := make([]byte, cap)
	scanner.Buffer(buf, cap)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			if len(data) > 0 {
				return len(data), data, nil
			}
			return 0, nil, nil
		}
		if i := bytes.Index(data, []byte("\r\n")); i > 0 {
			return i + 2, data[:i], nil
		}
		return 0, nil, nil
	})
	return scanner
}

func parseRequestLine(s string) (*RequestLine, error) {
	parts := strings.Split(s, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected request line to be 3 parts, got %d", len(parts))
	}
	return &RequestLine{
		method:  parts[0],
		target:  parts[1],
		version: parts[2],
	}, nil
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

func parseRequest(s *bufio.Scanner) (*Request, error) {
	s.Scan()
	if s.Err() != nil {
		return nil, s.Err()
	}
	requestLineStr := s.Text()
	rl, err := parseRequestLine(requestLineStr)
	if err != nil {
		return nil, err
	}
	return &Request{requestLine: rl}, nil
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

	scanner := getCrlfScanner(&conn)
	request, err := parseRequest(scanner)

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
