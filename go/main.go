package main

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type Part string

func (p Part) Encode() string {
	res := make([]byte, 4+len(p))
	binary.BigEndian.PutUint32(res, uint32(len(p)))
	copy(res[4:], p)
	return string(res)
}

// Decode the first 4 bytes to get the length, then read that many bytes for the part
func DecodePart(data []byte) (Part, []byte, error) {
	if len(data) < 4 {
		return "", nil, errors.New("data too short to decode length")
	}

	l := binary.BigEndian.Uint32(data)
	if len(data) < 4+int(l) {
		return "", nil, errors.New("data too short to decode part")
	}

	return Part(string(data[4 : 4+l])), data[4+l:], nil
}

type Sub []Part

func MakeSub(parts ...Part) Sub {
	return Sub(parts)
}

func main() {
	fmt.Println("sup")
	s := MakeSub("a:", " b:", " c:")
	fmt.Printf("%+v\n", s)
}
