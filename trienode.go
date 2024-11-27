package geoipgo

import (
	"net"
)

type IPTrieNode[T any] struct {
	v4 *IPv4TrieNode[T]
	v6 *IPv6TrieNode[T]
}

func NewIPTrieNode[T any]() *IPTrieNode[T] {
	return &IPTrieNode[T]{
		v4: &IPv4TrieNode[T]{},
		v6: &IPv6TrieNode[T]{},
	}
}

func (root *IPTrieNode[T]) Insert(cidr CidrParser, data T) error {
	tmpCIDR, err := cidr.ParseCIDR()
	if err != nil {
		return err
	}
	if tmpCIDR == nil {
		return nil
	}
	if tmpCIDR.IP.To4() != nil {
		return root.v4.InsertCIDR(tmpCIDR, data)
	}
	return root.v6.InsertCIDR(tmpCIDR, data)
}

func (root *IPTrieNode[T]) Search(ip string) *T {
	tmpIP := net.ParseIP(ip)
	if tmpIP == nil {
		return nil
	}
	return root.SearchIP(tmpIP)
}

func (root *IPTrieNode[T]) SearchIP(ip net.IP) *T {
	if ip == nil {
		return nil
	}
	if ip.To4() != nil {
		return root.v4.SearchIP(ip)
	}
	return root.v6.SearchIP(ip)
}
