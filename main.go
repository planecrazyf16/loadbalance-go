// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/netip"
	"os"
	"serverpool"
	"strconv"
	"time"
)

const (
	ADD int = iota + 1
	ADDNODE
	DELNODE
	MAP
	SHOWNODES
	SHOWBUCKETS
	ADDWORK
	REMWORK
	SHOWWORK
	EXIT
)

var r *rand.Rand
var addrs map[netip.Addr]struct{}

// Add the number of nodes specified to the load balancer
func addNodes(lb LoadBalancer[netip.Addr, int], numNodes int) {
	var bs [4]byte
	var nodes []serverpool.Node[netip.Addr, int]

	for i := 0; i < numNodes; i++ {
		// Generate a random IP address for each node in range [0, numNodes)
		addr := r.Intn(100000) + 1

		// Convert to byte array (little endian)
		binary.BigEndian.PutUint32(bs[:], uint32(addr))
		fmt.Println("Adding node with address:", bs)

		node := NewServerNodeBytes[int](bs)
		nodes = append(nodes, &node)

		addrs[node.Name()] = struct{}{}
	}
	lb.AddNodes(nodes)
}

// Add a node with given address
func addNode(lb LoadBalancer[netip.Addr, int], address string) {
	ip, err := netip.ParseAddr(address)
	if err != nil {
		fmt.Println("Invalid address")
		os.Exit(1)
	}

	if _, ok := addrs[ip]; ok {
		fmt.Println("Node already present")
		return
	}

	fmt.Println("Adding node with address:", ip)

	node := NewServerNode[int](ip)
	lb.AddNodes([]serverpool.Node[netip.Addr, int]{&node})

	addrs[ip] = struct{}{}
}

// Delete a node with given address
func delNode(lb LoadBalancer[netip.Addr,int], address string) {
	ip, err := netip.ParseAddr(address)
	if err != nil {
		fmt.Println("Invalid address")
		return
	}

	if _, ok := addrs[ip]; !ok {
		fmt.Println("Node not found")
		return
	}

	fmt.Println("Deleting node with address:", ip)

	node := NewServerNode[int](ip)
	lb.RemoveNodes([]serverpool.Node[netip.Addr, int]{&node})

	delete(addrs, ip)
}

// Add work to the load balancer
func addWork(lb LoadBalancer[netip.Addr, int], id string) {
	objid, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Invalid object ID")
		return
	}

	obj := NewWorkObject[netip.Addr](objid)

	if err := lb.AddObjects([]*serverpool.Object[netip.Addr, int]{&obj.Object}); err != nil {
		fmt.Println("Error adding work:", err)
		return
	}
	if err := lb.AssignObject(&obj.Object); err != nil {
		fmt.Println("Error assigning work:", err)
		return
	}	
}

// Remove work from the load balancer
func remWork(lb LoadBalancer[netip.Addr, int], id string) {
	objid, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println("Invalid object ID")
		return
	}

	if err := lb.UnassignObject(&serverpool.Object[netip.Addr, int]{Id: objid}); err != nil {
		fmt.Println("Error unassigning work:", err)
		return
	}

	if err := lb.RemoveObjects([]*serverpool.Object[netip.Addr, int]{{Id: objid}}); err != nil {
		fmt.Println("Error removing work:", err)
		return
	}
}

func readNewLine(reader *bufio.Reader) string {
	text, _ := reader.ReadString('\n') // Read until newline
	text = text[:len(text)-1]          // Remove newline character

	return text
}

func main() {
	lb := NewLoadBalancer[netip.Addr,int]()
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	addrs = make(map[netip.Addr]struct{})

	reader := bufio.NewReader(os.Stdin)

	op := 0
	for op < EXIT {
		fmt.Println()
		fmt.Println("1. Add nodes")
		fmt.Println("2. Add node")
		fmt.Println("3. Delete node")
		fmt.Println("4. Map Key")
		fmt.Println("5. Show Nodes")
		fmt.Println("6. Show Buckets")
		fmt.Println("7. Add Work")
		fmt.Println("8. Remove Work")
		fmt.Println("9. Show Work")
		fmt.Println("10. Exit")
		fmt.Print("Operation: ")
		text := readNewLine(reader)

		op, err := strconv.Atoi(text)
		if err != nil {
			fmt.Println("Invalid operation")
			os.Exit(1)
		}
		switch op {
		case ADD:
			fmt.Print("Enter number of nodes to add: ")
		text := readNewLine(reader)

			numNodes, err := strconv.Atoi(text)
			if err != nil {
				fmt.Println("Invalid number of nodes")
				os.Exit(1)
			}

			fmt.Println("Adding", numNodes, "nodes")
			addNodes(lb, numNodes)

		case ADDNODE:
			fmt.Print("Enter address of node to add: ")
			text := readNewLine(reader)

			fmt.Println("Adding node", text)
			addNode(lb, text)

		case DELNODE:
			fmt.Print("Enter address of node to delete: ")
			text := readNewLine(reader)

			fmt.Println("Deleting node", text)
			delNode(lb, text)

		case MAP:
			fmt.Print("Enter key to map: ")
			key := readNewLine(reader)

			node, err := lb.GetNode(key)
			if err != nil {
				fmt.Println("Error mapping key:", err)
			} else {
				fmt.Println("Key", key, "maps to node", node)
			}

		case SHOWNODES:
			fmt.Println("Nodes in the cluster:")
			for node, bucket := range lb.Nodes() {
				fmt.Printf("Node: %-15s Bucket: %d\n", node, bucket)
			}

		case SHOWBUCKETS:
			fmt.Println("Buckets in the cluster:")
			for bucket, node := range lb.Buckets() {
				fmt.Printf("Bucket: %d Node: %-15s\n", bucket, node)
			}
		case ADDWORK:
			fmt.Print("Enter id of work object to add: ")
			text := readNewLine(reader)

			fmt.Println("Adding work", text)
			addWork(lb, text)

		case REMWORK:
			fmt.Print("Enter id of work object to remove: ")
			text := readNewLine(reader)

			fmt.Println("Removing work", text)
			remWork(lb, text)

		case SHOWWORK:
			fmt.Println("Work assigned to nodes:")
			for obj := range lb.Objects() {
				fmt.Println(obj, "==>", *obj.Node())
			}

		case EXIT:
			os.Exit(0)
		}
		fmt.Print("Hit [Enter] to continue.")
		_ = readNewLine(reader)
	}
}
