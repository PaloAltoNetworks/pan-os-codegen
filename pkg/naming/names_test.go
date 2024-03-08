package naming

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCamelCase(t *testing.T) {
	assert.Equal(t, "CamelCase", CamelCase("", "camel_case", "", true))
	assert.Equal(t, "camelCase", CamelCase("", "camel_case", "", false))
}

func TestAlphaNumeric(t *testing.T) {
	assert.Equal(t, "alphanumeric", AlphaNumeric("alpha_numeric"))
	assert.Equal(t, "AlphaNumeric", AlphaNumeric("Alpha_Numeric"))
}
