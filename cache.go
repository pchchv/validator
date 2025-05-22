package validator

import (
	"reflect"
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

const (
	invalidValidation   = "Invalid validation tag on field '%s'"
	undefinedValidation = "Undefined validation function '%s' on field '%s'"
	keysTagNotDefined   = "'" + endKeysTag + "' tag encountered without a corresponding '" + keysTag + "' tag"
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

type cField struct {
	idx        int
	name       string
	altName    string
	namesEqual bool
	cTags      *cTag
}

type cStruct struct {
	name   string
	fields []*cField
	fn     StructLevelFuncCtx
}

type structCache struct {
	lock sync.Mutex
	m    atomic.Value
}

func (sc *structCache) Get(key reflect.Type) (c *cStruct, found bool) {
	c, found = sc.m.Load().(map[reflect.Type]*cStruct)[key]
	return
}

func (sc *structCache) Set(key reflect.Type, value *cStruct) {
	m := sc.m.Load().(map[reflect.Type]*cStruct)
	nm := make(map[reflect.Type]*cStruct, len(m)+1)
	for k, v := range m {
		nm[k] = v
	}

	nm[key] = value
	sc.m.Store(nm)
}

type tagCache struct {
	lock sync.Mutex
	m    atomic.Value
}

func (tc *tagCache) Get(key string) (c *cTag, found bool) {
	c, found = tc.m.Load().(map[string]*cTag)[key]
	return
}

func (tc *tagCache) Set(key string, value *cTag) {
	m := tc.m.Load().(map[string]*cTag)
	nm := make(map[string]*cTag, len(m)+1)
	for k, v := range m {
		nm[k] = v
	}

	nm[key] = value
	tc.m.Store(nm)
}
