package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

func readExact(r io.Reader, buf []byte) error {
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	if n < len(buf) {
		return errors.New("not enough bytes to read")
	}
	return nil
}

func serialiseString(s string) []byte {
	res := make([]byte, 4+len(s))
	binary.BigEndian.PutUint32(res, uint32(len(s)))
	copy(res[4:], s)
	return res
}

var (
	ErrDataTooShort       = errors.New("data too short to decode")
	ErrDataTooShortLength = fmt.Errorf("%w length", ErrDataTooShort)
	ErrDataTooShortString = fmt.Errorf("%w string", ErrDataTooShort)
)

func getLen(r io.Reader) (uint32, error) {
	var lengthBuf [4]byte
	if readExact(r, lengthBuf[:]) != nil {
		return 0, ErrDataTooShortLength
	}
	return binary.BigEndian.Uint32(lengthBuf[:]), nil
}

// Decode the first 4 bytes to get the length, then read that many bytes for the part
func deserialiseString(r io.Reader) (string, error) {
	l, err := getLen(r)
	if err != nil {
		return "", err
	}

	strBuf := make([]byte, l)
	if readExact(r, strBuf) != nil {
		return "", ErrDataTooShortString
	}

	return string(strBuf), nil
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

var (
	ErrMissingValue = errors.New("missing value for name")
	ErrExtraValue   = errors.New("extra value provided for name")
)

func (s Sub) Sub(to ToSub) (string, error) {
	seen := make(map[string]struct{})
	for _, pn := range s.PartNames {
		n := string(pn.Name)
		if _, ok := to[n]; !ok {
			return "", fmt.Errorf("%w: %q", ErrMissingValue, n)
		}
		seen[n] = struct{}{}
	}

	for k := range to {
		if _, ok := seen[k]; !ok {
			return "", fmt.Errorf("%w: %q", ErrExtraValue, k)
		}
	}

	var b strings.Builder

	for _, pn := range s.PartNames {
		val := to[string(pn.Name)]

		b.WriteString(string(pn.Part))
		b.WriteString(val)
	}

	b.WriteString(string(s.Final))

	return b.String(), nil
}

func (s Sub) Equals(other Sub) bool {
	if len(s.PartNames) != len(other.PartNames) {
		return false
	}

	for i, v := range s.PartNames {
		if v != other.PartNames[i] {
			return false
		}
	}

	return s.Final == other.Final
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
	l, err := getLen(r)
	if err != nil {
		return s, err
	}

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
