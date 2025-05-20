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

type cTag struct {
	tag                  string
	aliasTag             string
	actualAliasTag       string
	param                string
	keys                 *cTag // only populated when using tag's 'keys' and 'endkeys' for map key validation
	next                 *cTag
	fn                   FuncCtx
	typeof               tagType
	hasTag               bool
	hasAlias             bool
	hasParam             bool // true if parameter used e. g. eq = where the equal sign has been set
	isBlockEnd           bool // indicates the current tag represents the last validation in the block
	runValidationWhenNil bool
}

type structCache struct {
	lock sync.Mutex
	m    atomic.Value
}
