package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Headers map[string]string

type Server struct {
	filesDirectory string
}

func NewServer(filesDirectory string) *Server {
	return &Server{filesDirectory: filesDirectory}
}

type Request struct {
	requestLine *RequestLine
	headers     Headers
	body        []byte
}

type RequestLine struct {
	method  string
	target  string
	version string
}

func (s *Server) getResponse(request *Request) []byte {
	target := request.requestLine.target
	if target == "/user-agent" {
		return get200Response(request.headers["user-agent"], CONTENT_TEXT_PLAIN)
	}
	if strings.HasPrefix(target, "/files/") {
		return s.filesEndpoint(request)
	}
	if target == "/" {
		return get200Response("", CONTENT_TEXT_PLAIN)
	}
	if strings.HasPrefix(target, "/echo/") {
		return get200Response(strings.TrimPrefix(target, "/echo/"), CONTENT_TEXT_PLAIN)
	}
	return get404Response()
}

func (s *Server) filesEndpoint(request *Request) []byte {
	if s.filesDirectory == "" {
		return get404Response()
	}
	filename := strings.TrimPrefix(request.requestLine.target, "/files/")
	location := filepath.Join(s.filesDirectory, filename)
	switch strings.ToLower(request.requestLine.method) {
	case "get":
		data, err := os.ReadFile(location)
		if err != nil {
			if os.IsNotExist(err) {
				return get404Response()
			}
			return get500Response()
		}
		return get200Response(string(data), CONTENT_APP_OCTET_STREAM)
	case "post":
		err := os.WriteFile(location, request.body, 0644)
		if err != nil {
			return get500Response()
		}
		return get201Response()
	default:
		return get501Response()
	}

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
	body := []byte{}
	if bodyLenStr, ok := headers[strings.ToLower("Content-Length")]; ok {
		if bodyLen, err := strconv.Atoi(bodyLenStr); err == nil && bodyLen > 0 {
			body, err = parseBody(reader, bodyLen)
			if err != nil {
				return nil, err
			}
		}
	}
	return &Request{requestLine: rl, headers: headers, body: body}, nil
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

func parseBody(reader *bufio.Reader, len int) ([]byte, error) {
	body := make([]byte, len)
	_, err := reader.Read(body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func getTemplate() string {
	return "HTTP/1.1 %d %s\r\n" +
		"Content-Type: %s\r\n" +
		"Content-Length: %d\r\n\r\n%s"
}

type CONTENT_TYPE int

const (
	CONTENT_TEXT_PLAIN CONTENT_TYPE = iota
	CONTENT_APP_OCTET_STREAM
)

func (c CONTENT_TYPE) String() string {
	switch c {
	case CONTENT_TEXT_PLAIN:
		return "text/plain"
	case CONTENT_APP_OCTET_STREAM:
		return "application/octet-stream"
	default:
		return "Unknown"
	}
}

func get200Response(body string, contentType CONTENT_TYPE) []byte {
	return []byte(fmt.Sprintf(getTemplate(), 200, "OK", contentType, len(body), body))
}

func get201Response() []byte {
	const msg = "Created"
	return []byte(fmt.Sprintf(getTemplate(), 201, msg, CONTENT_TEXT_PLAIN, len(msg), msg))
}

func get404Response() []byte {
	const msg = "Not Found"
	return []byte(fmt.Sprintf(getTemplate(), 404, msg, CONTENT_TEXT_PLAIN, len(msg), msg))
}

func get500Response() []byte {
	const msg = "Internal Server Error"
	return []byte(fmt.Sprintf(getTemplate(), 500, msg, CONTENT_TEXT_PLAIN, len(msg), msg))

}

func get501Response() []byte {
	const msg = "Not Implemented"
	return []byte(fmt.Sprintf(getTemplate(), 501, msg, CONTENT_TEXT_PLAIN, len(msg), msg))

}

func (s *Server) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	request, err := parseRequest(reader)

	if err != nil {
		fmt.Println("Error getting request: ", err.Error())
		os.Exit(1)
	}
	resp := s.getResponse(request)

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

func main() {
	directory := flag.String("directory", "", "the absolute path of the directory where files are stored")
	flag.Parse()
	server := NewServer(*directory)

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
		go server.handleConnection(conn)
	}

}
