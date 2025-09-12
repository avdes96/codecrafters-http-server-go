package server

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/request"
	"github.com/codecrafters-io/http-server-starter-go/app/response"
)

type requestHandler func(*request.Request, []response.Option, *Server) []response.Option

type endpointRegistry map[string]map[string]requestHandler

func (r endpointRegistry) register(path string, method string, rh requestHandler) {
	if r[path] == nil {
		r[path] = make(map[string]requestHandler)
	}
	methodLower := strings.ToLower(method)
	r[path][methodLower] = rh
}

func NewEndpointRegistry() endpointRegistry {
	e := make(endpointRegistry)
	e.register("/files", "get", func(request *request.Request, opts []response.Option, server *Server) []response.Option {
		if server.filesDirectory == "" {
			opts = append(opts, response.WithStatusCode(response.CODE_404))
			return opts
		}
		filename := strings.TrimPrefix(request.Target(), "/files/")
		location := filepath.Join(server.filesDirectory, filename)
		data, err := os.ReadFile(location)
		if err != nil {
			if os.IsNotExist(err) {
				opts = append(opts, response.WithStatusCode(response.CODE_404))
				return opts
			}
			opts = append(opts, response.WithStatusCode(response.CODE_500))
			return opts
		}
		opts = append(opts, response.WithStatusCode(response.CODE_200))
		opts = append(opts, response.WithBody(string(data)))
		opts = append(opts, response.WithContentType(response.CONTENT_APP_OCTET_STREAM))
		return opts
	})
	e.register("/files", "post", func(request *request.Request, opts []response.Option, server *Server) []response.Option {
		if server.filesDirectory == "" {
			opts = append(opts, response.WithStatusCode(response.CODE_404))
			return opts
		}
		filename := strings.TrimPrefix(request.Target(), "/files/")
		location := filepath.Join(server.filesDirectory, filename)
		err := os.WriteFile(location, request.Body(), 0644)
		if err != nil {
			opts = append(opts, response.WithStatusCode(response.CODE_500))
			return opts
		}
		opts = append(opts, response.WithStatusCode(response.CODE_201))
		return opts
	})
	e.register("/", "get", func(request *request.Request, opts []response.Option, server *Server) []response.Option {
		opts = append(opts, response.WithStatusCode(response.CODE_200))
		opts = append(opts, response.WithContentType(response.CONTENT_TEXT_PLAIN))
		return opts
	})
	e.register("/user-agent", "get", func(request *request.Request, opts []response.Option, server *Server) []response.Option {
		opts = append(opts, response.WithStatusCode(response.CODE_200))
		opts = append(opts, response.WithBody(request.HeaderValue("user-agent")))
		opts = append(opts, response.WithContentType(response.CONTENT_TEXT_PLAIN))
		return opts
	})
	e.register("/echo", "get", func(request *request.Request, opts []response.Option, server *Server) []response.Option {
		opts = append(opts, response.WithStatusCode(response.CODE_200))
		opts = append(opts, response.WithBody(strings.TrimPrefix(request.Target(), "/echo/")))
		opts = append(opts, response.WithContentType(response.CONTENT_TEXT_PLAIN))
		return opts
	})

	return e
}
