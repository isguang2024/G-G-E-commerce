// Package pathmatch 提供 APP 入口解析使用的 host / path 模式匹配工具。
//
// 支持的 path 模式语法：
//   /admin/         纯前缀匹配
//   /admin/*        单段通配（不匹配 /）
//   /admin/**       多段通配
//   /shop/:id       命名参数（单段）
//   /shop/:id/**    命名参数 + 多段
package pathmatch

import (
	"regexp"
	"strings"
	"sync"
)

// MatchType 枚举
const (
	HostExact   = "host_exact"
	HostSuffix  = "host_suffix"
	PathPrefix  = "path_prefix"
	HostAndPath = "host_and_path"
)

// NormalizeHost 规范化 host：去除大小写/端口/前后空白。
func NormalizeHost(host string) string {
	h := strings.TrimSpace(strings.ToLower(host))
	if h == "" {
		return ""
	}
	if i := strings.Index(h, ":"); i > 0 {
		h = h[:i]
	}
	return strings.TrimSuffix(h, ".")
}

// NormalizePath 规范化 path：保证以 / 开头。空字符串返回 "/"。
func NormalizePath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return "/"
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// NormalizeHostPattern 规范化 host 模式（匹配时使用）。
// host_suffix 类型允许前缀 "*." 或 "."，统一为以 "." 开头的 suffix 字符串。
func NormalizeHostPattern(matchType, host string) string {
	h := NormalizeHost(host)
	if h == "" {
		return ""
	}
	if matchType == HostSuffix {
		h = strings.TrimPrefix(h, "*")
		if !strings.HasPrefix(h, ".") {
			h = "." + h
		}
	}
	return h
}

// NormalizePathPattern 规范化 path 模式：
// 去除多余空白，确保以 / 开头。
func NormalizePathPattern(pattern string) string {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// MatchHost 按匹配类型对比 host。
func MatchHost(matchType, pattern, host string) bool {
	h := NormalizeHost(host)
	if pattern == "" {
		return matchType == PathPrefix // path_prefix 时允许 host 为空
	}
	if h == "" {
		return false
	}
	switch matchType {
	case HostExact, HostAndPath:
		return strings.EqualFold(pattern, h)
	case HostSuffix:
		// pattern 已规范化为 ".aa.com"
		return h == strings.TrimPrefix(pattern, ".") || strings.HasSuffix(h, pattern)
	case PathPrefix:
		return true
	}
	return false
}

// CompilePathPattern 把路径模式编译为正则。
// 返回 nil 表示空模式（视为始终匹配）。
func CompilePathPattern(pattern string) (*regexp.Regexp, error) {
	p := NormalizePathPattern(pattern)
	if p == "" {
		return nil, nil
	}
	// 纯前缀（不含通配符）
	if !strings.ContainsAny(p, "*:") {
		quoted := regexp.QuoteMeta(p)
		return regexp.Compile("^" + quoted)
	}
	var sb strings.Builder
	sb.WriteString("^")
	segments := strings.Split(p, "/")
	for i, seg := range segments {
		if i > 0 {
			sb.WriteString("/")
		}
		switch {
		case seg == "":
			// 跳过（开头/末尾的空段）
		case seg == "**":
			sb.WriteString(".*")
		case seg == "*":
			sb.WriteString("[^/]+")
		case strings.HasPrefix(seg, ":"):
			sb.WriteString("[^/]+")
		default:
			sb.WriteString(regexp.QuoteMeta(seg))
		}
	}
	// 末尾以 / 结尾的纯前缀视作前缀匹配（不强制结尾）
	if !strings.HasSuffix(p, "**") && !strings.HasSuffix(p, "/") {
		sb.WriteString("(/|$)")
	}
	return regexp.Compile(sb.String())
}

// MatchPath 判断 path 是否匹配 pattern。空 pattern 视为匹配任意 path。
func MatchPath(pattern, path string) bool {
	if strings.TrimSpace(pattern) == "" {
		return true
	}
	re, err := compileCached(pattern)
	if err != nil || re == nil {
		return false
	}
	return re.MatchString(NormalizePath(path))
}

// PatternSpecificity 给一条规则打具体度分，分高的优先命中。
// 用于在同 match_type 内排序。
func PatternSpecificity(matchType, host, pathPattern string) int {
	score := 0
	switch matchType {
	case HostAndPath:
		score += 1000
	case HostExact:
		score += 800
	case PathPrefix:
		score += 600
	case HostSuffix:
		score += 400
	}
	score += len(host)
	// path 段越多越具体；通配符段降权。
	for _, seg := range strings.Split(NormalizePathPattern(pathPattern), "/") {
		if seg == "" {
			continue
		}
		switch {
		case seg == "**":
			score += 1
		case seg == "*", strings.HasPrefix(seg, ":"):
			score += 3
		default:
			score += 10
		}
	}
	return score
}

// IsHostInScope 检查 child 的 host 模式是否落在 parent 的 host 模式内。
// 用于 Level 2 规则不能超出 Level 1 规则范围的校验。
// parent 为空 host（path_prefix 类型）时，child 不受 host 约束。
func IsHostInScope(parentMatchType, parentHost, childMatchType, childHost string) bool {
	parent := NormalizeHostPattern(parentMatchType, parentHost)
	child := NormalizeHostPattern(childMatchType, childHost)
	if parent == "" {
		return true
	}
	if child == "" {
		return false
	}
	switch parentMatchType {
	case HostExact, HostAndPath:
		return strings.EqualFold(parent, child)
	case HostSuffix:
		// child 必须落在 parent 的后缀范围内
		stripped := strings.TrimPrefix(parent, ".")
		return strings.EqualFold(child, stripped) ||
			strings.HasSuffix(child, parent) ||
			strings.HasSuffix("."+child, parent)
	}
	return true
}

// IsPathInScope 检查 child path 模式是否落在 parent path 模式内。
// 简化策略：若 parent 为空（无 path 约束），child 自由；否则要求 child 以 parent 的字面前缀（去除尾部通配符后）开头。
func IsPathInScope(parentPattern, childPattern string) bool {
	parent := strings.TrimSpace(parentPattern)
	if parent == "" {
		return true
	}
	parentLiteral := literalPrefix(parent)
	childLiteral := literalPrefix(childPattern)
	if parentLiteral == "" {
		return true
	}
	return strings.HasPrefix(childLiteral, parentLiteral)
}

// literalPrefix 返回 pattern 中第一个通配符前的字面前缀。
func literalPrefix(pattern string) string {
	p := NormalizePathPattern(pattern)
	for i := 0; i < len(p); i++ {
		if p[i] == '*' || p[i] == ':' {
			return p[:i]
		}
	}
	return p
}

var (
	patternCache   = make(map[string]*regexp.Regexp)
	patternCacheMu sync.RWMutex
)

func compileCached(pattern string) (*regexp.Regexp, error) {
	patternCacheMu.RLock()
	if re, ok := patternCache[pattern]; ok {
		patternCacheMu.RUnlock()
		return re, nil
	}
	patternCacheMu.RUnlock()
	re, err := CompilePathPattern(pattern)
	if err != nil {
		return nil, err
	}
	patternCacheMu.Lock()
	patternCache[pattern] = re
	patternCacheMu.Unlock()
	return re, nil
}
