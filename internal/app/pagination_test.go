package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		page, pageSize         int
		wantPage, wantPageSize int
	}{
		{0, 0, 1, 50},
		{-5, -1, 1, 50},
		{2, 100, 2, 100},
		{1, 1000, 1, 200},
	}

	for _, tt := range tests {
		page, pageSize := normalizePage(tt.page, tt.pageSize)
		assert.Equal(t, tt.wantPage, page)
		assert.Equal(t, tt.wantPageSize, pageSize)
	}
}

func TestIntSetKeys(t *testing.T) {
	assert.Empty(t, intSetKeys(map[int]struct{}{}))

	got := intSetKeys(map[int]struct{}{1: {}, 2: {}, 3: {}})
	assert.Len(t, got, 3)
	assert.ElementsMatch(t, []int{1, 2, 3}, got)
}
