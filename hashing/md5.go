// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.




// Provides MD5 hashing functions.
package hashing

import (
	"crypto/md5"
	"encoding/binary"
)

type md5Hash struct{}

func md5Hasher() Hasher {
	return &md5Hash{}
}

func (m *md5Hash) hash(bytes []byte) uint64 {
	h := md5.New()
	h.Write(bytes)
	sum := h.Sum(nil)
	return binary.BigEndian.Uint64(sum[:8])
}
