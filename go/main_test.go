package main_test

import (
	"bytes"
	"errors"
	"testing"

	sub "github.com/Heliodex/substitution"
)

// Test multiple substitutions of the same name
var s = sub.Sub{
	PartNames: []sub.PartName{
		{Part: "Hello ", Name: "name"},
		{Part: "! You have ", Name: "count"},
		{Part: " new messages. Thanks, ", Name: "name"},
	},
	Final: "!",
}

func TestSubstitute(t *testing.T) {
	toSub := sub.ToSub{
		"name":  "Heliodex",
		"count": "67",
	}

	result, err := s.Sub(toSub)
	if err != nil {
		t.Fatal(err)
	}

	if result != "Hello Heliodex! You have 67 new messages. Thanks, Heliodex!" {
		t.Fatalf("unexpected result: %s", result)
	}
}

func TestSubstituteMissing(t *testing.T) {
	toSub := sub.ToSub{
		"name": "Heliodex",
	}

	result, err := s.Sub(toSub)
	if err == nil {
		t.Fatalf("expected error, got result: %s", result)
	}
	if !errors.Is(err, sub.ErrMissingValue) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubstituteExtra(t *testing.T) {
	toSub := sub.ToSub{
		"name":  "Heliodex",
		"count": "67",
		"extra": "value",
	}

	result, err := s.Sub(toSub)
	if err == nil {
		t.Fatalf("expected error, got result: %s", result)
	}
	if !errors.Is(err, sub.ErrExtraValue) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSerialisation(t *testing.T) {
	data := s.Serialise()

	s2, err := sub.DeserialiseSub(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if !s.Equals(s2) {
		t.Fatalf("deserialisation mismatch: expected %q, got %q", s.Final, s2.Final)
	}
}

func TestDeserialisationTooShortLength(t *testing.T) {
	s, err := sub.DeserialiseSub(bytes.NewReader([]byte{0, 0, 0}))
	if err == nil {
		t.Fatalf("expected error, got s: %v", s)
	}
	if !errors.Is(err, sub.ErrDataTooShortLength) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeserialisationTooShortString(t *testing.T) {
	_, err := sub.DeserialiseSub(bytes.NewReader([]byte{0, 0, 0, 1, 0, 0, 0, 1}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sub.ErrDataTooShortString) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEquals(t *testing.T) {
	s2 := sub.Sub{
		PartNames: []sub.PartName{
			{Part: "Hello ", Name: "name"},
			{Part: "! You have ", Name: "count"},
			{Part: " new messages. Thanks, ", Name: "name"},
		},
		Final: "!",
	}

	if !s.Equals(s2) {
		t.Fatal("expected subs to be equal")
	}

	s3 := sub.Sub{
		PartNames: []sub.PartName{
			{Part: "Sup ", Name: "name"},
			{Part: ", you got ", Name: "num"},
			{Part: " new pings. Thanks, ", Name: "name"},
		},
		Final: "!",
	}

	if s.Equals(s3) {
		t.Fatal("expected subs to not be equal")
	}
}
