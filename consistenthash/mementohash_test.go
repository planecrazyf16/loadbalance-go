// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

package consistenthash

import (
	"hashing"
	"testing"
)

func TestReplace(t *testing.T) {
	tests := []struct {
		name     string
		removed  map[int]replace
		bucket   int
		expected int
	}{
		{
			name: "bucket not removed",
			removed: map[int]replace{
				1: {bucket: 1, replacement: 2, prevRemoved: -1},
			},
			bucket:   0,
			expected: -1,
		},
		{
			name: "bucket removed",
			removed: map[int]replace{
				1: {bucket: 1, replacement: 2, prevRemoved: -1},
			},
			bucket:   1,
			expected: 2,
		},
		{
			name: "multiple buckets removed",
			removed: map[int]replace{
				1: {bucket: 1, replacement: 2, prevRemoved: -1},
				3: {bucket: 3, replacement: 4, prevRemoved: 1},
			},
			bucket:   3,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mementohash{
				removed: tt.removed,
			}
			if got := m.replace(tt.bucket); got != tt.expected {
				t.Errorf("replace() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestRestore(t *testing.T) {
	tests := []struct {
		name     string
		removed  map[int]replace
		bucket   int
		expected int
	}{
		{
			name:     "empty removed map",
			removed:  map[int]replace{},
			bucket:   0,
			expected: 1,
		},
		{
			name: "bucket in removed map",
			removed: map[int]replace{
				1: {bucket: 1, replacement: 2, prevRemoved: -1},
			},
			bucket:   1,
			expected: -1,
		},
		{
			name: "bucket not in removed map",
			removed: map[int]replace{
				1: {bucket: 1, replacement: 2, prevRemoved: -1},
			},
			bucket:   2,
			expected: -1,
		},
		{
			name: "multiple buckets in removed map",
			removed: map[int]replace{
				1: {bucket: 1, replacement: 2, prevRemoved: -1},
				3: {bucket: 3, replacement: 4, prevRemoved: 1},
			},
			bucket:   3,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mementohash{
				removed: tt.removed,
			}
			if got := m.restore(tt.bucket); got != tt.expected {
				t.Errorf("restore() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestGetBucket(t *testing.T) {
	tests := []struct {
		name     string
		buckets  int
		removed  map[int]replace
		key      string
		expected int
	}{
		{
			name:     "no buckets removed",
			buckets:  5,
			removed:  map[int]replace{},
			key:      "testkey1",
			expected: jumpHash(hashing.NewHashFunction(hashing.DefaultHashAlgorithm).HashString("testkey1"), 5),
		},
		{
			name:    "bucket removed",
			buckets: 5,
			removed: map[int]replace{
				1: {bucket: 1, replacement: 4, prevRemoved: 5},
			},
			key:      "testkey2",
			expected: 3, // Assuming the hash function and seed result in bucket 3
		},
		{
			name:    "multiple buckets removed",
			buckets: 5,
			removed: map[int]replace{
				1: {bucket: 1, replacement: 4, prevRemoved: 5},
				3: {bucket: 3, replacement: 3, prevRemoved: 1},
			},
			key:      "testkey3",
			expected: 4, // Assuming the hash function and seed result in bucket 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mementohash{
				HashFn:  hashing.NewHashFunction(hashing.DefaultHashAlgorithm),
				buckets: tt.buckets,
				removed: tt.removed,
			}
			if got := m.GetBucket(tt.key); got != tt.expected {
				t.Errorf("GetBucket() = %v, want %v", got, tt.expected)
			}
		})
	}
}
func TestRemoveBucket(t *testing.T) {
	tests := []struct {
		name        string
		buckets     int
		removed     map[int]replace
		bucket      int
		expected    int
		expectedLR  int
		expectedBkt int
	}{
		{
			name:        "no buckets added, removing bucket",
			buckets:     0,
			removed:     map[int]replace{},
			bucket:      0,
			expected:    -1,
			expectedLR:  0,
			expectedBkt: 0,
		},
		{
			name:        "no buckets removed, removing last bucket",
			buckets:     5,
			removed:     map[int]replace{},
			bucket:      4,
			expected:    4,
			expectedLR:  4,
			expectedBkt: 4,
		},
		{
			name:        "no buckets removed, removing non-last bucket",
			buckets:     5,
			removed:     map[int]replace{},
			bucket:      2,
			expected:    2,
			expectedLR:  2,
			expectedBkt: 5,
		},
		{
			name:        "some buckets removed, removing non-last bucket",
			buckets:     5,
			removed:     map[int]replace{1: {bucket: 1, replacement: 4, prevRemoved: -1}},
			bucket:      3,
			expected:    3,
			expectedLR:  3,
			expectedBkt: 5,
		},
		{
			name:        "some buckets removed, removing last bucket",
			buckets:     5,
			removed:     map[int]replace{1: {bucket: 1, replacement: 4, prevRemoved: -1}},
			bucket:      4,
			expected:    4,
			expectedLR:  4,
			expectedBkt: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mementohash{
				buckets: tt.buckets,
				removed: tt.removed,
			}
			if got := m.RemoveBucket(tt.bucket); got != tt.expected {
				t.Errorf("RemoveBucket() = %v, want %v", got, tt.expected)
			}
			if m.lastRemoved != tt.expectedLR {
				t.Errorf("lastRemoved = %v, want %v", m.lastRemoved, tt.expectedLR)
			}
			if m.buckets != tt.expectedBkt {
				t.Errorf("buckets = %v, want %v", m.buckets, tt.expectedBkt)
			}
		})
	}
}
func TestAddBucket(t *testing.T) {
	tests := []struct {
		name        string
		buckets     int
		lastRemoved int
		removed     map[int]replace
		expected    int
	}{
		{
			name:        "one bucket removed",
			buckets:     5,
			lastRemoved: 1,
			removed: map[int]replace{
				1: {bucket: 1, replacement: 4, prevRemoved: 0},
			},
			expected: 1,
		},
		{
			name:        "multiple buckets removed",
			buckets:     5,
			lastRemoved: 3,
			removed: map[int]replace{
				1: {bucket: 1, replacement: 4, prevRemoved: 0},
				3: {bucket: 3, replacement: 4, prevRemoved: 1},
			},
			expected: 3,
		},
		{
			name:        "restored bucket larger than current number of buckets",
			buckets:     2,
			lastRemoved: 3,
			removed: map[int]replace{
				3: {bucket: 3, replacement: 4, prevRemoved: 0},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mementohash{
				buckets:     tt.buckets,
				lastRemoved: tt.lastRemoved,
				removed:     tt.removed,
			}
			got := m.AddBucket()
			if got != tt.expected {
				t.Errorf("AddBucket() = %v, want %v", got, tt.expected)
			}
		})
	}
}
