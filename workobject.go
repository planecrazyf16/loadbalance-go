// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

// Simple work object implementation
package main

import (
	"fmt"
	"serverpool"
)

type workObject[T comparable] struct {
	serverpool.Object[T, int]
}

func NewWorkObject[T comparable](id int) *workObject[T] {
	return &workObject[T]{serverpool.Object[T, int]{Id: id}}
}

func (wo *workObject[T]) String() string {
	return fmt.Sprintf("WorkObject(%d)", wo.Id)
}

