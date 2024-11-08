// Copyright (c) 2024 Rishabh Parekh
// MIT License

// Use of this source code is governed by an MIT license that can be
// found in the LICENSE file.




// Simple server node implementation
package main

import (
	"net/netip"
)

type ServerNode struct {
	ip netip.Addr
}

func NewServerNode(ip netip.Addr) ServerNode {
	return ServerNode{ip}
}

func NewServerNodeBytes(addr [4]byte) ServerNode {
	return NewServerNode(netip.AddrFrom4(addr))
}

func NewServerNodeString(addr string) (ServerNode, error) {
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return ServerNode{}, err
	}
	return NewServerNode(ip), nil
}

func (sn ServerNode) Name() netip.Addr {
	return sn.ip
}

// Print the server node
func (sn ServerNode) String() string {
	return sn.ip.String()
}
