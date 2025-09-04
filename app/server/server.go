package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
)

type Server struct {
	filesDirectory string
}

func NewServer(filesDirectory string) *Server {
	return &Server{filesDirectory: filesDirectory}
}

func (s *Server) Run() {
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
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	request, err := request.ParseRequest(reader)

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

func (s *Server) getResponse(request *request.Request) []byte {
	target := request.Target()
	if target == "/user-agent" {
		return response.Get200Response(request.HeaderValue("user-agent"), response.CONTENT_TEXT_PLAIN)
	}
	if strings.HasPrefix(target, "/files/") {
		return s.filesEndpoint(request)
	}
	if target == "/" {
		return response.Get200Response("", response.CONTENT_TEXT_PLAIN)
	}
	if strings.HasPrefix(target, "/echo/") {
		return response.Get200Response(strings.TrimPrefix(target, "/echo/"), response.CONTENT_TEXT_PLAIN)
	}
	return response.Get404Response()
}

func (s *Server) filesEndpoint(request *request.Request) []byte {
	if s.filesDirectory == "" {
		return response.Get404Response()
	}
	filename := strings.TrimPrefix(request.Target(), "/files/")
	location := filepath.Join(s.filesDirectory, filename)
	switch strings.ToLower(request.Method()) {
	case "get":
		data, err := os.ReadFile(location)
		if err != nil {
			if os.IsNotExist(err) {
				return response.Get404Response()
			}
			return response.Get500Response()
		}
		return response.Get200Response(string(data), response.CONTENT_APP_OCTET_STREAM)
	case "post":
		err := os.WriteFile(location, request.Body(), 0644)
		if err != nil {
			return response.Get500Response()
		}
		return response.Get201Response()
	default:
		return response.Get501Response()
	}
}
