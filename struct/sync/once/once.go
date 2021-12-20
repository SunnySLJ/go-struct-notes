package once

import "sync"

type singleton map[string]string

var (
	o        sync.Once
	instance singleton
)

func New() singleton {
	o.Do(func() {
		instance = make(singleton)
	})
	return instance
}
