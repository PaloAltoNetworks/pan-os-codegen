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

// CamelCase converts a string to CamelCase format, allowing for optional prefixes/suffixes,
// and control over the capitalization of the first character. It also skips over templated
// sections within the string, preserving their original form.
func CamelCase(prefix, value, suffix string, capitalizeFirstRune bool) string {
	var builder strings.Builder
	prepareStringBuilder(prefix, value, suffix, builder)

	isFirstCharacter := true
	hasFirstRuneBeenCapitalized := false
	isIgnoringTemplate := false

	for _, runeValue := range value {
		if isTemplateDelimiter(runeValue, &isIgnoringTemplate) {
			continue
		}

		writeRuneAndApplyChangesForCamelCase(runeValue, &builder, &isFirstCharacter, &hasFirstRuneBeenCapitalized, &capitalizeFirstRune)
	}

	builder.WriteString(suffix)
	return builder.String()
}

// writeRuneAndApplyChangesForCamelCase contains logic to check all conditions for create CamelCase names and writes rune value.
func writeRuneAndApplyChangesForCamelCase(runeValue int32, builder *strings.Builder, isFirstCharacter *bool, hasFirstRuneBeenCapitalized *bool, capitalizeFirstRune *bool) {
	if shouldResetFirstCharacterFlag(runeValue) {
		*isFirstCharacter = true
	} else if *isFirstCharacter {
		capitalizeAndWriteRune(builder, runeValue, *hasFirstRuneBeenCapitalized || *capitalizeFirstRune)
		*hasFirstRuneBeenCapitalized = true
		*isFirstCharacter = false
	} else {
		builder.WriteRune(runeValue)
	}
}

// Underscore converts a string to under_score format, allowing for optional prefixes/suffixes.
// It also skips over templated sections within the string, preserving their original form.
func Underscore(prefix, value, suffix string) string {
	var builder strings.Builder
	prepareStringBuilder(prefix, value, suffix, builder)

	isFirstCharacter := true
	isIgnoringTemplate := false

	for _, runeValue := range value {
		if isTemplateDelimiter(runeValue, &isIgnoringTemplate) {
			continue
		}

		writeRuneAndApplyChangesForUnderscore(runeValue, &builder, &isFirstCharacter)
	}

	builder.WriteString(suffix)
	return builder.String()
}

// writeRuneAndApplyChangesForUnderscore contains logic to check all conditions for create under_score names and writes rune value.
func writeRuneAndApplyChangesForUnderscore(runeValue int32, builder *strings.Builder, isFirstCharacter *bool) {
	if shouldResetFirstCharacterFlag(runeValue) {
		appendUnderscore(builder)
		*isFirstCharacter = true
	} else if *isFirstCharacter {
		capitalizeAndWriteRune(builder, runeValue, false)
		*isFirstCharacter = false
	} else if unicode.IsUpper(runeValue) {
		appendUnderscore(builder)
		capitalizeAndWriteRune(builder, runeValue, false)
	} else {
		capitalizeAndWriteRune(builder, runeValue, false)
	}
}

// appendUnderscore add _ to builder, from which final result is built.
func appendUnderscore(builder *strings.Builder) {
	builder.WriteRune('_')
}

// isTemplateDelimiter return true, if { or } are runeValue or isIgnoringTemplate is set to true.
func isTemplateDelimiter(runeValue int32, isIgnoringTemplate *bool) bool {
	switch runeValue {
	case '{':
		*isIgnoringTemplate = true
		return true
	case '}':
		*isIgnoringTemplate = false
		return true // if we finish template, then we always ignore that character, even if set isIgnoringTemplate to false
	}

	return *isIgnoringTemplate
}

func prepareStringBuilder(prefix string, value string, suffix string, builder strings.Builder) {
	builder.Grow(len(prefix) + len(value) + len(suffix))
	builder.WriteString(prefix)
}

// shouldResetFirstCharacterFlag checks if the given rune is a separator that should trigger
// the next character to be capitalized in CamelCase conversion.
func shouldResetFirstCharacterFlag(r rune) bool {
	return r == '_' || r == '-' || r == '|' || r == '/' || r == ':' || r == ' '
}

// capitalizeAndWriteRune writes the rune to the builder, capitalizing it if needed.
func capitalizeAndWriteRune(builder *strings.Builder, r rune, capitalize bool) {
	if capitalize {
		builder.WriteRune(unicode.ToTitle(r))
	} else {
		builder.WriteRune(unicode.ToLower(r))
	}
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
