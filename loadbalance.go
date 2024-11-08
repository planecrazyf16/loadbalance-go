// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.




// Load balancer

package main

import (
	"consistenthash"
	"errors"
	"fmt"
	"iter"
	"serverpool"
)

type LoadBalancer[T comparable] interface {
	// Add a list of nodes to the hash ring
	AddNodes(nodes []serverpool.Node[T]) error

	// Remove a node from the hash ring
	RemoveNodes(nodes []serverpool.Node[T]) error

	// Get the node responsible for the given key
	GetNode(key string) (serverpool.Node[T], error)

	// Count of nodes in the cluster
	NodeCount() int

	// Iterate over all nodes in the load balancer
	Nodes() iter.Seq2[serverpool.Node[T], int]

	// Iterate over all buckets in the load balancer
	Buckets() iter.Seq2[int, serverpool.Node[T]]
}

type loadBalancer[T comparable] struct {
	// serverPool is the pool of servers
	sp serverpool.ServerPool[T]

	// consistentHasher is the consistent hash algorithm implementation
	ch consistenthash.ConsistentHasher
}

// Create a new load balancer
func NewLoadBalancer[T comparable]() LoadBalancer[T] {
	return &loadBalancer[T]{sp: serverpool.NewServerPool[T](),
		ch: consistenthash.NewConsistentHasher()}
}

// Add a list of nodes to the load balancer
func (lb *loadBalancer[T]) AddNodes(nodes []serverpool.Node[T]) error {
	if len(nodes) == 0 {
		return errors.New("no nodes to add")
	}

	for _, node := range nodes {
		bucket := lb.ch.AddBucket()
		if err := lb.sp.AddNode(node, bucket); err != nil {
			return err
		}
	}
	return nil
}

// Remove a list of nodes from the load balancer
func (lb *loadBalancer[T]) RemoveNodes(nodes []serverpool.Node[T]) error {
	if len(nodes) == 0 {
		return errors.New("no nodes to remove")
	}

	if len(nodes) > lb.ch.Size() {
		return fmt.Errorf("cannot remove more nodes than the size of the working set %d", lb.ch.Size())
	}

	for _, node := range nodes {
		bucket, err := lb.sp.RemoveNode(node)
		if err != nil {
			return err
		}
		lb.ch.RemoveBucket(bucket)
	}
	return nil
}

// Get the node responsible for the given key
func (lb *loadBalancer[T]) GetNode(key string) (serverpool.Node[T], error) {
	if len(key) == 0 {
		return nil, errors.New("key cannot be empty")
	}
	bucket := lb.ch.GetBucket(key)
	node, ok := lb.sp.GetNode(bucket)
	if !ok {
		return nil, fmt.Errorf("node not found for bucket %d", bucket)
	}
	return node, nil
}

// Count of nodes in the cluster
func (lb *loadBalancer[T]) NodeCount() int {
	return lb.ch.Size()
}

// Iterate over all nodes in the load balancer
func (lb *loadBalancer[T]) Nodes() iter.Seq2[serverpool.Node[T], int] {
	return lb.sp.Nodes()
}

// Iterate over all buckets in the load balancer
func (lb *loadBalancer[T]) Buckets() iter.Seq2[int, serverpool.Node[T]] {
	return lb.sp.Buckets()
}
