package ginUpgradingOverriding

import (
	"sync"

	"github.com/pchchv/validator"
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}
