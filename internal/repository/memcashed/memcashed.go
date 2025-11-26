// Package memcashed Реализация репозитория с хранением в оперативной памяти
package memcashed

import (
	"context"
	"fmt"
	"sync"

	"github.com/s-turchinskiy/urlsavailability/models"
)

type MemCashed struct {
	urlsKits   models.Data
	currentNum uint64
	mutex      sync.RWMutex
}

func New() *MemCashed {

	return &MemCashed{
		urlsKits: make(map[uint64]models.URLsKit),
	}
}

func (s *MemCashed) Update(ctx context.Context, kit map[string]bool) (uint64, error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.currentNum++
	s.urlsKits[s.currentNum] = kit

	return s.currentNum, nil
}

func (s *MemCashed) GetAllData(ctx context.Context) (models.FileStore, error) {

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return models.FileStore{Data: s.urlsKits, Num: s.currentNum}, nil
}

func (s *MemCashed) LoadAllData(ctx context.Context, data models.FileStore) error {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, v := range data.Data {
		s.urlsKits[i] = v
	}

	s.currentNum = data.Num

	return nil
}

func (s *MemCashed) GetDataWithFilter(ctx context.Context, nums []uint64) (models.URLsKit, error) {

	kits := make(models.Data)
	var missingNums []uint64

	result := make(models.URLsKit)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, num := range nums {
		val, ok := s.urlsKits[num]
		if ok {
			kits[num] = val
		} else {
			missingNums = append(missingNums, num)
		}
	}

	if len(missingNums) > 0 {
		return nil, fmt.Errorf("nums %v missing", missingNums)
	}

	for _, kit := range kits {
		for idx, val := range kit {
			result[idx] = val
		}
	}

	return result, nil
}
