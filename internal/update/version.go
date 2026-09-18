package update

import (
	"fmt"
	"strings"
)

// NormalizeVersion strips a leading "v"/"V" and any build metadata (+...).
func NormalizeVersion(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	return s
}

// Compare returns -1 if a < b, 0 if equal, +1 if a > b (semver-ish major.minor.patch).
// Pre-release (a-b) is less than the same core version without a pre-release tag.
func Compare(a, b string) int {
	a = NormalizeVersion(a)
	b = NormalizeVersion(b)
	ap, aPre := splitPre(a)
	bp, bPre := splitPre(b)
	as := parseNums(ap)
	bs := parseNums(bp)
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		var av, bv int
		if i < len(as) {
			av = as[i]
		}
		if i < len(bs) {
			bv = bs[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	if aPre == "" && bPre != "" {
		return 1
	}
	if aPre != "" && bPre == "" {
		return -1
	}
	if aPre < bPre {
		return -1
	}
	if aPre > bPre {
		return 1
	}
	return 0
}

func splitPre(v string) (core, pre string) {
	if i := strings.IndexByte(v, '-'); i >= 0 {
		return v[:i], v[i+1:]
	}
	return v, ""
}

func parseNums(core string) []int {
	if core == "" {
		return nil
	}
	parts := strings.Split(core, ".")
	out := make([]int, len(parts))
	for i, p := range parts {
		out[i] = leadingInt(p)
	}
	return out
}

func leadingInt(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// NewerAvailable reports whether latest is strictly newer than current.
func NewerAvailable(current, latest string) bool {
	return Compare(current, latest) < 0
}

// FormatPair is a short human/agent-friendly status line.
func FormatPair(current, latest string) string {
	cur := NormalizeVersion(current)
	lat := NormalizeVersion(latest)
	switch Compare(cur, lat) {
	case 0:
		return fmt.Sprintf("up to date (%s)", cur)
	case -1:
		return fmt.Sprintf("update available: %s → %s", cur, lat)
	default:
		return fmt.Sprintf("current %s is newer than latest release %s", cur, lat)
	}
}
