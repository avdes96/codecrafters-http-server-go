package response

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/http-server-starter-go/app/utils"
)

type Response struct {
	version    string
	statusCode RESPONSE_CODE
	headers    utils.Headers
	body       string
}

func (r *Response) Serialise() []byte {
	s := fmt.Sprintf("HTTP/%s %d %s\r\n", r.version, r.statusCode, r.statusCode)
	for k, v := range r.headers {
		s += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	s += "\r\n"
	s += r.body
	return []byte(s)
}

type RESPONSE_CODE int

const (
	CODE_200 = 200
	CODE_201 = 201
	CODE_404 = 404
	CODE_500 = 500
	CODE_501 = 501
)

func (c RESPONSE_CODE) String() string {
	switch c {
	case CODE_200:
		return "OK"
	case CODE_201:
		return "Created"
	case CODE_404:
		return "Not Found"
	case CODE_500:
		return "Internal Server Error"
	case CODE_501:
		return "Not Implemented"
	default:
		return "Unknown Code"
	}
}

const DEFAULT_VERSION = "1.1"

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

type Option func(*Response)

func WithContentType(c CONTENT_TYPE) Option {
	return func(r *Response) {
		r.headers["Content-Type"] = c.String()
	}
}

func WithBody(b string) Option {
	return func(r *Response) {
		r.headers["Content-Length"] = strconv.Itoa(len(b))
		r.body = b
	}
}

func New200Response(opts ...Option) *Response {
	r := &Response{
		version:    DEFAULT_VERSION,
		statusCode: CODE_200,
		headers:    make(utils.Headers),
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

func New201Response() *Response {
	return &Response{
		version:    DEFAULT_VERSION,
		statusCode: CODE_201,
		headers:    make(utils.Headers),
	}
}

func New404Response() *Response {
	return &Response{
		version:    DEFAULT_VERSION,
		statusCode: CODE_404,
		headers:    make(utils.Headers),
	}
}

func New500Response() *Response {
	return &Response{
		version:    DEFAULT_VERSION,
		statusCode: CODE_500,
		headers:    make(utils.Headers),
	}
}

func New501Response() *Response {
	return &Response{
		version:    DEFAULT_VERSION,
		statusCode: CODE_501,
		headers:    make(utils.Headers),
	}
}
