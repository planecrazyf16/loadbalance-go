// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.




// Provides SHA256 hashing functions.
package hashing

import (
	"crypto/sha256"
	"encoding/binary"
)

type sha256Hash struct{}

func sha256Hasher() Hasher {
	return &sha256Hash{}
}

func (s *sha256Hash) hash(bytes []byte) uint64 {
	h := sha256.New()
	h.Write(bytes)
	sum := h.Sum(nil)
	return binary.BigEndian.Uint64(sum[:8])
}
