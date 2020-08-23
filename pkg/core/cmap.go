package core

import "sync"

type ConcurrentMap struct {
	mp      map[interface{}]interface{}
	mapLock sync.RWMutex
}

func (m *ConcurrentMap) ComputeIfAbsent(key interface{}, valGetter func(interface{}) (interface{}, error)) (interface{}, error) {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	if v, ok := m.mp[key]; ok {
		return v, nil
	} else {
		if val, err := valGetter(key); err == nil {
			m.mp[key] = val
			return val, nil
		} else {
			return nil, err
		}
	}
}

func (m *ConcurrentMap) Get(key interface{}) interface{} {
	m.mapLock.RLock()
	defer m.mapLock.RUnlock()
	return m.mp[key]
}

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		mp:      make(map[interface{}]interface{}),
		mapLock: sync.RWMutex{},
	}
}
