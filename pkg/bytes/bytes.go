package bytes

import (
	"bytes"
	"fmt"
)

//CreateByteBlock DOC TODO
func CreateByteBlock(tam int, body []byte) []byte {
	if len(body) > tam {
		fmt.Println("body muito grande para o tamanho especificado.")
		return nil
	}

	sizeDiff := tam - len(body)

	for i := 0; i < sizeDiff; i++ {
		body = append(body, []byte(" ")...)
	}

	return body
}

//ReadByteBlockAsString DOC TODO
func ReadByteBlockAsString(start int, end int, body []byte) string {
	return string(bytes.TrimSpace(body[start:end]))
}
