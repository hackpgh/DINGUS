package utils

import (
	"sync"
)

// Singleton is a utility structure used to manage singleton instances.
// A singleton instance is created only once and the same instance is returned
// on subsequent calls.
type Singleton struct {
	once sync.Once
	obj  interface{}
}

func NewSingleton(obj interface{}) *Singleton {
	return &Singleton{obj: obj}
}

func (s *Singleton) Get(initialize func() interface{}) interface{} {
	s.once.Do(func() {
		s.obj = initialize()
	})
	return s.obj
}
