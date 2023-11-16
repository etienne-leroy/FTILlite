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
	"fmt"

	"golang.org/x/exp/slices"
)

func SliceRemove[T any](xs []T, is []int64) ([]T, error) {
	result := xs
	for i := len(result) - 1; i >= 0; i-- {
		if slices.Contains(is, int64(i)) {
			result = append(result[:i], result[i+1:]...)
		}
	}
	return result, nil
}

func SliceBroadcast[T any](xs []T, length int64) ([]T, error) {
	if len(xs) != 1 {
		return nil, fmt.Errorf("cannot broadcast array of size %d", len(xs))
	}

	result := make([]T, length)

	for i := int64(0); i < length; i++ {
		result[i] = xs[0]
	}

	return result, nil
}
func SliceSet[T any](xs []T, is []int64, values []T) error {
	if len(is) != len(values) {
		return errors.New("wrong number of values for keys")
	}

	for i, k := range is {
		key := int(k)
		if key < 0 {
			key = key + len(xs)
		}
		if key >= len(xs) || key < 0 {
			return fmt.Errorf("out of range: %d", key)
		} else {
			xs[key] = values[i]
		}
	}
	return nil
}
func SliceMux[T any](xs []T, cond []int64, ifFalse []T) ([]T, error) {
	if len(xs) != len(ifFalse) || len(xs) != len(cond) {
		return nil, errors.New("array sizes must match")
	}
	results := make([]T, len(xs))
	for i, v := range cond {
		if v != 0 {
			results[i] = xs[i]
		} else {
			results[i] = ifFalse[i]
		}
	}
	return results, nil
}

func SliceContains[T any](xs []T, x T, p func(a, b T) bool) bool {
	for _, v := range xs {
		if p(v, x) {
			return true
		}
	}
	return false
}

func SliceMapUnary[T any, U any](as []T, f func(a T) U) ([]U, error) {
	us := make([]U, len(as))
	for i, a := range as {
		us[i] = f(a)
	}
	return us, nil
}
func SliceMapBinary[T any, U any, V any](as []T, bs []U, f func(a T, b U) V) ([]V, error) {
	if len(bs) < len(as) {
		return nil, errors.New("not enough values in other array")
	}
	us := make([]V, len(as))
	for i, a := range as {
		us[i] = f(a, bs[i])
	}
	return us, nil
}
func SliceMapBinary2[T any, U any, V any, W any](as []T, bs []U, f func(a T, b U) (V, W)) ([]V, []W, error) {
	if len(bs) < len(as) {
		return nil, nil, errors.New("not enough values in other array")
	}
	us := make([]V, len(as))
	vs := make([]W, len(as))
	for i, a := range as {
		u, v := f(a, bs[i])
		us[i] = u
		vs[i] = v
	}
	return us, vs, nil
}
func SliceMapBinaryE[T any, U any, V any](as []T, bs []U, f func(a T, b U) (V, error)) ([]V, error) {
	if len(bs) < len(as) {
		return nil, errors.New("not enough values in other array")
	}

	us := make([]V, len(as))

	var err error
	for i, a := range as {
		us[i], err = f(a, bs[i])
		if err != nil {
			return nil, err
		}
	}
	return us, nil
}
