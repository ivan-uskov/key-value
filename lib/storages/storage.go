package storages

import (
	cm "github.com/orcaman/concurrent-map"
)

type storage struct {
	data cm.ConcurrentMap
}

type Storage interface {
	Set(string, string)
	Get(string) (string, bool)
	Remove(key string) bool
	List() map[string]string
}

func New() Storage {
	return &storage{
		data: cm.New(),
	}
}

func (s *storage) Set(key string, value string) {
	s.data.Set(key, value)
	s.data.IterBuffered()
}

func (s *storage) Get(key string) (string, bool) {
	data, ok := s.data.Get(key)
	if ok {
		return data.(string), ok
	} else {
		return ``, ok
	}
}

func (s *storage) Remove(key string) bool {
	_, ok := s.data.Pop(key)
	return ok
}

func (s *storage) List() map[string]string {
	result := make(map[string]string)
	for key, value := range s.data.Items() {
		result[key] = value.(string)
	}
	return result
}