package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCorrectVersion(t *testing.T) {
	// given

	// when
	version1011, err1011 := New("10.1.1")
	version1011h2, err1011h2 := New("10.1.1-h2")

	// then
	assert.NoError(t, err1011)
	assert.Equal(t, 10, version1011.Major)
	assert.Equal(t, 1, version1011.Minor)
	assert.Equal(t, 1, version1011.Patch)
	assert.Equal(t, "", version1011.Hotfix)
	assert.NoError(t, err1011h2)
	assert.Equal(t, 10, version1011h2.Major)
	assert.Equal(t, 1, version1011h2.Minor)
	assert.Equal(t, 1, version1011h2.Patch)
	assert.Equal(t, "h2", version1011h2.Hotfix)
}

func TestIncorrectVersion(t *testing.T) {
	// given

	// when
	_, err101 := New("10.1")
	_, err1011h2h2 := New("10.1.1-h2-h2")

	// then
	assert.Error(t, err101)
	assert.Error(t, err1011h2h2)
}

func TestVersionComparison(t *testing.T) {
	// given

	// when
	v1, _ := New("10.1.1")
	v2, _ := New("10.2.1-h2")

	// then
	assert.True(t, v2.Gte(v1))
}
