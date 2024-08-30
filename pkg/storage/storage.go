package storage

import (
	"sync"
)

type Storage struct {
	// 定义一个存储 pod 与 node 对应关系的 map
	// eg: podNamespace/podName: []string{nodeName-1,nodeName-2}
	storage map[string][]string
	m       sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		storage: make(map[string][]string),
	}
}

func (s *Storage) DeletePodOfStorage(key string) error {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.storage, key)

	return nil
}

func (s *Storage) AddOrUpdatePodOfStorage(key string, nodeNameSlice []string) error {
	s.m.Lock()
	defer s.m.Unlock()

	s.storage[key] = nodeNameSlice

	return nil
}

func (s *Storage) GetPodOfStorage(key string) ([]string, bool) {
	s.m.RLock()
	defer s.m.RUnlock()

	nodeNames, exist := s.storage[key]

	return nodeNames, exist
}

func (s *Storage) GetKeySliceOfStorage() []string {
	s.m.RLock()
	defer s.m.RUnlock()

	pods := make([]string, 0)
	for pod := range s.storage {
		pods = append(pods, pod)
	}

	return pods
}
