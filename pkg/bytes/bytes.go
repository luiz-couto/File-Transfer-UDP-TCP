package bytes

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

/*
CreateByteBlock receives an array of bytes and returns a block of the given size
with that information. It fills the rest of the bytes with spaces
*/
func CreateByteBlock(tam int, body []byte) []byte {
	if len(body) > tam {
		fmt.Println(tam)
		fmt.Println(string(body))
		fmt.Println("body muito grande para o tamanho especificado.")
		return nil
	}

	sizeDiff := tam - len(body)

	for i := 0; i < sizeDiff; i++ {
		body = append(body, []byte(" ")...)
	}

	return body
}

/*
ReadByteBlockAsString reads a block of bytes as a string
*/
func ReadByteBlockAsString(start int, end int, body []byte) string {
	return string(bytes.TrimSpace(body[start:end]))
}

/*
WriteIntAsBytes returns the byte representation of an integer value
*/
func WriteIntAsBytes(tam, i int) []byte {
	bs := make([]byte, tam)

	switch tam {
	case 2:
		binary.LittleEndian.PutUint16(bs, uint16(i))
	case 4:
		binary.LittleEndian.PutUint32(bs, uint32(i))
	case 8:
		binary.LittleEndian.PutUint64(bs, uint64(i))
	}

	return bs
}

/*
ReadByteBlockAsInt reads a block of bytes as an integer
*/
func ReadByteBlockAsInt(start int, end int, body []byte) int {
	tam := end - start
	buf := bytes.NewReader(body[start:end])

	switch tam {
	case 2:
		var i uint16
		err := binary.Read(buf, binary.LittleEndian, &i)
		if err != nil {
			fmt.Println("binary read failed:", err)
		}
		return int(i)
	case 4:
		var i uint32
		err := binary.Read(buf, binary.LittleEndian, &i)
		if err != nil {
			fmt.Println("binary read failed:", err)
		}
		return int(i)
	case 8:
		var i uint64
		err := binary.Read(buf, binary.LittleEndian, &i)
		if err != nil {
			fmt.Println("binary read failed:", err)
		}
		return int(i)
	}
	return 0
}

/*
DivideInPackages receives an array of bytes and splits it into several pkgSize size packets
*/
func DivideInPackages(content []byte, pkgSize int) map[int][]byte {
	var divided [][]byte

	for i := 0; i < len(content); i += pkgSize {
		end := i + pkgSize

		if end > len(content) {
			end = len(content)
		}

		divided = append(divided, content[i:end])
	}

	pkgs := make(map[int][]byte)
	for idx, pkg := range divided {
		pkgs[idx] = pkg
	}

	return pkgs
}
