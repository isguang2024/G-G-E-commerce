package handlers

import (
	"encoding/json"

	"github.com/go-faster/jx"

	"github.com/gg-ecommerce/backend/api/gen"
)

// anyObject is map[string]jx.Raw — mirrors the former gen.AnyObject alias.
type anyObject = map[string]jx.Raw

// optAnyObject wraps an optional anyObject value.
type optAnyObject struct {
	Value anyObject
	Set   bool
}

// marshalAnyObject marshals any value to anyObject (map[string]jx.Raw).
// Returns empty anyObject on error.
func marshalAnyObject(v interface{}) anyObject {
	b, err := json.Marshal(v)
	if err != nil {
		return anyObject{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return anyObject{}
	}
	out := make(anyObject, len(m))
	for k, val := range m {
		b2, _ := json.Marshal(val)
		out[k] = jx.Raw(b2)
	}
	return out
}

// marshalList converts a slice to []anyObject using marshalAnyObject.
func marshalList[T any](items []T) []anyObject {
	out := make([]anyObject, 0, len(items))
	for i := range items {
		out = append(out, marshalAnyObject(items[i]))
	}
	return out
}

// unmarshalAnyObject converts anyObject to a map then unmarshals into target.
func unmarshalAnyObject(src anyObject, target interface{}) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func optAnyObjectToMap(src optAnyObject) map[string]interface{} {
	if !src.Set {
		return nil
	}
	target := map[string]interface{}{}
	if err := unmarshalAnyObject(src.Value, &target); err != nil {
		return nil
	}
	return target
}

func optSystemMetaToMap(src gen.OptSystemMeta) map[string]interface{} {
	if !src.Set {
		return nil
	}
	target := map[string]interface{}{}
	b, err := json.Marshal(src.Value)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(b, &target); err != nil {
		return nil
	}
	return target
}

func permissionBatchTemplatePayloadToMap(src gen.PermissionActionBatchTemplatePayload) map[string]interface{} {
	target := map[string]interface{}{}
	if len(src.Ids) > 0 {
		ids := make([]string, 0, len(src.Ids))
		for _, id := range src.Ids {
			ids = append(ids, id.String())
		}
		target["ids"] = ids
	}
	if src.Status.Set {
		target["status"] = src.Status.Value
	}
	if src.ModuleGroupID.Set {
		target["module_group_id"] = src.ModuleGroupID.Value.String()
	}
	if src.FeatureGroupID.Set {
		target["feature_group_id"] = src.FeatureGroupID.Value.String()
	}
	if src.TemplateName.Set {
		target["template_name"] = src.TemplateName.Value
	}
	if len(target) == 0 {
		return nil
	}
	return target
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

// optString extracts string from gen.OptString, returning "" if not set.
func optString(o gen.OptString) string {
	if !o.Set {
		return ""
	}
	return o.Value
}

func optStringValue(v string) gen.OptString {
	if v == "" {
		return gen.OptString{}
	}
	return gen.OptString{
		Value: v,
		Set:   true,
	}
}

// optInt extracts int from gen.OptInt, returning def if not set.
func optInt(o gen.OptInt, def int) int {
	if !o.Set {
		return def
	}
	return o.Value
}

func optIntValue(v int) gen.OptInt {
	return gen.OptInt{
		Value: v,
		Set:   true,
	}
}
