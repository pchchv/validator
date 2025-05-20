package validator

import (
	"sync"
	"sync/atomic"
)

type structCache struct {
	lock sync.Mutex
	m    atomic.Value // map[reflect.Type]*cStruct
}
