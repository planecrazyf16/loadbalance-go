// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

package serverpool

import "iter"

type Node[T,O comparable] interface {
	// Get name of the node
	Name() T

	// Assign an object to the node
	AssignObject(obj *Object[T,O])

	// Unassign an object from the node
	UnassignObject(obj *Object[T,O])

	// Get all objects assigned to the node
	Objects() iter.Seq[*Object[T,O]]
}
