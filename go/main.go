package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

func EncodeString(s string) []byte {
	res := make([]byte, 4+len(s))
	binary.BigEndian.PutUint32(res, uint32(len(s)))
	copy(res[4:], s)
	return res
}

// Decode the first 4 bytes to get the length, then read that many bytes for the part
func DeserialiseString(r io.Reader) (string, error) {
	var lengthBuf [4]byte
	if _, err := r.Read(lengthBuf[:]); err != nil {
		return "", errors.New("data too short to decode length")
	}

	l := binary.BigEndian.Uint32(lengthBuf[:])
	partBuf := make([]byte, l)
	if _, err := r.Read(partBuf); err != nil {
		return "", errors.New("data too short to decode string")
	}

	return string(partBuf), nil
}

// A Part is a string that represents a literal part of the file
type Part string

func (p Part) Encode() []byte {
	return EncodeString(string(p))
}

func DeserialisePart(r io.Reader) (Part, error) {
	s, err := DeserialiseString(r)
	return Part(s), err
}

// A Name is a string that represents a section of the file that can be substituted with a value
type Name string

func (n Name) Encode() []byte {
	return EncodeString(string(n))
}

func DeserialiseName(r io.Reader) (Name, error) {
	s, err := DeserialiseString(r)
	return Name(s), err
}

// A PartName is a Part followed by a Name
type PartName struct {
	Part
	Name
}

func (pn PartName) Encode() []byte {
	return append(pn.Part.Encode(), pn.Name.Encode()...)
}

func DeserialisePartName(r io.Reader) (pn PartName, err error) {
	if pn.Part, err = DeserialisePart(r); err != nil {
		return pn, fmt.Errorf("decode part: %w", err)
	}
	if pn.Name, err = DeserialiseName(r); err != nil {
		return pn, fmt.Errorf("decode name: %w", err)
	}

	return
}

type ToSub map[string]string

type Sub struct {
	PartNames []PartName
	Final     Part
}

func (s Sub) Sub(to ToSub) (string, error) {
	var b strings.Builder

	for _, pn := range s.PartNames {
		n := string(pn.Name)

		b.WriteString(string(pn.Part))
		val, ok := to[n]
		if !ok {
			return "", fmt.Errorf("missing value for name %q", pn.Name)
		}
		b.WriteString(val)

		delete(to, n)
	}

	if len(to) > 0 {
		return "", fmt.Errorf("extra values provided for names: %v", to)
	}

	b.WriteString(string(s.Final))

	return b.String(), nil
}

func (s Sub) Encode() []byte {
	res := make([]byte, 4)
	binary.BigEndian.PutUint32(res, uint32(len(s.PartNames)))
	for _, pn := range s.PartNames {
		res = append(res, pn.Encode()...)
	}

	return append(res, s.Final.Encode()...)
}

func DeserialiseSub(r io.Reader) (s Sub, err error) {
	var lengthBuf [4]byte
	if _, err = r.Read(lengthBuf[:]); err != nil {
		return s, errors.New("data too short to decode length")
	}

	l := binary.BigEndian.Uint32(lengthBuf[:])
	s.PartNames = make([]PartName, l)
	for i := range l {
		if s.PartNames[i], err = DeserialisePartName(r); err != nil {
			return s, fmt.Errorf("decode partname %d: %w", i, err)
		}
	}

	if s.Final, err = DeserialisePart(r); err != nil {
		return s, fmt.Errorf("decode final part: %w", err)
	}

	return
}

func main() {
	s := Sub{
		PartNames: []PartName{
			{Part: "Hello ", Name: "name"},
			{Part: "! You have ", Name: "count"},
		},
		Final: " new messages.",
	}

	fmt.Println(s)

	toSub := ToSub{
		"name":  "Heliodex",
		"count": "67",
	}

	result, err := s.Sub(toSub)
	if err != nil {
		fmt.Printf("Error substituting: %v\n", err)
		return
	}

	fmt.Printf("Substituted result: %s\n", result)

	data := s.Encode()
	fmt.Printf("Encoded data: %s\n", data)
}
