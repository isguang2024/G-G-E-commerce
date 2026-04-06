package permission

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
)

const (
	permissionUsagePatternUnused        = "unused"
	permissionUsagePatternAPIOnly       = "api_only"
	permissionUsagePatternPageOnly      = "page_only"
	permissionUsagePatternPackageOnly   = "package_only"
	permissionUsagePatternMultiConsumer = "multi_consumer"

	permissionDuplicatePatternNone               = "none"
	permissionDuplicatePatternCrossContextMirror = "cross_context_mirror"
	permissionDuplicatePatternSuspectedDuplicate = "suspected_duplicate"
)

type PermissionListItem struct {
	PermissionKey user.PermissionKey
	Audit         PermissionAuditProfile
}

type PermissionAuditProfile struct {
	APICount         int64
	PageCount        int64
	PackageCount     int64
	ConsumerTypes    []string
	UsagePattern     string
	UsageNote        string
	DuplicatePattern string
	DuplicateGroup   string
	DuplicateKeys    []string
	DuplicateNote    string
}

type PermissionAuditSummary struct {
	TotalCount              int64
	UnusedCount             int64
	APIOnlyCount            int64
	PageOnlyCount           int64
	PackageOnlyCount        int64
	MultiConsumerCount      int64
	CrossContextMirrorCount int64
	SuspectedDuplicateCount int64
}

type permissionUsageCounters struct {
	APICount     int64
	PageCount    int64
	PackageCount int64
}

type permissionDuplicateSource struct {
	ID            uuid.UUID
	PermissionKey string
	ContextType   string
}

type permissionDuplicateProfile struct {
	Pattern  string
	GroupKey string
	Keys     []string
	Note     string
}

func buildPermissionAuditProfile(
	item user.PermissionKey,
	counts permissionUsageCounters,
	duplicate permissionDuplicateProfile,
) PermissionAuditProfile {
	consumerTypes := buildPermissionConsumerTypes(counts)
	usagePattern, usageNote := buildPermissionUsagePattern(consumerTypes)
	currentKey := canonicalPermissionKey(item.PermissionKey)
	relatedKeys := filterPermissionRelatedKeys(duplicate.Keys, currentKey)

	profile := PermissionAuditProfile{
		APICount:         counts.APICount,
		PageCount:        counts.PageCount,
		PackageCount:     counts.PackageCount,
		ConsumerTypes:    consumerTypes,
		UsagePattern:     usagePattern,
		UsageNote:        usageNote,
		DuplicatePattern: permissionDuplicatePatternNone,
	}
	if len(relatedKeys) == 0 || duplicate.Pattern == "" || duplicate.Pattern == permissionDuplicatePatternNone {
		return profile
	}
	profile.DuplicatePattern = duplicate.Pattern
	profile.DuplicateGroup = duplicate.GroupKey
	profile.DuplicateKeys = relatedKeys
	profile.DuplicateNote = duplicate.Note
	return profile
}

func buildPermissionConsumerTypes(counts permissionUsageCounters) []string {
	consumers := make([]string, 0, 3)
	if counts.APICount > 0 {
		consumers = append(consumers, "api")
	}
	if counts.PageCount > 0 {
		consumers = append(consumers, "page")
	}
	if counts.PackageCount > 0 {
		consumers = append(consumers, "package")
	}
	return consumers
}

func buildPermissionUsagePattern(consumerTypes []string) (string, string) {
	switch len(consumerTypes) {
	case 0:
		return permissionUsagePatternUnused, "未被 API、页面或功能包消费"
	case 1:
		switch consumerTypes[0] {
		case "api":
			return permissionUsagePatternAPIOnly, "仅 API 消费"
		case "page":
			return permissionUsagePatternPageOnly, "仅页面消费"
		default:
			return permissionUsagePatternPackageOnly, "仅功能包消费"
		}
	default:
		return permissionUsagePatternMultiConsumer, joinPermissionAuditLabels(consumerTypes) + "复合消费"
	}
}

