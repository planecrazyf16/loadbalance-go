// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

// Hashing functions used for consistent hashing algorithm

package hashing

import (
	"encoding/binary"
)

type HashAlgorithm int

const (
	CRC32 HashAlgorithm = iota
	MD5
	SHA256
)

var hashAlgorithmNames = map[HashAlgorithm]string{
	CRC32:  "crc32",
	MD5:    "md5",
	SHA256: "sha256",
}

const (
	// DefaultHashAlgorithm is the default hashing algorithm used by the consistent hash ring
	DefaultHashAlgorithm = CRC32
)

type Hasher interface {
	// Hash generates a hash value for a given byte slice and seed
	hash(bytes []byte) uint64

	// // HashString generates a hash value for a given string
	// HashString(input string) uint64

	// // HashStringWithSeed generates a hash value for a given string and seed
	// HashStringWithSeed(input string, seed int) uint64
}

type HashFn struct {
	hashAlgo HashAlgorithm
	Hasher
}

func (h HashFn) Hash(bytes []byte) uint64 {
	return h.hash(bytes)
}

// HashString generates a hash value for a given string using the configured algorithm
func (h HashFn) HashString(input string) uint64 {
	return h.hash([]byte(input))
}

// HashStringWithSeed generates a hash value for a given string and seed using the configured algorithm
func (h HashFn) HashStringWithSeed(input string, seed int) uint64 {
	strBytes := []byte(input)

	seedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(seedBytes, uint64(seed))

	combined := append(strBytes, seedBytes...)

	return h.hash(combined)
}

func (h HashFn) String() string {
	return hashAlgorithmNames[h.hashAlgo]
}

func NewHashFunction(algorithm HashAlgorithm) HashFn {
	var hasher Hasher
	switch algorithm {
	case CRC32:
		hasher = crc32Hasher()
	case MD5:
		hasher = md5Hasher()
	case SHA256:
		hasher = sha256Hasher()
	default:
		hasher = crc32Hasher()
	}
	return HashFn{hashAlgo: algorithm, Hasher: hasher}
}
