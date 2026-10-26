package slug

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	allowedRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,48}[a-z0-9]$`)
	sanitizeRe = regexp.MustCompile(`[^a-z0-9_-]`)
)

// MaxLength slug 最大长度
const MaxLength = 50

// Generate 生成 slug：desired 合法则直接用，否则对 fallback 做 sanitize
func Generate(desired, fallback string) string {
	if s := strings.TrimSpace(strings.ToLower(desired)); allowedRe.MatchString(s) {
		return s
	}
	s := sanitize(strings.ToLower(fallback))
	if s == "" {
		s = "founder"
	}
	// 再夹逼一次长度
	if len(s) > MaxLength {
		s = s[:MaxLength]
	}
	// 若首位不是字母数字，补前导 'a'
	if len(s) > 0 && !isAlnum(s[0]) {
		s = "a" + s
	}
	if len(s) > MaxLength {
		s = s[:MaxLength]
	}
	return s
}

// IsValid 判断 slug 是否满足格式要求
func IsValid(s string) bool {
	return allowedRe.MatchString(s)
}

// IsUnique 返回唯一 slug（若冲突自动追加 -2 -3 ...）
func IsUnique(desired string, exists func(s string) (bool, error)) (string, error) {
	if exists == nil {
		return "", fmt.Errorf("exists func is nil")
	}
	candidate := desired
	for i := 2; i <= 9999; i++ {
		found, err := exists(candidate)
		if err != nil {
			return "", err
		}
		if !found {
			return candidate, nil
		}
		suffix := fmt.Sprintf("-%d", i)
		candidate = trimSuffixLen(desired, MaxLength-len(suffix)) + suffix
	}
	return "", fmt.Errorf("too many slug conflicts for %q", desired)
}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, "@", "-")
	s = sanitizeRe.ReplaceAllString(s, "")
	// 把连续的 -/_ 合并
	s = collapseRuns(s, '-')
	s = collapseRuns(s, '_')
	return s
}

func isAlnum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}

func trimSuffixLen(s string, n int) string {
	if n < 0 {
		n = 0
	}
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func collapseRuns(s string, r byte) string {
	var b strings.Builder
	last := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == r && last == r {
			continue
		}
		b.WriteByte(c)
		last = c
	}
	return strings.Trim(b.String(), "-_")
}
