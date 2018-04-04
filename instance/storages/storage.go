package storages

type SetHandler func(key string, val string, ver int64)
type RemoveHandler func(key string, ver int64)

type storage struct {
	data          ConcurrentMap
	setHandler    SetHandler
	removeHandler RemoveHandler
}

type record struct {
	value string
	ver   int64
}

type Storage interface {
	Set(string, string)
	Get(string) (string, bool)
	Remove(key string) bool
	List() map[string]string

	SetWithVersion(string, string, int64)
	RemoveWithVersion(string, int64)
	AddSetHandler(sh SetHandler)
	AddRemoveHandler(rh RemoveHandler)
}

func New() Storage {
	return &storage{
		data:          NewConcurrentMap(),
		setHandler:    nil,
		removeHandler: nil,
	}
}

func (s *storage) SetWithVersion(key string, val string, ver int64) {
	s.data.Upsert(key, func(exist bool, valueInMap interface{}) interface{} {
		if !exist {
			return record{val, ver}
		}

		rec := valueInMap.(record)
		if rec.ver < ver {
			rec.value = val
			rec.ver = ver
		}

		return rec
	})
}

func (s *storage) RemoveWithVersion(key string, ver int64) {
	s.data.PopIf(key, func(b bool, i interface{}) bool {
		return b && (i.(record).ver < ver)
	})
}

func (s *storage) AddSetHandler(sh SetHandler) {
	s.setHandler = sh
}

func (s *storage) AddRemoveHandler(rh RemoveHandler) {
	s.removeHandler = rh
}

func (s *storage) Set(key string, value string) {
	upserter := func(exist bool, valueInMap interface{}) interface{} {
		if exist {
			rec := valueInMap.(record)
			rec.ver++
			rec.value = value
			return rec
		}

		return record{
			value,
			1,
		}
	}

	s.data.Upsert(key, func(exist bool, valueInMap interface{}) interface{} {
		newValue := upserter(exist, valueInMap)
		if s.setHandler != nil {
			rec := newValue.(record)
			go s.setHandler(key, rec.value, rec.ver)
		}
		return newValue
	})
}

func (s *storage) Get(key string) (string, bool) {
	data, ok := s.data.Get(key, func(value interface{}) interface{} {
		rec := value.(record)
		rec.ver++
		return rec
	})

	if !ok {
		return ``, ok
	}

	return data.(record).value, ok
}

func (s *storage) Remove(key string) bool {
	value, ok := s.data.Pop(key)
	if ok && s.removeHandler != nil {
		go s.removeHandler(key, value.(record).ver+1)
	}
	return ok
}

func (s *storage) List() map[string]string {
	result := make(map[string]string)
	for key, value := range s.data.Items() {
		result[key] = value.(record).value
	}
	return result
}
