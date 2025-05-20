package validator

import (
	"sync"
	"sync/atomic"
)

const (
	typeDefault tagType = iota
	typeOmitEmpty
	typeIsDefault
	typeNoStructLevel
	typeStructOnly
	typeDive
	typeOr
	typeKeys
	typeEndKeys
	typeOmitNil
	typeOmitZero
)

type tagType uint8

type structCache struct {
	lock sync.Mutex
	m    atomic.Value
}
