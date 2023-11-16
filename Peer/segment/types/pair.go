// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package types

import (
	"errors"

	"golang.org/x/exp/constraints"
)

type Pair[T constraints.Ordered] struct {
	value T
	index int64
}

type PairSlice[T constraints.Ordered] []Pair[T]

func NewPairSlice[T constraints.Ordered](values []T, indices []int64) (PairSlice[T], error) {
	if len(values) != len(indices) {
		return nil, errors.New("array of values and indices must be same length")
	}
	p := make([]Pair[T], len(values))
	for i := 0; i < len(values); i++ {
		pair := Pair[T]{values[i], indices[i]}
		p[i] = pair
	}
	return p, nil
}

func (p PairSlice[T]) Len() int {
	return len(p)
}

func (p PairSlice[T]) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p PairSlice[T]) Less(i, j int) bool {
	if p[i].value == p[j].value {
		return p[i].index < p[j].index
	}
	return p[i].value < p[j].value
}

func (p PairSlice[T]) Index() []int64 {
	r := make([]int64, len(p))
	for i, vi := range p {
		r[i] = vi.index
	}
	return r
}
