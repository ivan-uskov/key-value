package main

import (
	"time"
	"encoding/gob"
	"os"
	"fmt"
	"sync"
)

type dataProvider func() map[string]string
type setter func(string, string)

type Persister struct {
	filePath string
	lister   dataProvider
	sync.Mutex
}

func NewPersister(filePath string, lister dataProvider) *Persister {
	return &Persister{filePath, lister, sync.Mutex{}}
}

func (p *Persister) Load(s setter) {
	f, err := os.Open(p.filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	var decodedMap map[string]string
	err = gob.NewDecoder(f).Decode(&decodedMap)
	if err != nil {
		fmt.Println(err)
		return
	}

	for k, v := range decodedMap {
		s(k, v)
	}
}

func (p *Persister) RunSaveLoop(delay time.Duration) {
	go func() {
		for {
			time.Sleep(delay)
			p.Persists()
		}
	}()
}

func (p *Persister) Persists() {
	p.Lock()
	defer p.Unlock()
	f, err := os.Create(p.filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	e := gob.NewEncoder(f)
	err = e.Encode(p.lister())
	if err != nil {
		fmt.Println(err)
		return
	}
}
