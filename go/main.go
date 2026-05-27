package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

func serialiseString(s string) []byte {
	res := make([]byte, 4+len(s))
	binary.BigEndian.PutUint32(res, uint32(len(s)))
	copy(res[4:], s)
	return res
}

// Decode the first 4 bytes to get the length, then read that many bytes for the part
func deserialiseString(r io.Reader) (string, error) {
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

// A part is a string that represents a literal part of the file
type part string

func (p part) Serialise() []byte {
	return serialiseString(string(p))
}

func deserialisePart(r io.Reader) (part, error) {
	s, err := deserialiseString(r)
	return part(s), err
}

// A name is a string that represents a section of the file that can be substituted with a value
type name string

func (n name) serialise() []byte {
	return serialiseString(string(n))
}

func deserialiseName(r io.Reader) (name, error) {
	s, err := deserialiseString(r)
	return name(s), err
}

// A PartName is a Part followed by a Name
type PartName struct {
	Part part
	Name name
}

func (pn PartName) serialise() []byte {
	return append(pn.Part.Serialise(), pn.Name.serialise()...)
}

func deserialisePartName(r io.Reader) (pn PartName, err error) {
	if pn.Part, err = deserialisePart(r); err != nil {
		return pn, fmt.Errorf("decode part: %w", err)
	}
	if pn.Name, err = deserialiseName(r); err != nil {
		return pn, fmt.Errorf("decode name: %w", err)
	}

	return
}

type ToSub map[string]string

type Sub struct {
	PartNames []PartName
	Final     part
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

func (s Sub) Serialise() []byte {
	res := make([]byte, 4)
	binary.BigEndian.PutUint32(res, uint32(len(s.PartNames)))
	for _, pn := range s.PartNames {
		res = append(res, pn.serialise()...)
	}

	return append(res, s.Final.Serialise()...)
}

func DeserialiseSub(r io.Reader) (s Sub, err error) {
	var lengthBuf [4]byte
	if _, err = r.Read(lengthBuf[:]); err != nil {
		return s, errors.New("data too short to decode length")
	}

	l := binary.BigEndian.Uint32(lengthBuf[:])
	s.PartNames = make([]PartName, l)
	for i := range l {
		if s.PartNames[i], err = deserialisePartName(r); err != nil {
			return s, fmt.Errorf("decode partname %d: %w", i, err)
		}
	}

	if s.Final, err = deserialisePart(r); err != nil {
		return s, fmt.Errorf("decode final part: %w", err)
	}

	return
}
