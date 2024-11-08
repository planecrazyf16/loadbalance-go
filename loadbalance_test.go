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

type mockServerPool[T comparable] struct {
	nodes map[int]serverpool.Node[T]
}

func (m *mockServerPool[T]) AddNode(node serverpool.Node[T], bucket int) error {
	if _, exists := m.nodes[bucket]; exists {
		return errors.New("bucket already exists")
	}
	m.nodes[bucket] = node
	return nil
}

func (m *mockServerPool[T]) RemoveNode(node serverpool.Node[T]) (int, error) {
	for bucket, n := range m.nodes {
		if n == node {
			delete(m.nodes, bucket)
			return bucket, nil
		}
	}
	return 0, errors.New("node not found")
}

func (m *mockServerPool[T]) GetNode(bucket int) (serverpool.Node[T], bool) {
	node, exists := m.nodes[bucket]
	return node, exists
}

func (m *mockServerPool[T]) Nodes() iter.Seq2[serverpool.Node[T], int] {
	// Implement as needed for tests
	return func(yield func(serverpool.Node[T], int) bool) {
		for bucket, node := range m.nodes {
			if !yield(node, bucket) {
				return
			}
		}
	}
}

func (m *mockServerPool[T]) Buckets() iter.Seq2[int, serverpool.Node[T]] {
	// Implement as needed for tests
	return func(yield func(int, serverpool.Node[T]) bool) {
		for bucket, node := range m.nodes {
			if !yield(bucket, node) {
				return
			}
		}
	}
}

type mockNode struct {
	ID string
}

func (n *mockNode) Name() string {
	return n.ID
}

type mockConsistentHasher struct {
	buckets int
}

func (m *mockConsistentHasher) AddBucket() int {
	m.buckets++
	return m.buckets
}

func (m *mockConsistentHasher) RemoveBucket(bucket int) int {
	m.buckets--
	return m.buckets
}

func (m *mockConsistentHasher) GetBucket(key string) int {
	if m.buckets == 0 {
		return 0
	}
	h := hashing.NewHashFunction(hashing.DefaultHashAlgorithm)
	return int(h.HashString(key)) % m.buckets
}

func (m *mockConsistentHasher) Size() int {
	return m.buckets
}

func TestAddNodes(t *testing.T) {
	//sp := serverpool.NewServerPool[string]()
	sp := &mockServerPool[string]{nodes: make(map[int]serverpool.Node[string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string]{sp: sp, ch: ch}

	nodes := []serverpool.Node[string]{
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
	sp := &mockServerPool[string]{nodes: make(map[int]serverpool.Node[string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string]{sp: sp, ch: ch}

	err := lb.AddNodes([]serverpool.Node[string]{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "no nodes to add" {
		t.Fatalf("expected 'no nodes to add' error, got %v", err)
	}
}
func TestRemoveNodes(t *testing.T) {
	sp := &mockServerPool[string]{nodes: make(map[int]serverpool.Node[string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string]{sp: sp, ch: ch}

	nodes := []serverpool.Node[string]{
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
	sp := &mockServerPool[string]{nodes: make(map[int]serverpool.Node[string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string]{sp: sp, ch: ch}

	err := lb.RemoveNodes([]serverpool.Node[string]{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "no nodes to remove" {
		t.Fatalf("expected 'no nodes to remove' error, got %v", err)
	}
}

func TestRemoveNodesMoreThanExist(t *testing.T) {
	sp := &mockServerPool[string]{nodes: make(map[int]serverpool.Node[string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string]{sp: sp, ch: ch}

	nodes := []serverpool.Node[string]{
		&mockNode{ID: "node1"},
	}

	// Add one node first
	err := lb.AddNodes(nodes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Try to remove more nodes than exist
	err = lb.RemoveNodes([]serverpool.Node[string]{
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
	sp := &mockServerPool[string]{nodes: make(map[int]serverpool.Node[string])}
	ch := &mockConsistentHasher{}
	lb := &loadBalancer[string]{sp: sp, ch: ch}

	nodes := []serverpool.Node[string]{
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

	expectedErr := fmt.Sprintf("node not found for bucket %d", 0)
	if err.Error() != expectedErr {
		t.Fatalf("expected '%s' error, got %v", expectedErr, err)
	}
}
