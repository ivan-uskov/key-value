package main

import (
	"sync"
)

type registerStructure struct {
	data sync.Map
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
		data: sync.Map{},
	}
}

func (s *registerStructure) Get(key string) (Instance, bool) {
	data, ok := s.data.Load(key)
	if ok {
		return data.(Instance), ok
	} else {
		return nil, ok
	}
}

func (s *registerStructure) Exists(address string) bool {
	_, ok := s.data.Load(address)
	return ok
}

func (s *registerStructure) Add(address string, i Instance) {
	s.data.Store(address, i)
}

func (s *registerStructure) Remove(key string) bool {
	_, ok := s.data.Load(key)
	return ok
}

func (s *registerStructure) List() map[string]Instance {
	result := make(map[string]Instance)
	s.data.Range(func(k, v interface{}) bool {
		result[k.(string)] = v.(Instance)
		return true
	})
	return result
}