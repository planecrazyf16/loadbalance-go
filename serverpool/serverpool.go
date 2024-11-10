// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

package serverpool

import (
	"fmt"
	"iter"
)

// ServerPoolInterface defines the methods required for a server pool that manages nodes and their associated buckets.
// T is a comparable type that represents the type of the node.
type ServerPool[T,O comparable] interface {

	// AddNode adds a node to the server pool with the specified bucket.
	AddNode(node Node[T, O], bucket int) error

	// RemoveNode removes a node from the server pool.
	RemoveNode(node Node[T, O]) (int, Node[T, O], error)

	// GetNode retrieves a node from the server pool for the specified bucket.
	GetNode(bucket int) (Node[T, O], bool)

	// Nodes returns an iterator sequence of all nodes and their associated buckets in the server pool.
	Nodes() iter.Seq2[Node[T, O], int]

	// Buckets returns an iterator sequence of all buckets and their associated nodes in the server pool.
	Buckets() iter.Seq2[int, Node[T, O]]
}

type serverPool[T,O comparable] struct {
	// nodeToBucket associates each Node  with an integer representing its bucket.
	// This mapping is used to distribute nodes across different buckets for load balancing purposes.
	nodeToBucket map[T]int

	// bucketToNode associates bucket indexes and the corresponding Node in the consistent hash ring.
	// Each bucket represents a position in the hash space and maps to a specific node responsible for that range.
	bucketToNode map[int]Node[T, O]
}

// Create a new server pool
func NewServerPool[T, O comparable]() *serverPool[T, O] {
	return &serverPool[T, O]{
		nodeToBucket: make(map[T]int),
		bucketToNode: make(map[int]Node[T, O]),
	}
}

// Add a new node with a given bucket index to the server pool
func (sp *serverPool[T, O]) AddNode(node Node[T, O], bucket int) error {
	if _, ok := sp.bucketToNode[bucket]; ok {
		return fmt.Errorf("bucket %d already exists", bucket)
	}
	if _, ok := sp.nodeToBucket[node.Name()]; ok {
		return fmt.Errorf("node already exists")
	}
	sp.nodeToBucket[node.Name()] = bucket
	sp.bucketToNode[bucket] = node

	return nil
}

// Remove a node from the server pool
func (sp *serverPool[T, O]) RemoveNode(node Node[T, O]) (int, Node[T, O], error) {
	bucket, ok := sp.nodeToBucket[node.Name()]
	if !ok {
		return -1, nil, fmt.Errorf("node not found")
	}
	delete(sp.nodeToBucket, node.Name())

	n, ok := sp.bucketToNode[bucket]
	if !ok {
		return -1, nil, fmt.Errorf("bucket not found")
	}
	delete(sp.bucketToNode, bucket)

	return bucket, n, nil
}

// Get the node responsible for the given bucket
func (sp *serverPool[T, O]) GetNode(bucket int) (Node[T, O], bool) {
	node, ok := sp.bucketToNode[bucket]
	return node, ok
}

// Iterate over all nodes in the server pool
func (sp *serverPool[T, O]) Nodes() iter.Seq2[Node[T, O], int] {
	return func(yield func(Node[T,O], int) bool) {
		for k, v := range sp.bucketToNode {
			if !yield(v, k) {
				return
			}
		}
	}
}

// Iterate over all buckets in the server pool
func (sp *serverPool[T, O]) Buckets() iter.Seq2[int, Node[T, O]] {
	return func(yield func(int, Node[T,O]) bool) {
		for k, v := range sp.bucketToNode {
			if !yield(k, v) {
				return
			}
		}
	}
}
