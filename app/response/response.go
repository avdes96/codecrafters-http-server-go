package response

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/http-server-starter-go/app/utils"
)

type Response struct {
	version    string
	statusCode STATUS_CODE
	headers    utils.Headers
	body       string
}

func (r *Response) Serialise() []byte {
	body := r.getBodyAsBytes()
	serialised := []byte(fmt.Sprintf("HTTP/%s %d %s\r\n", r.version, r.statusCode, r.statusCode))
	for k, v := range r.headers {
		serialised = append(serialised, []byte(fmt.Sprintf("%s: %s\r\n", k, v))...)
	}
	serialised = append(serialised, []byte("\r\n")...)
	serialised = append(serialised, body...)
	return serialised

}

func (r *Response) getBodyAsBytes() []byte {
	bodyBytes := []byte{}
	doneEncoding := false
	if scheme, ok := r.headers["Content-Encoding"]; ok {
		if compressedBytes, err := utils.CompressData(scheme, []byte(r.body)); err == nil {
			bodyBytes = compressedBytes
			doneEncoding = true
		}
	}
	if !doneEncoding {
		bodyBytes = []byte(r.body)
	}
	if len(bodyBytes) > 0 {
		r.headers["Content-Length"] = strconv.Itoa(len(bodyBytes))
	}
	return bodyBytes
}

type STATUS_CODE int

const (
	CODE_200 = 200
	CODE_201 = 201
	CODE_404 = 404
	CODE_500 = 500
	CODE_501 = 501
)

func (c STATUS_CODE) String() string {
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

func WithStatusCode(s STATUS_CODE) Option {
	return func(r *Response) {
		r.statusCode = s
	}
}

func WithContentType(c CONTENT_TYPE) Option {
	return func(r *Response) {
		r.headers["Content-Type"] = c.String()
	}
}

func WithBody(b string) Option {
	return func(r *Response) {
		r.body = b
	}
}

func WithEncodings(e string) Option {
	return func(r *Response) {
		if e == "" {
			return
		}
		r.headers["Content-Encoding"] = e
	}
}

func NewResponse(opts ...Option) *Response {
	r := &Response{
		version:    DEFAULT_VERSION,
		statusCode: CODE_500,
		headers:    make(utils.Headers),
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}
