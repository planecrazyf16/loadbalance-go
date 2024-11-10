// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"hashing"
	"iter"
	"testing"

	"serverpool"
)

type mockServerPool[T,O comparable] struct {
	nodes map[int]serverpool.Node[T,O]
}

func (m *mockServerPool[T,O]) AddNode(node serverpool.Node[T,O], bucket int) error {
	if _, exists := m.nodes[bucket]; exists {
		return errors.New("bucket already exists")
	}
	m.nodes[bucket] = node
	return nil
}

func (m *mockServerPool[T,O]) RemoveNode(node serverpool.Node[T,O]) (int, serverpool.Node[T,O], error) {
	for bucket, n := range m.nodes {
		if n == node {
			delete(m.nodes, bucket)
			return bucket, n, nil
		}
	}
	return 0, nil, errors.New("node not found")
}

func (m *mockServerPool[T,O]) GetNode(bucket int) (serverpool.Node[T,O], bool) {
	node, exists := m.nodes[bucket]
	return node, exists
}

func (m *mockServerPool[T,O]) Nodes() iter.Seq2[serverpool.Node[T,O], int] {
	// Implement as needed for tests
	return func(yield func(serverpool.Node[T,O], int) bool) {
		for bucket, node := range m.nodes {
			if !yield(node, bucket) {
				return
			}
		}
	}
}

func (m *mockServerPool[T,O]) Buckets() iter.Seq2[int, serverpool.Node[T,O]] {
	// Implement as needed for tests
	return func(yield func(int, serverpool.Node[T,O]) bool) {
		for bucket, node := range m.nodes {
			if !yield(bucket, node) {
				return
			}
		}
	}
}

type mockNode struct {
	ID string

	objects map[string]*serverpool.Object[string, string]
}

func (n *mockNode) Name() string {
	return n.ID
}

func (n *mockNode) AssignObject(obj *serverpool.Object[string, string]) {
	n.objects[obj.Id] = obj
}

func (n *mockNode) UnassignObject(obj *serverpool.Object[string, string]) {
	delete(n.objects, obj.Id)
}

func (n *mockNode) Objects() iter.Seq[*serverpool.Object[string, string]] {
	return func(yield func(*serverpool.Object[string, string]) bool) {
		for _, obj := range n.objects {
			if !yield(obj) {
				break
			}
		}
	}
}

type mockConsistentHasher struct {
	buckets int
}

func (m *mockConsistentHasher) AddBucket() int {
	bucket := m.buckets
	m.buckets++
	return bucket
}

func (m *mockConsistentHasher) RemoveBucket(bucket int) int {
	m.buckets--
	return m.buckets
}

func (m *mockConsistentHasher) GetBucket(key string) int {
	if m.buckets == 0 {
		return -1
	}
	h := hashing.NewHashFunction(hashing.DefaultHashAlgorithm)
	return int(h.HashString(key)) % m.buckets
}

func (m *mockConsistentHasher) Size() int {
	return m.buckets
}

func TestAddNodes(t *testing.T) {
	//sp := serverpool.NewServerPool[string,string]()
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch}

	nodes := []serverpool.Node[string, string]{
		&mockNode{ID: "node1"},
		&mockNode{ID: "node2"},
	}

	err := lb.AddNodes(nodes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(sp.nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(sp.nodes))
	}

	for _, node := range nodes {
		found := false
		for _, n := range sp.nodes {
			if n == node {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected node %v to be added", node)
		}
	}
}

