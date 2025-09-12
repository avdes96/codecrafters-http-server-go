package request

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/utils"
)

type Request struct {
	requestLine *RequestLine
	headers     utils.Headers
	body        []byte
}

type RequestLine struct {
	method  string
	target  string
	version string
}

func ParseRequest(reader *bufio.Reader) (*Request, error) {
	rl, err := ParseRequestLine(reader)
	if err != nil {
		return nil, err
	}
	headers, err := ParseHeaders(reader)
	if err != nil {
		return nil, err
	}
	body := []byte{}
	if bodyLenStr, ok := headers[strings.ToLower("Content-Length")]; ok {
		if bodyLen, err := strconv.Atoi(bodyLenStr); err == nil && bodyLen > 0 {
			body, err = ParseBody(reader, bodyLen)
			if err != nil {
				return nil, err
			}
		}
	}
	return &Request{requestLine: rl, headers: headers, body: body}, nil
}

func ParseRequestLine(reader *bufio.Reader) (*RequestLine, error) {
	requestLineStr, err := utils.GetLineToCrlf(reader)
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(requestLineStr, []byte{' '})
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected request line to be 3 parts, got %d", len(parts))
	}
	return &RequestLine{
		method:  strings.ToLower(string(parts[0])),
		target:  string(parts[1]),
		version: string(parts[2]),
	}, nil
}

func ParseHeaders(reader *bufio.Reader) (utils.Headers, error) {
	headers := make(utils.Headers)
	for {
		line, err := utils.GetLineToCrlf(reader)
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

func ParseBody(reader *bufio.Reader, len int) ([]byte, error) {
	body := make([]byte, len)
	_, err := reader.Read(body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (r *Request) Target() string {
	return r.requestLine.target
}

func (r *Request) Method() string {
	return r.requestLine.method
}

func (r *Request) Body() []byte {
	return r.body
}

func (r *Request) HeaderValue(key string) (string, bool) {
	if value, ok := r.headers[strings.ToLower(key)]; ok {
		return value, true
	}
	return "", false
}

func (r *Request) Encodings() []string {
	encodingStr, ok := r.HeaderValue("accept-encoding")
	if !ok {
		return nil
	}
	return strings.Split(encodingStr, ", ")
}

func (r *Request) Endpoint() string {
	idx := 0
	for idx < len(r.Target()) {
		if r.Target()[idx] == '/' && idx > 0 {
			break
		}
		idx++
	}
	return r.Target()[:idx]
}
