// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package variables

import (
	"errors"
	"sort"
	"sync"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
)

type Store interface {
	Get(h Handle) (types.TypeVal, error)
	Exists(h Handle) bool
	Set(h Handle, value types.TypeVal)
	Delete(h Handle)
	Clear()

	ForEach(f func(h Handle, v types.TypeVal))

	Stats() (totalCount int64, totalSize int64, stats []*Stats)
}

type Stats struct {
	Type          string
	Count         int64
	EstimatedSize int64
}

type store struct {
	variables map[Handle]types.TypeVal
	m         *sync.RWMutex
}

func NewStore() Store {
	return &store{
		variables: make(map[Handle]types.TypeVal),
		m:         &sync.RWMutex{},
	}
}

func (s *store) Stats() (totalCount int64, totalSize int64, stats []*Stats) {
	s.m.RLock()
	defer s.m.RUnlock()

	m := make(map[string]*Stats)
	for _, v := range s.variables {
		s := v.EstimatedSize()
		totalSize += s
		n := v.Name()

		if nt, ok := m[n]; ok {
			nt.Count++
			nt.EstimatedSize = nt.EstimatedSize + s
		} else {
			m[n] = &Stats{
				Type:          n,
				Count:         1,
				EstimatedSize: s,
			}
		}
	}

	stats = make([]*Stats, 0, len(m))

	for _, v := range m {
		stats = append(stats, v)
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].EstimatedSize < stats[j].EstimatedSize
	})

	return int64(len(s.variables)), totalSize, stats
}

func (s *store) Get(h Handle) (types.TypeVal, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	if x, ok := s.variables[h]; ok {
		return x, nil
	}
	return nil, errors.New("variable does not exist: " + string(h))
}

func (s *store) Set(h Handle, value types.TypeVal) {
	s.m.Lock()
	defer s.m.Unlock()

	p, overwritting := s.variables[h]

	if m, ok := value.(types.Freer); ok {
		m.IncrementReferenceCount()
	}

	s.variables[h] = value

	if overwritting {
		if pF, ok := p.(types.Freer); ok {
			pF.Free()
		}
	}
}

func (s *store) deleteNoLock(h Handle) {
	if v, ok := s.variables[h]; ok {
		if m, ok := v.(types.Freer); ok {
			m.Free()
		}
	}

	delete(s.variables, h)
}

func (s *store) Delete(h Handle) {
	s.m.Lock()
	defer s.m.Unlock()

	s.deleteNoLock(h)
}

func (s *store) Clear() {
	s.m.Lock()
	defer s.m.Unlock()

	for h := range s.variables {
		s.deleteNoLock(h)
	}
}

func (s *store) Exists(h Handle) bool {
	s.m.RLock()
	defer s.m.RUnlock()

	_, exists := s.variables[h]
	return exists
}

func (s *store) ForEach(f func(h Handle, v types.TypeVal)) {
	s.m.RLock()
	defer s.m.RUnlock()

	for k, v := range s.variables {
		f(k, v)
	}
}
