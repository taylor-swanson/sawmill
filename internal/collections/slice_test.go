package collections

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSliceRemoveElement(t *testing.T) {
	tests := map[string]struct {
		In      []int
		InIndex int
		Want    []int
	}{
		"nil_0": {
			In:      nil,
			InIndex: 0,
			Want:    nil,
		},
		"nil_1": {
			In:      nil,
			InIndex: 0,
			Want:    nil,
		},
		"nil_-1": {
			In:      nil,
			InIndex: -1,
			Want:    nil,
		},
		"empty_0": {
			In:      []int{},
			InIndex: 0,
			Want:    []int{},
		},
		"empty_1": {
			In:      []int{},
			InIndex: 0,
			Want:    []int{},
		},
		"empty_-1": {
			In:      []int{},
			InIndex: -1,
			Want:    []int{},
		},
		"filled_0": {
			In:      []int{1, 3, 4, 7},
			InIndex: 0,
			Want:    []int{3, 4, 7},
		},
		"filled_1": {
			In:      []int{1, 3, 4, 7},
			InIndex: 1,
			Want:    []int{1, 4, 7},
		},
		"filled_last-elem": {
			In:      []int{1, 3, 4, 7},
			InIndex: 3,
			Want:    []int{1, 3, 4},
		},
		"filled_out-of-range": {
			In:      []int{1, 3, 4, 7},
			InIndex: 4,
			Want:    []int{1, 3, 4, 7},
		},
		"filled_-1": {
			In:      []int{1, 3, 4, 7},
			InIndex: -1,
			Want:    []int{1, 3, 4, 7},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := SliceRemoveElement(tc.In, tc.InIndex)

			require.Equal(t, tc.Want, got)
		})
	}
}
