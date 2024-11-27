package geoipgo

import (
	"math/big"
	"net"
	"sort"
)

type IPv6TrieNode[T any] struct {
	IsEnd    bool
	Data     *T
	Children [2]*IPv6TrieNode[T]
	Ranges   []IPv6Range[T]
}

type IPv6Range[T any] struct {
	Start *big.Int
	End   *big.Int
	Data  T
}

func (root *IPv6TrieNode[T]) Insert(cidr string, data T) error {
	tmpCIDR, err := ParseCIDR(cidr)
	if err != nil {
		return err
	}
	return root.InsertCIDR(tmpCIDR, data)
}

func (root *IPv6TrieNode[T]) InsertCIDR(cidr *Cidr, data T) error {
	start, end, mask := cidr.IPv6Uint()

	node := root
	for i := 0; i < mask; i++ {
		bit := boolToInt((start.Rsh(start, uint(127-i)).Uint64() & 1) == 1)
		if node.Children[bit] == nil {
			node.Children[bit] = &IPv6TrieNode[T]{}
		}
		node = node.Children[bit]
	}

	node.IsEnd = true
	node.Data = &data

	if mask < 32 {
		node.Ranges = append(node.Ranges, IPv6Range[T]{Start: start, End: end, Data: data})
	}

	return nil
}

func (root *IPv6TrieNode[T]) Search(ip string) *T {
	tmpIP := net.ParseIP(ip)
	if tmpIP == nil {
		return nil
	}
	return root.SearchIP(tmpIP)
}

func (root *IPv6TrieNode[T]) SearchIP(ip net.IP) *T {
	ipBig := ipToUint128(ip)
	node := root
	var lastMatchedNode *IPv6TrieNode[T]

	for i := 0; i < 32; i++ {
		bit := boolToInt((ipBig.Rsh(ipBig, uint(127-i)).Uint64() & 1) == 1)
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
			return ipBig.Cmp(node.Ranges[i].End) != 1 // <= end
		})

		if idx < len(node.Ranges) && ipBig.Cmp(node.Ranges[idx].Start) != -1 { // >= Start
			return &node.Ranges[idx].Data
		}
	}

	return nil
}
