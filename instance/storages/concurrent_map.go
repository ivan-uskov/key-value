package storages

import "sync"

const ShardCount = 32

type ConcurrentMap []*ConcurrentMapShared
type ConcurrentMapShared struct {
	items map[string]interface{}
	sync.RWMutex
}

func NewConcurrentMap() ConcurrentMap {
	m := make(ConcurrentMap, ShardCount)
	for i := 0; i < ShardCount; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]interface{})}
	}
	return m
}

func (m ConcurrentMap) getShard(key string) *ConcurrentMapShared {
	return m[uint(fnv32(key))%uint(ShardCount)]
}

func (m ConcurrentMap) Get(key string, updater func(interface{}) interface{}) (interface{}, bool) {
	shard := m.getShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	if ok {
		shard.items[key] = updater(v)
	}
	shard.Unlock()

	return v, ok
}

type Upserter func(exist bool, valueInMap interface{}) interface{}
type Updater func(exist bool, valueInMap interface{})

func (m ConcurrentMap) Upsert(key string, cb Upserter) (res interface{}) {
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	v, ok := shard.items[key]
	res = cb(ok, v)
	shard.items[key] = res

	return res
}

func (m ConcurrentMap) Update(key string, cb Updater) {
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	v, ok := shard.items[key]
	cb(ok, v)
}

func (m ConcurrentMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

type Tuple struct {
	Key string
	Val interface{}
}

func fanIn(chans []chan Tuple, out chan Tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

func (m ConcurrentMap) IterBuffered() <-chan Tuple {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple, total)
	go fanIn(chans, ch)
	return ch
}

func snapshot(m ConcurrentMap) (chans []chan Tuple) {
	chans = make([]chan Tuple, ShardCount)
	wg := sync.WaitGroup{}
	wg.Add(ShardCount)
	for index, shard := range m {
		go func(index int, shard *ConcurrentMapShared) {
			shard.RLock()
			chans[index] = make(chan Tuple, len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- Tuple{key, val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

func (m ConcurrentMap) Pop(key string) (v interface{}, exists bool) {
	shard := m.getShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

func (m ConcurrentMap) PopIf(key string, pred func(bool, interface{}) bool) (v interface{}, exists bool) {
	shard := m.getShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	if pred(exists, v) {
		delete(shard.items, key)
	}
	shard.Unlock()
	return v, exists
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
