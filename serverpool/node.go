// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.




// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

package serverpool

type Node[T comparable] interface {
	// Get name of the node
	Name() T
}
