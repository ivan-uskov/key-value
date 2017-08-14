package main

import (
	cm "github.com/orcaman/concurrent-map"
	"key-value/lib/processes"
)

type registerStructure struct {
	data cm.ConcurrentMap
}

type Register interface {
	Add(address string, worker processes.Worker)
	Exists(address string) bool
	Remove(key string) bool
	Get(key string) (processes.Worker, bool)
	List() map[string]processes.Worker
}

func NewRegister() Register {
	return &registerStructure{
		data: cm.New(),
	}
}

func (s *registerStructure) Get(key string) (processes.Worker, bool) {
	data, ok := s.data.Get(key)
	if ok {
		return data.(processes.Worker), ok
	} else {
		return nil, ok
	}
}

func (s *registerStructure) Exists(address string) bool {
	return s.data.Has(address)
}

func (s *registerStructure) Add(address string, worker processes.Worker) {
	s.data.Set(address, worker)
}

func (s *registerStructure) Remove(key string) bool {
	_, ok := s.data.Pop(key)
	return ok
}

func (s *registerStructure) List() map[string]processes.Worker {
	result := make(map[string]processes.Worker)
	for key, value := range s.data.Items() {
		result[key] = value.(processes.Worker)
	}
	return result
}