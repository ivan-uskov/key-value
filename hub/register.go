package main

import (
	cm "github.com/orcaman/concurrent-map"
	"key-value/lib/processes"
)

type registerStructure struct {
	data cm.ConcurrentMap
}

type Register interface {
	Add(address string, i Instance)
	Exists(address string) bool
	Remove(key string) bool
	Get(key string) (Instance, bool)
	List() map[string]Instance
}

func NewRegister() Register {
	return &registerStructure{
		data: cm.New(),
	}
}

func (s *registerStructure) Get(key string) (Instance, bool) {
	data, ok := s.data.Get(key)
	if ok {
		return data.(Instance), ok
	} else {
		return nil, ok
	}
}

func (s *registerStructure) Exists(address string) bool {
	return s.data.Has(address)
}

func (s *registerStructure) Add(address string, i Instance) {
	s.data.Set(address, i)
}

func (s *registerStructure) Remove(key string) bool {
	_, ok := s.data.Pop(key)
	return ok
}

func (s *registerStructure) List() map[string]Instance {
	result := make(map[string]Instance)
	for key, value := range s.data.Items() {
		result[key] = value.(Instance)
	}
	return result
}