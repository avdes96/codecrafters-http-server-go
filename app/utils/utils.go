package utils

import (
	"bufio"
	"bytes"
	"compress/gzip"
)

type Headers map[string]string

func GetLineToCrlf(reader *bufio.Reader) ([]byte, error) {
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

func CompressData(encodingScheme string, data []byte) ([]byte, error) {
	switch encodingScheme {
	case "gzip":
		return GzipData(data)
	default:
		return data, nil
	}
}

func GzipData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	if _, err := writer.Write(data); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil

}
