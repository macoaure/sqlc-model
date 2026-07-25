package unit

import (
	"encoding/json"
	"testing"

	"github.com/macoaure/sqlc-gen-richmodel/internal/config"
)

// json.Unmarshal/Decoder only ever hand a json.Unmarshaler syntactically
// valid raw bytes for the sub-value it bounds, so these malformed-input
// branches can only be exercised by invoking UnmarshalJSON directly, as any
// caller holding an OrderedObject value is free to do.

func TestOrderedObject_UnmarshalJSON_EmptyInputErrors(t *testing.T) {
	var oo config.OrderedObject
	if err := oo.UnmarshalJSON([]byte(``)); err == nil {
		t.Fatal("expected an error decoding empty input")
	}
}

func TestOrderedObject_UnmarshalJSON_NotAnObjectErrors(t *testing.T) {
	var oo config.OrderedObject
	if err := oo.UnmarshalJSON([]byte(`5`)); err == nil {
		t.Fatal("expected an error when the top-level value is not an object")
	}
}

func TestOrderedObject_UnmarshalJSON_MalformedKeyTokenErrors(t *testing.T) {
	var oo config.OrderedObject
	if err := oo.UnmarshalJSON([]byte(`{,"a":1}`)); err == nil {
		t.Fatal("expected an error reading a malformed key token")
	}
}

func TestOrderedObject_UnmarshalJSON_MalformedValueErrors(t *testing.T) {
	var oo config.OrderedObject
	if err := oo.UnmarshalJSON([]byte(`{"a":}`)); err == nil {
		t.Fatal("expected an error decoding a malformed value")
	}
}

func TestOrderedObject_UnmarshalJSON_UnterminatedObjectErrors(t *testing.T) {
	var oo config.OrderedObject
	if err := oo.UnmarshalJSON([]byte(`{"a":1`)); err == nil {
		t.Fatal("expected an error for an unterminated object")
	}
}

func TestOrderedObject_UnmarshalJSON_PreservesOrder(t *testing.T) {
	var oo config.OrderedObject
	if err := json.Unmarshal([]byte(`{"z":1,"a":2}`), &oo); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(oo) != 2 || oo[0].Key != "z" || oo[1].Key != "a" {
		t.Fatalf("expected order-preserving decode, got %+v", oo)
	}
}
