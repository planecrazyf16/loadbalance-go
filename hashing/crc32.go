// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.




// Provides CRC32 hashing functions.
package hashing

import (
	"hash/crc32"
)

type crc32Hash struct{}

func crc32Hasher() Hasher {
	return &crc32Hash{}
}

func (c *crc32Hash) hash(bytes []byte) uint64 {
	h := crc32.NewIEEE()
	h.Write(bytes)
	return uint64(h.Sum32())
}
