package utils

import "bufio"

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
