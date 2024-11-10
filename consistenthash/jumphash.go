// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

// Implementation of JunpHash consistent hashing algorithm.
package consistenthash

func jumpHash(key uint64, numBuckets int) int {
	var b int64 = -1
	var j int64

	for j < int64(numBuckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}

	return int(b)
}
