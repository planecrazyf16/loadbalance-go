// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.


// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

package serverpool

import (
	"fmt"
	"iter"
)

// ServerPoolInterface defines the methods required for a server pool that manages nodes and their associated buckets.
// T is a comparable type that represents the type of the node.
type ServerPool[T comparable] interface {

	// AddNode adds a node to the server pool with the specified bucket.
	AddNode(node Node[T], bucket int) error

	// RemoveNode removes a node from the server pool.
	RemoveNode(node Node[T]) (int, error)

	// GetNode retrieves a node from the server pool for the specified bucket.
	GetNode(bucket int) (Node[T], bool)

	// Nodes returns an iterator sequence of all nodes and their associated buckets in the server pool.
	Nodes() iter.Seq2[Node[T], int]

	// Buckets returns an iterator sequence of all buckets and their associated nodes in the server pool.
	Buckets() iter.Seq2[int, Node[T]]
}

type serverPool[T comparable] struct {
	// nodeToBucket associates each Node  with an integer representing its bucket.
	// This mapping is used to distribute nodes across different buckets for load balancing purposes.
	nodeToBucket map[Node[T]]int

	// bucketToNode associates bucket indexes and the corresponding Node in the consistent hash ring.
	// Each bucket represents a position in the hash space and maps to a specific node responsible for that range.
	bucketToNode map[int]Node[T]
}

// Create a new server pool
func NewServerPool[T comparable]() *serverPool[T] {
	return &serverPool[T]{
		nodeToBucket: make(map[Node[T]]int),
		bucketToNode: make(map[int]Node[T]),
	}
}

// Add a new node with a given bucket index to the server pool
func (sp *serverPool[T]) AddNode(node Node[T], bucket int) error {
	if _, ok := sp.bucketToNode[bucket]; ok {
		return fmt.Errorf("bucket %d already exists", bucket)
	}
	if _, ok := sp.nodeToBucket[node]; ok {
		return fmt.Errorf("node already exists")
	}
	sp.nodeToBucket[node] = bucket
	sp.bucketToNode[bucket] = node

	return nil
}

// Remove a node from the server pool
func (sp *serverPool[T]) RemoveNode(node Node[T]) (int, error) {

	bucket, ok := sp.nodeToBucket[node]
	if !ok {
		return -1, fmt.Errorf("node not found")
	}
	delete(sp.nodeToBucket, node)
	delete(sp.bucketToNode, bucket)

	return bucket, nil
}

// Get the node responsible for the given bucket
func (sp *serverPool[T]) GetNode(bucket int) (Node[T], bool) {
	node, ok := sp.bucketToNode[bucket]
	return node, ok
}

// Iterate over all nodes in the server pool
func (sp *serverPool[T]) Nodes() iter.Seq2[Node[T], int] {
	return func(yield func(Node[T], int) bool) {
		for k, v := range sp.nodeToBucket {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Iterate over all buckets in the server pool
func (sp *serverPool[T]) Buckets() iter.Seq2[int, Node[T]] {
	return func(yield func(int, Node[T]) bool) {
		for k, v := range sp.bucketToNode {
			if !yield(k, v) {
				return
			}
		}
	}
}
