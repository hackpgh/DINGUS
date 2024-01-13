package utils

import (
	"log"
	"sync"
)

// Singleton is a utility structure used to manage singleton instances.
// A singleton instance is created only once and the same instance is returned
// on subsequent calls.
type Singleton struct {
	once sync.Once   // Ensures that the object is initialized only once.
	obj  interface{} // The singleton object.
}

func NewSingleton(obj interface{}) *Singleton {
	return &Singleton{obj: obj}
}

func (s *Singleton) Get(initialize func() interface{}) interface{} {
	s.once.Do(func() {
		s.obj = initialize()
		log.Println("Singleton object initialized")
	})
	return s.obj
}
