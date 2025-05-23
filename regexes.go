package validator

import (
	"regexp"
	"sync"
)

func lazyRegexCompile(str string) func() (regex *regexp.Regexp) {
	var regex *regexp.Regexp
	var once sync.Once
	return func() *regexp.Regexp {
		once.Do(func() {
			regex = regexp.MustCompile(str)
		})
		return regex
	}
}
