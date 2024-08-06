package naming

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCamelCase(t *testing.T) {
	assert.Equal(t, "CamelCase", CamelCase("", "camel_case", "", true))
	assert.Equal(t, "camelCase", CamelCase("", "camel_case", "", false))
	assert.Equal(t, "camelCase", CamelCase("", "camel_case", "", false))
	assert.Equal(t, "camelCase", CamelCase("", "camel_{template_example}case", "", false))
	assert.Equal(t, "camelCase", CamelCase("", "camel_case{template_example}", "", false))
}

func TestUnderscore(t *testing.T) {
	assert.Equal(t, "under_score", Underscore("", "under score", ""))
	assert.Equal(t, "under_score", Underscore("", "under_score", ""))
	assert.Equal(t, "under_score", Underscore("", "UnderScore", ""))
	assert.Equal(t, "under_score", Underscore("", "underScore", ""))
	assert.Equal(t, "under_score", Underscore("", "underScore{template_example}", ""))
	assert.Equal(t, "under_score", Underscore("", "under{template_example}Score", ""))
}

func TestAlphaNumeric(t *testing.T) {
	assert.Equal(t, "alphanumeric", AlphaNumeric("alpha_numeric"))
	assert.Equal(t, "AlphaNumeric", AlphaNumeric("Alpha_Numeric"))
}
