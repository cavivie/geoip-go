package geoipgo

type CidrItem[C CidrParser, T any] struct {
	Cidr C
	Info T
}

type CidrList[C CidrParser, T any] []CidrItem[C, T]

type IPDataSource[C CidrParser, T any] interface {
	Fetch() (CidrList[C, T], error)
}
