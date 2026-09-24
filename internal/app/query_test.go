package app

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeQuery(t *testing.T) {
	assert.Nil(t, normalizeQuery(""))
	assert.Nil(t, normalizeQuery("   "))

	q := normalizeQuery("  abc  ")
	require.NotNil(t, q)
	assert.Equal(t, "abc", *q)

	long := strings.Repeat("中", maxQueryRunes+50)
	q = normalizeQuery(long)
	require.NotNil(t, q)
	assert.Equal(t, maxQueryRunes, utf8.RuneCountInString(*q))
}

func TestNormalizeListQuery(t *testing.T) {
	q, page, pageSize := normalizeListQuery("  x ", 0, 0)
	require.NotNil(t, q)
	assert.Equal(t, "x", *q)
	assert.Equal(t, 1, page)
	assert.Equal(t, 50, pageSize)
}
