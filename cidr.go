package geoipgo

import (
	"math/big"
	"net"
)

type Cidr struct {
	IP    net.IP
	IPNet *net.IPNet
}

type CidrParser interface {
	ParseCIDR() (*Cidr, error)
}

type StringCidr string

func (c StringCidr) ParseCIDR() (*Cidr, error) {
	return ParseCIDR(string(c))
}

func (c *Cidr) ParseCIDR() (*Cidr, error) {
	return c, nil
}

func (c *Cidr) IPv4Uint() (uint32, uint32, int) {
	mask := c.IPNet.Mask
	prefix, _ := c.IPNet.Mask.Size()
	start := ipToUint32(c.IP)
	end := start | ^ipToUint32(net.IP(mask))
	return start, end, prefix
}

func (c *Cidr) IPv6Uint() (*big.Int, *big.Int, int) {
	mask := c.IPNet.Mask
	prefix, _ := c.IPNet.Mask.Size()
	start := ipToUint128(c.IP)
	ones, bits := mask.Size()
	end := new(big.Int).Set(start)
	maskBig := new(big.Int).Lsh(big.NewInt(1), uint(bits-ones))
	maskBig.Sub(maskBig, big.NewInt(1))
	end.Add(end, maskBig)
	return start, end, prefix
}

func ParseCIDR(cidr string) (*Cidr, error) {
	ip, net, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	return &Cidr{IP: ip, IPNet: net}, nil
}
