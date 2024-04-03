package naming

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCamelCase(t *testing.T) {
	assert.Equal(t, "CamelCase", CamelCase("", "camel_case", "", true))
	assert.Equal(t, "camelCase", CamelCase("", "camel_case", "", false))
}

func TestUnderscore(t *testing.T) {
	assert.Equal(t, "under_score", Underscore("", "under score", ""))
	assert.Equal(t, "under_score", Underscore("", "under_score", ""))
	assert.Equal(t, "under_score", Underscore("", "UnderScore", ""))
	assert.Equal(t, "under_score", Underscore("", "underScore", ""))
}

func TestAlphaNumeric(t *testing.T) {
	assert.Equal(t, "alphanumeric", AlphaNumeric("alpha_numeric"))
	assert.Equal(t, "AlphaNumeric", AlphaNumeric("Alpha_Numeric"))
}
