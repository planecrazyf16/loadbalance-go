// Copyright (c) 2024 Rishabh Parekh
// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.

// Simple server node implementation
package main

import (
	"fmt"
	"iter"
	"net/netip"
	"serverpool"
)

type serverNode[O comparable] struct {
	ip netip.Addr

	// Objects assigned to the server node
	objects map[O]*serverpool.Object[netip.Addr,O]
}

func NewServerNode[O comparable](ip netip.Addr) serverNode[O] {
	return serverNode[O]{ip: ip, objects: make(map[O]*serverpool.Object[netip.Addr,O])}
}

func NewServerNodeBytes[O comparable](addr [4]byte) serverNode[O] {
	return NewServerNode[O](netip.AddrFrom4(addr))
}

func NewServerNodeString[O comparable](addr string) (serverNode[O], error) {
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return serverNode[O]{}, err
	}
	return NewServerNode[O](ip), nil
}

func (sn *serverNode[O]) Name() netip.Addr {
	return sn.ip
}


func (sn *serverNode[O]) AssignObject(obj *serverpool.Object[netip.Addr,O]) {
	sn.objects[obj.Id] = obj
}

func (sn *serverNode[O]) UnassignObject(obj *serverpool.Object[netip.Addr,O]) {
	delete(sn.objects, obj.Id)
}

func (sn *serverNode[O]) Objects() iter.Seq[*serverpool.Object[netip.Addr,O]] {
	return func(yield func(*serverpool.Object[netip.Addr,O]) bool) {
		for _, obj := range sn.objects {
			if !yield(obj) {
				break
			}
		}
	}
}

// Print the server node
func (sn *serverNode[O]) String() string {
	return fmt.Sprintf("ServerNode(%s)", sn.ip.String())
}
