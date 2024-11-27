package geoipgo

import (
	"net"
	"sort"
)

type IPv4TrieNode[T any] struct {
	IsEnd    bool
	Data     *T
	Children [2]*IPv4TrieNode[T]
	Ranges   []IPv4Range[T]
}

type IPv4Range[T any] struct {
	Start uint32
	End   uint32
	Data  *T
}

func (root *IPv4TrieNode[T]) Insert(cidr string, data T) error {
	tmpCIDR, err := ParseCIDR(cidr)
	if err != nil {
		return err
	}
	return root.InsertCIDR(tmpCIDR, data)
}

func (root *IPv4TrieNode[T]) InsertCIDR(cidr *Cidr, data T) error {
	start, end, mask := cidr.IPv4Uint()

	node := root
	for i := 0; i < mask; i++ {
		bit := (start >> (31 - i)) & 1
		if node.Children[bit] == nil {
			node.Children[bit] = &IPv4TrieNode[T]{}
		}
		node = node.Children[bit]
	}

	node.IsEnd = true
	node.Data = &data

	if mask < 32 {
		node.Ranges = append(node.Ranges, IPv4Range[T]{Start: start, End: end, Data: &data})
	}

	return nil
}

func (root *IPv4TrieNode[T]) Search(ip string) *T {
	tmpIP := net.ParseIP(ip)
	if tmpIP == nil {
		return nil
	}
	return root.SearchIP(tmpIP)
}

func (root *IPv4TrieNode[T]) SearchIP(ip net.IP) *T {
	ipUint := ipToUint32(ip)
	node := root
	var lastMatchedNode *IPv4TrieNode[T]

	for i := 0; i < 32; i++ {
		bit := (ipUint >> (31 - i)) & 1
		if node.Children[bit] == nil {
			break
		}
		node = node.Children[bit]

		if node.IsEnd {
			lastMatchedNode = node
		}
	}

	if lastMatchedNode != nil && lastMatchedNode.Data != nil {
		return lastMatchedNode.Data
	}

	if len(node.Ranges) > 0 {
		idx := sort.Search(len(node.Ranges), func(i int) bool {
			return ipUint <= node.Ranges[i].End
		})

		if idx < len(node.Ranges) && ipUint >= node.Ranges[idx].Start {
			return node.Ranges[idx].Data
		}
	}

	return nil
}
