package geoipgo

import (
	"log"
	"net"
	"sync"
)

type IPManager[C CidrParser, T any] interface {
	Update(CidrList[C, T]) error
	Query(ip string) *T
	QueryIP(ip net.IP) *T
}

type syncIPManager[C CidrParser, T any] struct {
	mu   sync.RWMutex
	root *IPTrieNode[T]
}

func NewSyncIPManager[C CidrParser, T any](dataSource IPDataSource[C, T]) IPManager[C, T] {
	manager := &syncIPManager[C, T]{
		root: &IPTrieNode[T]{},
	}

	ipData, err := dataSource.Fetch()
	if err != nil {
		log.Fatalf("Failed to fetch IP data source: %v", err)
	}

	if err := manager.Update(ipData); err != nil {
		log.Fatalf("Failed to initialize IP manager: %v", err)
	}

	return manager
}

func (m *syncIPManager[C, T]) Update(cidrList CidrList[C, T]) error {
	newRoot := NewIPTrieNode[T]()
	for _, cidr := range cidrList {
		if err := newRoot.Insert(cidr.Cidr, cidr.Info); err != nil {
			return err
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.root = newRoot
	return nil
}

func (m *syncIPManager[C, T]) Query(ip string) *T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.root.Search(ip)
}

func (m *syncIPManager[C, T]) QueryIP(ip net.IP) *T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.root.SearchIP(ip)
}