func TestAddNodesEmpty(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string,string]{sp: sp, ch: ch}

	err := lb.AddNodes([]serverpool.Node[string,string]{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "no nodes to add" {
		t.Fatalf("expected 'no nodes to add' error, got %v", err)
	}
}
func TestRemoveNodes(t *testing.T) {
	sp := &mockServerPool[string,string]{nodes: make(map[int]serverpool.Node[string,string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string,string]{sp: sp, ch: ch}

	nodes := []serverpool.Node[string,string]{
		&mockNode{ID: "node1"},
		&mockNode{ID: "node2"},
	}

	// Add nodes first
	err := lb.AddNodes(nodes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Now remove nodes
	err = lb.RemoveNodes(nodes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(sp.nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(sp.nodes))
	}
}

func TestRemoveNodesEmpty(t *testing.T) {
	sp := &mockServerPool[string,string]{nodes: make(map[int]serverpool.Node[string,string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string,string]{sp: sp, ch: ch}

	err := lb.RemoveNodes([]serverpool.Node[string,string]{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "no nodes to remove" {
		t.Fatalf("expected 'no nodes to remove' error, got %v", err)
	}
}

func TestRemoveNodesMoreThanExist(t *testing.T) {
	sp := &mockServerPool[string,string]{nodes: make(map[int]serverpool.Node[string,string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string,string]{sp: sp, ch: ch}

	nodes := []serverpool.Node[string,string]{
		&mockNode{ID: "node1"},
	}

	// Add one node first
	err := lb.AddNodes(nodes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Try to remove more nodes than exist
	err = lb.RemoveNodes([]serverpool.Node[string,string]{
		&mockNode{ID: "node1"},
		&mockNode{ID: "node2"},
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedErr := fmt.Sprintf("cannot remove more nodes than the size of the working set %d", ch.Size())
	if err.Error() != expectedErr {
		t.Fatalf("expected '%s' error, got %v", expectedErr, err)
	}
}
func TestGetNode(t *testing.T) {
	sp := &mockServerPool[string,string]{nodes: make(map[int]serverpool.Node[string,string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string,string]{sp: sp, ch: ch}

	nodes := []serverpool.Node[string,string]{
		&mockNode{ID: "node1"},
		&mockNode{ID: "node2"},
	}

	// Add nodes first
	err := lb.AddNodes(nodes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test getting a node with a valid key
	key := "someKey"
	node, err := lb.GetNode(key)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if node == nil {
		t.Fatalf("expected a node, got nil")
	}

	// Test getting a node with an empty key
	_, err = lb.GetNode("")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "key cannot be empty" {
		t.Fatalf("expected 'key cannot be empty' error, got %v", err)
	}

	// Test getting a node with a key that does not map to any node
	ch.buckets = 0 // Reset buckets to simulate no nodes
	_, err = lb.GetNode("nonExistentKey")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedErr := fmt.Sprintf("node not found for bucket %d", -1)
	if err.Error() != expectedErr {
		t.Fatalf("expected '%s' error, got %v", expectedErr, err)
	}
}
func TestAddObjects(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch, objects: make(map[string]*serverpool.Object[string, string])}

	objects := []*serverpool.Object[string, string]{
		{Id: "obj1"},
		{Id: "obj2"},
	}

	err := lb.AddObjects(objects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(lb.objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(lb.objects))
	}

	for _, obj := range objects {
		if _, exists := lb.objects[obj.Id]; !exists {
			t.Fatalf("expected object %v to be added", obj)
		}
	}
}

func TestAddObjectsEmpty(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch, objects: make(map[string]*serverpool.Object[string, string])}

	err := lb.AddObjects([]*serverpool.Object[string, string]{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "no objects to add" {
		t.Fatalf("expected 'no objects to add' error, got %v", err)
	}
}
func TestRemoveObjects(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch, objects: make(map[string]*serverpool.Object[string, string])}

	objects := []*serverpool.Object[string, string]{
		{Id: "obj1"},
		{Id: "obj2"},
	}

	// Add objects first
	err := lb.AddObjects(objects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Now remove objects
	err = lb.RemoveObjects(objects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(lb.objects) != 0 {
		t.Fatalf("expected 0 objects, got %d", len(lb.objects))
	}
}

func TestRemoveObjectsEmpty(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch, objects: make(map[string]*serverpool.Object[string, string])}

	err := lb.RemoveObjects([]*serverpool.Object[string, string]{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "no objects to remove" {
		t.Fatalf("expected 'no objects to remove' error, got %v", err)
	}
}
func TestAssignObject(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch, objects: make(map[string]*serverpool.Object[string, string])}

	nodes := []serverpool.Node[string, string]{
		&mockNode{ID: "node1", objects: make(map[string]*serverpool.Object[string, string])},
		&mockNode{ID: "node2", objects: make(map[string]*serverpool.Object[string, string])},
	}

	// Add nodes first
	err := lb.AddNodes(nodes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	objects := []*serverpool.Object[string, string]{
		{Id: "obj1"},
		{Id: "obj2"},
	}

	// Add objects to the load balancer
	err = lb.AddObjects(objects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assign objects to nodes
	for _, obj := range objects {
		err = lb.AssignObject(obj)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify that the object is assigned to a node
		node, err := lb.GetNode(obj.Name())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := node.(*mockNode).objects[obj.Id]; !exists {
			t.Fatalf("expected object %v to be assigned to node %v", obj, node)
		}
	}
}

func TestAssignObjectNotFound(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch, objects: make(map[string]*serverpool.Object[string, string])}

	obj := &serverpool.Object[string, string]{Id: "obj1"}

	err := lb.AssignObject(obj)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedErr := fmt.Sprintf("%v not found", obj)
	if err.Error() != expectedErr {
		t.Fatalf("expected '%s' error, got %v", expectedErr, err)
	}
}
func TestUnassignObject(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch, objects: make(map[string]*serverpool.Object[string, string])}

	nodes := []serverpool.Node[string, string]{
		&mockNode{ID: "node1", objects: make(map[string]*serverpool.Object[string, string])},
		&mockNode{ID: "node2", objects: make(map[string]*serverpool.Object[string, string])},
	}

	// Add nodes first
	err := lb.AddNodes(nodes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	objects := []*serverpool.Object[string, string]{
		{Id: "obj1"},
		{Id: "obj2"},
	}

	// Add objects to the load balancer
	err = lb.AddObjects(objects)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Assign objects to nodes
	for _, obj := range objects {
		err = lb.AssignObject(obj)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	// Unassign objects from nodes
	for _, obj := range objects {
		err = lb.UnassignObject(obj)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify that the object is unassigned from the node
		node, err := lb.GetNode(obj.Name())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, exists := node.(*mockNode).objects[obj.Id]; exists {
			t.Fatalf("expected object %v to be unassigned from node %v", obj, node)
		}
	}
}

func TestUnassignObjectNotFound(t *testing.T) {
	sp := &mockServerPool[string, string]{nodes: make(map[int]serverpool.Node[string, string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string, string]{sp: sp, ch: ch, objects: make(map[string]*serverpool.Object[string, string])}

	obj := &serverpool.Object[string, string]{Id: "obj1"}

	err := lb.UnassignObject(obj)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	expectedErr := fmt.Sprintf("%v not found", obj)
	if err.Error() != expectedErr {
		t.Fatalf("expected '%s' error, got %v", expectedErr, err)
	}
}