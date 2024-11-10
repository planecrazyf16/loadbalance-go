// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

// consistenthash package provides a consistent hash algorithm implementation.
package consistenthash

import (
	"hashing"
)

type ConsistentHasher interface {
	// Add a bucket to the hash ring
	AddBucket() (int)

	// Remove a bucket from the hash ring
	RemoveBucket(bucket int) int

	// Get the bucket responsible for the given key
	GetBucket(key string) int

	// Get the size of the working set
	Size() int
}

func NewConsistentHasher() ConsistentHasher {
	return NewMementoHasher(hashing.DefaultHashAlgorithm)
}

func NewConsistentHasherWithAlgo(algo hashing.HashAlgorithm) ConsistentHasher {
	return NewMementoHasher(algo)
}
