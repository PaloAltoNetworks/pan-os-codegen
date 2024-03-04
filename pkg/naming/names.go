package naming

import (
	"fmt"
	"math/rand"
	"strings"
	"unicode"
)

type Namer struct {
	varnum int
	slugs  map[string]bool
}

// NewNamer returns a new Namer instance.
func NewNamer() *Namer {
	return &Namer{
		slugs: make(map[string]bool),
	}
}

// NextVarName returns a unique variable name (e.g. - "var2").
func (o *Namer) NextVarName() string {
	ans := fmt.Sprintf("var%d", o.varnum)
	o.varnum++

	return ans
}

// ResetVarNaming resets the variable naming, starting variable naming back
// at "var0".
func (o *Namer) ResetVarNaming() {
	o.varnum = 0
}

// NewSlug returns a new unique string.
//
// The random number generator is seeded with the given name in an attempt
// to minimize diffs between versions.
func (o *Namer) NewSlug(name string) string {
	const first = "abcdefghijklmnopqrstuvwxyz"
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var sum int64
	for r := range name {
		sum += int64(r)
	}
	rs := rand.New(rand.NewSource(sum))

	var ans string
	for {
		var b strings.Builder
		b.Grow(7)

		b.WriteByte(first[rs.Intn(len(first))])
		for i := 1; i < 7; i++ {
			b.WriteByte(letters[rs.Intn(len(letters))])
		}

		if !o.slugs[b.String()] {
			ans = b.String()
			break
		}
	}

	// Mark the slug as used now.
	o.slugs[ans] = true

	return ans
}

// CamelCase returns a camel case version of the given string.
func CamelCase(prefix, value, suffix string, capitalizeFirstRune bool) string {
	var b strings.Builder
	b.Grow(len(prefix) + len(value) + len(suffix))

	b.WriteString(prefix)

	first := true
	didFirstRune := false
	ignoreVar := false
	for _, r := range value {
		if ignoreVar {
			if r == '}' {
				ignoreVar = false
			}
			continue
		} else if r == '{' {
			ignoreVar = true
			first = true
			continue
		}

		if r == '_' || r == '-' || r == '|' || r == '/' || r == ':' || r == ' ' {
			first = true
		} else if first {
			if didFirstRune {
				b.WriteRune(unicode.ToTitle(r))
			} else {
				if capitalizeFirstRune {
					b.WriteRune(unicode.ToTitle(r))
				} else {
					b.WriteRune(unicode.ToLower(r))
				}
				didFirstRune = true
			}
			first = false
		} else {
			b.WriteRune(r)
		}
	}

	b.WriteString(suffix)

	return b.String()
}

// AlphaNumeric returns an alphanumeric version of the given string.
func AlphaNumeric(value string) string {
	var b strings.Builder
	b.Grow(len(value))

	for _, r := range value {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// ContainsReservedWord returns if any of the strings use builtin keywords.
func ContainsReservedWord(v []string) bool {
	checks := map[string]bool{
		"break":       true,
		"default":     true,
		"func":        true,
		"interface":   true,
		"select":      true,
		"case":        true,
		"defer":       true,
		"go":          true,
		"map":         true,
		"struct":      true,
		"chan":        true,
		"else":        true,
		"goto":        true,
		"package":     true,
		"switch":      true,
		"const":       true,
		"fallthrough": true,
		"if":          true,
		"range":       true,
		"type":        true,
		"continue":    true,
		"for":         true,
		"import":      true,
		"return":      true,
		"var":         true,
		"float":       true,
		"int":         true,
		"string":      true,
		"bool":        true,
		"make":        true,
	}

	for _, x := range v {
		if checks[x] {
			return true
		}
	}

	return false
}