func buildPermissionDuplicateProfiles(items []permissionDuplicateSource) map[string]permissionDuplicateProfile {
	type familyGroup struct {
		Keys     []string
		Contexts map[string]struct{}
	}

	groups := make(map[string]*familyGroup, len(items))
	for _, item := range items {
		key := canonicalPermissionKey(item.PermissionKey)
		if key == "" {
			continue
		}
		familyKey := buildPermissionDuplicateFamilyKey(key)
		if familyKey == "" {
			continue
		}
		group := groups[familyKey]
		if group == nil {
			group = &familyGroup{
				Keys:     make([]string, 0, 2),
				Contexts: make(map[string]struct{}, 2),
			}
			groups[familyKey] = group
		}
		group.Keys = append(group.Keys, key)
		group.Contexts[normalizePermissionAuditContext(item.ContextType, key)] = struct{}{}
	}

	result := make(map[string]permissionDuplicateProfile, len(items))
	for familyKey, group := range groups {
		if len(group.Keys) <= 1 {
			continue
		}
		sort.Strings(group.Keys)
		related := strings.Join(group.Keys, " / ")
		pattern := permissionDuplicatePatternSuspectedDuplicate
		note := "同上下文疑似重复权限：" + related
		if len(group.Contexts) > 1 {
			pattern = permissionDuplicatePatternCrossContextMirror
			note = "跨上下文镜像权限：" + related
		}
		for _, key := range group.Keys {
			result[key] = permissionDuplicateProfile{
				Pattern:  pattern,
				GroupKey: familyKey,
				Keys:     append([]string(nil), group.Keys...),
				Note:     note,
			}
		}
	}

	return result
}

func buildPermissionDuplicateFamilyKey(permissionKey string) string {
	mapping := permissionkey.FromKey(permissionKey)
	resource := normalizePermissionAuditResource(mapping.ResourceCode)
	action := strings.TrimSpace(mapping.ActionCode)
	if resource == "" || action == "" {
		return ""
	}
	return resource + "." + action
}

func normalizePermissionAuditResource(value string) string {
	target := strings.Trim(strings.ReplaceAll(strings.TrimSpace(value), ".", "_"), "_")
	switch {
	case strings.HasPrefix(target, "system_"):
		target = strings.TrimPrefix(target, "system_")
	case strings.HasPrefix(target, "personal_"):
		target = strings.TrimPrefix(target, "personal_")
	}
	return target
}

func normalizePermissionAuditContext(contextType, permissionKey string) string {
	target := strings.TrimSpace(contextType)
	if target != "" {
		return target
	}
	mapping := permissionkey.FromKey(permissionKey)
	if strings.TrimSpace(mapping.ContextType) != "" {
		return strings.TrimSpace(mapping.ContextType)
	}
	return "collaboration"
}

func filterPermissionRelatedKeys(keys []string, currentKey string) []string {
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		target := canonicalPermissionKey(key)
		if target == "" || target == currentKey {
			continue
		}
		result = append(result, target)
	}
	return result
}

func joinPermissionAuditLabels(values []string) string {
	labels := make([]string, 0, len(values))
	for _, value := range values {
		switch value {
		case "api":
			labels = append(labels, "API")
		case "page":
			labels = append(labels, "页面")
		case "package":
			labels = append(labels, "功能包")
		default:
			labels = append(labels, value)
		}
	}
	return strings.Join(labels, "、")
}

func accumulatePermissionAuditSummary(summary *PermissionAuditSummary, profile PermissionAuditProfile) {
	if summary == nil {
		return
	}
	summary.TotalCount++
	switch profile.UsagePattern {
	case permissionUsagePatternUnused:
		summary.UnusedCount++
	case permissionUsagePatternAPIOnly:
		summary.APIOnlyCount++
	case permissionUsagePatternPageOnly:
		summary.PageOnlyCount++
	case permissionUsagePatternPackageOnly:
		summary.PackageOnlyCount++
	case permissionUsagePatternMultiConsumer:
		summary.MultiConsumerCount++
	}
	switch profile.DuplicatePattern {
	case permissionDuplicatePatternCrossContextMirror:
		summary.CrossContextMirrorCount++
	case permissionDuplicatePatternSuspectedDuplicate:
		summary.SuspectedDuplicateCount++
	}
}
