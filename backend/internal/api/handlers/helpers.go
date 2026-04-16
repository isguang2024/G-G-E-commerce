package handlers

import (
	"encoding/json"

	"github.com/maben/backend/api/gen"
)

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

func optSystemAppCapabilitiesToMap(src gen.OptSystemAppCapabilities) map[string]interface{} {
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

