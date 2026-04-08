package handlers

import (
	"encoding/json"

	"github.com/go-faster/jx"

	"github.com/gg-ecommerce/backend/api/gen"
)

// marshalAnyObject marshals any value to gen.AnyObject (map[string]jx.Raw).
// Returns empty AnyObject on error.
func marshalAnyObject(v interface{}) gen.AnyObject {
	b, err := json.Marshal(v)
	if err != nil {
		return gen.AnyObject{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return gen.AnyObject{}
	}
	out := make(gen.AnyObject, len(m))
	for k, val := range m {
		b2, _ := json.Marshal(val)
		out[k] = jx.Raw(b2)
	}
	return out
}

// marshalList converts a slice to []gen.AnyObject using marshalAnyObject.
func marshalList[T any](items []T) []gen.AnyObject {
	out := make([]gen.AnyObject, 0, len(items))
	for i := range items {
		out = append(out, marshalAnyObject(items[i]))
	}
	return out
}

// unmarshalAnyObject converts gen.AnyObject to a map then unmarshals into target.
func unmarshalAnyObject(src gen.AnyObject, target interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

// optNilStringVal extracts string from gen.OptNilString.
func optNilStringVal(s gen.OptNilString) string {
	if !s.Set {
		return ""
	}
	return s.Value
}

// ok returns &gen.MutationResult{Success: true}.
func ok() *gen.MutationResult {
	return &gen.MutationResult{Success: true}
}

// optBool extracts bool from gen.OptBool, returning false if not set.
func optBool(o gen.OptBool) bool {
	if !o.Set {
		return false
	}
	return o.Value
}
