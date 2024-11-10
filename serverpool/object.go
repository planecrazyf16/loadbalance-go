// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

// Generic object
package serverpool

import "fmt"

type Object[T,O comparable] struct {
	// Unique identifier for the object
	Id O

	// Node the object is assigned to
	node *Node[T,O]
}

func (o *Object[T,O]) Name() string {
	return fmt.Sprintf("%v", o.Id)
}

func (o *Object[T,O]) AssignToNode(node *Node[T,O]) {
	o.node = node
}

func (o *Object[T,O]) UnassignFromNode() {
	o.node = nil
}

func (o *Object[T,O]) Node() *Node[T,O] {
	return o.node
}

func (o *Object[T,O]) String() string {
	return fmt.Sprintf("Object(%v)", o.Id)
}