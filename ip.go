package geoipgo

import (
	"math/big"
	"net"
)

func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func ipToUint128(ip net.IP) *big.Int {
	return new(big.Int).SetBytes(ip)
}

func boolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}
