package config

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// OrderedEntry is one key/value pair from a JSON object, in declaration
// order.
type OrderedEntry struct {
	Key   string
	Value json.RawMessage
}

// OrderedObject decodes a JSON object while preserving the declaration order
// of its keys — standard map[string]T decoding loses this order, which
// would violate FR-018/FR-011's requirement that `contexts[].models` and
// `fields` be processed in the order the developer wrote them (see
// research.md "Declaration-order determinism").
type OrderedObject []OrderedEntry

func (o *OrderedObject) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("expected JSON object, got %v", tok)
	}

	var entries OrderedObject
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		// Token() guarantees a string for object member names once it
		// returns without error, so no type-assertion check is needed here.
		key := keyTok.(string)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return fmt.Errorf("decoding value for key %q: %w", key, err)
		}
		entries = append(entries, OrderedEntry{Key: key, Value: raw})
	}
	if _, err := dec.Token(); err != nil { // consume closing '}'
		return err
	}
	*o = entries
	return nil
}

// DecodeStrict decodes v from data, rejecting any JSON object field with no
// corresponding struct field (the equivalent of the config schema's
// `additionalProperties: false`).
func DecodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
