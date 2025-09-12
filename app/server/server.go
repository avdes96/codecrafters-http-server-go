package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
	mapset "github.com/deckarep/golang-set/v2"
)

type Server struct {
	filesDirectory string
	validEncodings mapset.Set[string]
	endpoints      endpointRegistry
}

func NewServer(filesDirectory string) *Server {
	validEncodings := mapset.NewSet("gzip")
	return &Server{
		filesDirectory: filesDirectory,
		validEncodings: validEncodings,
		endpoints:      NewEndpointRegistry(),
	}
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
	for {
		reader := bufio.NewReader(conn)
		request, err := request.ParseRequest(reader)

		if err != nil {
			if err == io.EOF {
				continue
			}
			fmt.Println("Error getting request: ", err.Error())
			os.Exit(1)
		}
		resp := s.getResponse(request)

		writer := bufio.NewWriter(conn)
		_, err = writer.Write(resp.Serialise())
		if err != nil {
			fmt.Println("Error writing to connection: ", err.Error())
			os.Exit(1)
		}
		if err := writer.Flush(); err != nil {
			fmt.Println("Error flushing to connection: ", err.Error())
			os.Exit(1)
		}
	}
}

func (s *Server) getResponse(request *request.Request) *response.Response {
	opts := []response.Option{}
	opts = s.addEncodingOption(request, opts)
	availableMethods, ok := s.endpoints[request.Endpoint()]
	if !ok {
		opts = append(opts, response.WithStatusCode(response.CODE_404))
		return response.NewResponse(opts...)
	}
	rh, ok := availableMethods[request.Method()]
	if !ok {
		opts = append(opts, response.WithStatusCode(response.CODE_501))
		return response.NewResponse(opts...)
	}
	opts = rh(request, opts, s)
	return response.NewResponse(opts...)
}

func (s *Server) addEncodingOption(r *request.Request, opts []response.Option) []response.Option {
	encodings := r.Encodings()
	if encodings != nil {
		enc := s.chooseEncoding(encodings)
		if enc != "" {
			opts = append(opts, response.WithEncodings(enc))
		}
	}
	return opts
}

func (s *Server) chooseEncoding(encodingList []string) string {
	for _, encoding := range encodingList {
		if s.validEncodings.Contains(encoding) {
			return encoding
		}
	}
	return ""
}
