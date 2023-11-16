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
	"fmt"

	"filippo.io/edwards25519"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
)

type Handle string

func GetAs[T types.TypeVal](s Store, h Handle) (result T, err error) {
	v, err := s.Get(h)
	if err != nil {
		return result, err
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("GetAs: %v", e)
		}
	}()

	result = v.(T)

	return result, err
}

func GetAsInteger(s Store, h Handle) (int64, error) {
	xs, err := GetAs[*types.FTIntegerArray](s, h)
	if err != nil {
		return 0, err
	}
	return xs.Single()
}
func GetAsFloat(s Store, h Handle) (float64, error) {
	xs, err := GetAs[*types.FTFloatArray](s, h)
	if err != nil {
		return 0, err
	}
	return xs.Single()
}
func GetAsEd25519Integer(s Store, h Handle) (*edwards25519.Scalar, error) {
	xs, err := GetAs[*types.FTEd25519IntArray](s, h)
	if err != nil {
		return nil, err
	}
	return xs.Single()
}
func GetAsBytes(s Store, h Handle) ([]byte, error) {
	xs, err := GetAs[*types.FTBytearrayArray](s, h)
	if err != nil {
		return nil, err
	}
	return xs.Single()
}
