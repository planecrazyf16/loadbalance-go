// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.




// Implementation of the mementohash consistent hashing algorithm.
package consistenthash

import (
	"fmt"
	"hashing"
)

type replace struct {
	// Removed bucket
	bucket int

	// Bucket that replaces the removed bucket
	// This is also the size of working set after removal of the current bucket
	replacement int

	// The buket removed before the current bucket
	prevRemoved int
}

func (r *replace) String() string {
	return fmt.Sprintf("%d -> (%d, %d)", r.bucket, r.replacement, r.prevRemoved)
}

// mementohash is an implementation of the ConsistentHasher interface
type mementohash struct {
	hashing.HashFn

	// The number of buckets in the hash ring
	buckets int

	// Last removed bucket
	lastRemoved int

	// Information about the removed buckets
	removed map[int]replace
}

// Function to add a removed buck to the replace table
// Store the previous removed bucket to create a chain of removed buckets
func (m *mementohash) remove(bucket, replacement, prevRemoved int) int {
	m.removed[bucket] = replace{bucket, replacement, prevRemoved}
	return bucket
}

// Returns replace bucket for the given bucket else -1
// The return value is also the size of the working set after removal of the current bucket
func (m *mementohash) replace(bucket int) int {
	if r, ok := m.removed[bucket]; ok {
		return r.replacement
	}
	return -1
}

// Restore the removed bucket and return the previous removed bucket
// If table is empty, return the next bucket
func (m *mementohash) restore(bucket int) int {
	if len(m.removed) == 0 {
		return bucket + 1
	}
	if r, ok := m.removed[bucket]; ok {
		delete(m.removed, bucket)
		return r.prevRemoved
	}
	return -1
}

// Returns the getBucket for the given key
func (m *mementohash) GetBucket(key string) int {
	// Use Jump Hash to get buck in range of [0, m.buckets)
	bucket := jumpHash(m.HashString(key), m.buckets)

	replace := m.replace(bucket)
	// Check if the bucket has been removed and needs replacement
	for replace >= 0 {
		// Get new bucket in remaining working set
		// The replacement bucket is the size of the working set after removal
		// Find new bucket in [0, replace - 1)
		bucket = int(m.HashStringWithSeed(key, bucket)) % replace

		// If bucket is removed, follow replacement chain till we find a valid bucket
		// in [0, replace -1)
		r := m.replace(bucket)
		for r >= replace {
			bucket = r
			r = m.replace(bucket)
		}
		replace = r
	}
	return bucket
}

// Add a new bucket to the hash ring
func (m *mementohash) AddBucket() int {
	// New bucket is the last removed bucket
	bucket := m.lastRemoved

	// Restore the last removed bucket and update the last removed bucket
	m.lastRemoved = m.restore(bucket)

	// If the restored bucket is larger than the current number of buckets,
	// add the bucket to the end of the ring
	if m.buckets <= bucket {
		m.buckets = bucket + 1
	}

	return bucket
}

// Remove a bucket from the hash ring
func (m *mementohash) RemoveBucket(bucket int) int {
	// If the bucket is not in the hash ring, return
	if bucket >= m.buckets {
		return -1
	}

	// If no buckets have been removed and the bucket to remove is last,
	// just update the number of buckets
	if len(m.removed) == 0 && bucket == m.buckets-1 {
		m.lastRemoved = bucket
		m.buckets = bucket
		return bucket
	}
	// Remove the bucket and add it to the replace table
	m.lastRemoved = m.remove(bucket, m.Size()-1, m.lastRemoved)

	return bucket
}

// Get size of the working set
func (m *mementohash) Size() int {
	return m.buckets - len(m.removed)
}

// NewMementoHasher creates a new instance of the mementohash consistent hashing algorithm
func NewMementoHasher(hashAlgo hashing.HashAlgorithm) ConsistentHasher {
	return &mementohash{removed: make(map[int]replace),
		HashFn: hashing.NewHashFunction(hashAlgo)}
}

func (m *mementohash) String() string {
	return fmt.Sprintf("MementoHasher{buckets: %d, lastRemoved: %d, removed: %v}", m.buckets, m.lastRemoved, m.removed)
}
