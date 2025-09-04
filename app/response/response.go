package response

import "fmt"

func Get200Response(body string, contentType CONTENT_TYPE) []byte {
	return []byte(fmt.Sprintf(getTemplate(), 200, "OK", contentType, len(body), body))
}

func Get201Response() []byte {
	const msg = "Created"
	return []byte(fmt.Sprintf(getTemplate(), 201, msg, CONTENT_TEXT_PLAIN, len(msg), msg))
}

func Get404Response() []byte {
	const msg = "Not Found"
	return []byte(fmt.Sprintf(getTemplate(), 404, msg, CONTENT_TEXT_PLAIN, len(msg), msg))
}

func Get500Response() []byte {
	const msg = "Internal Server Error"
	return []byte(fmt.Sprintf(getTemplate(), 500, msg, CONTENT_TEXT_PLAIN, len(msg), msg))

}

func Get501Response() []byte {
	const msg = "Not Implemented"
	return []byte(fmt.Sprintf(getTemplate(), 501, msg, CONTENT_TEXT_PLAIN, len(msg), msg))

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
