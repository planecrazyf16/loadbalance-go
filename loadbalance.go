// Copyright (c) 2024 Rishabh Parekh
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

type LoadBalancer[T,O comparable] interface {
	// Add a list of nodes to the hash ring
	AddNodes(nodes []serverpool.Node[T, O]) error

	// Remove a node from the hash ring
	RemoveNodes(nodes []serverpool.Node[T, O]) error

	// Get the node responsible for the given key
	GetNode(key string) (serverpool.Node[T,O], error)

	// Count of nodes in the cluster
	NodeCount() int

	// Iterate over all nodes in the load balancer
	Nodes() iter.Seq2[serverpool.Node[T,O], int]

	// Iterate over all buckets in the load balancer
	Buckets() iter.Seq2[int, serverpool.Node[T,O]]

	// Add objects to the load balancer
	AddObjects(objects []*serverpool.Object[T,O]) error

	// Remove objects from the load balancer
	RemoveObjects(objects []*serverpool.Object[T,O]) error

	// Assign an object to a node
	AssignObject(obj *serverpool.Object[T,O]) error

	// Unassign an object from a node
	UnassignObject(obj *serverpool.Object[T,O]) error

	// Iterate over all objects in the load balancer
	Objects() iter.Seq[*serverpool.Object[T,O]]
}

type loadBalancer[T,O comparable] struct {
	// serverPool is the pool of servers
	sp serverpool.ServerPool[T,O]

	// consistentHasher is the consistent hash algorithm implementation
	ch consistenthash.ConsistentHasher

	// Objects assigned to the nodes
	objects map[O]*serverpool.Object[T,O]
}

// Create a new load balancer
func NewLoadBalancer[T,O comparable]() LoadBalancer[T,O] {
	return &loadBalancer[T,O]{sp: serverpool.NewServerPool[T,O](),
		ch: consistenthash.NewConsistentHasher(),
	objects: make(map[O]*serverpool.Object[T,O])}
}

// Add a list of nodes to the load balancer
func (lb *loadBalancer[T,O]) AddNodes(nodes []serverpool.Node[T,O]) error {
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
func (lb *loadBalancer[T,O]) RemoveNodes(nodes []serverpool.Node[T,O]) error {
	if len(nodes) == 0 {
		return errors.New("no nodes to remove")
	}

	if len(nodes) > lb.ch.Size() {
		return fmt.Errorf("cannot remove more nodes than the size of the working set %d", lb.ch.Size())
	}

	for _, node := range nodes {
		bucket, removedNode, err := lb.sp.RemoveNode(node)
		if err != nil {
			return err
		}
		lb.ch.RemoveBucket(bucket)

		// Re-assign objects assigned to the deleted after removing the bucket 
		// so they are reassined to other nodes
		for obj := range removedNode.Objects() {
			lb.AssignObject(obj)
		}
	}
	return nil
}

// Get the node responsible for the given key
func (lb *loadBalancer[T,O]) GetNode(key string) (serverpool.Node[T,O], error) {
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

// AddObjects adds a list of objects to the load balancer's object pool.
func (lb *loadBalancer[T,O]) AddObjects(objects []*serverpool.Object[T,O]) error {
	if len(objects) == 0 {
		return errors.New("no objects to add")
	}

	for _, obj := range objects {
		lb.objects[obj.Id] = obj
	}
	return nil
}

// RemoveObjects removes the specified objects from the load balancer's pool.
func (lb *loadBalancer[T,O]) RemoveObjects(objects []*serverpool.Object[T,O]) error {
	if len(objects) == 0 {
		return errors.New("no objects to remove")
	}

	for _, obj := range objects {
		delete(lb.objects, obj.Id)
	}
	return nil
}

// AssignObject assigns an object to a node in the load balancer
func (lb *loadBalancer[T,O]) AssignObject(obj *serverpool.Object[T,O]) error {
	o, ok := lb.objects[obj.Id]
	if !ok {
		return fmt.Errorf("%v not found", obj)
	}

	node, err := lb.GetNode(obj.Name())
	if err != nil {
		return err
	}

	node.AssignObject(o)
	o.AssignToNode(&node)

	return nil
}

// UnassignObject unassigns an object from a node in the load balancer
func (lb *loadBalancer[T,O]) UnassignObject(obj *serverpool.Object[T,O]) error {
	o, ok := lb.objects[obj.Id]
	if !ok {
		return fmt.Errorf("%v not found", obj)
	}
	
	node, err := lb.GetNode(o.Name())
	if err != nil {
		return err
	}

	node.UnassignObject(o)
	o.UnassignFromNode()

	return nil
}


// Objects returns a sequence of pointers to serverpool.Object[O].
func (lb *loadBalancer[T,O]) Objects() iter.Seq[*serverpool.Object[T,O]] {
	return func(yield func(*serverpool.Object[T,O]) bool) {
		for _, obj := range lb.objects {
			if !yield(obj) {
				break
			}
		}
	}
}

// Count of nodes in the cluster
func (lb *loadBalancer[T,O]) NodeCount() int {
	return lb.ch.Size()
}

// Iterate over all nodes in the load balancer
func (lb *loadBalancer[T,O]) Nodes() iter.Seq2[serverpool.Node[T,O], int] {
	return lb.sp.Nodes()
}

// Iterate over all buckets in the load balancer
func (lb *loadBalancer[T,O]) Buckets() iter.Seq2[int, serverpool.Node[T,O]] {
	return lb.sp.Buckets()
}
