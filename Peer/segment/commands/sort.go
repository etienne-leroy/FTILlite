// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package commands

import (
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const (
	CommandSorted      = "command_sorted"      // command_sorted <hTarget :: Handle→[](int64|float64)> <hValues :: Handle→[](int64|float64)>
	CommandIndexSorted = "command_indexsorted" // command_indexsorted <hTarget :: Handle→[](int64|float64)> <hValues :: Handle→[](int64|float64)>  //TODO: fix this comment
)

type SortableArrayTypeVal interface {
	types.ArrayTypeVal

	Sort() types.ArrayTypeVal
	IndexSort(indexes *types.FTIntegerArray) (*types.FTIntegerArray, error)
}

func Sorted(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hValues := variables.Handle(args[1])

	values, err := variables.GetAs[SortableArrayTypeVal](s.Variables(), hValues)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, values.Sort())
	return Ack, nil
}

func IndexSorted(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hValues := variables.Handle(args[1])

	values, err := variables.GetAs[SortableArrayTypeVal](s.Variables(), hValues)
	if err != nil {
		return "", err
	}

	var indexes *types.FTIntegerArray

	if len(args) == 3 {
		hIndex := variables.Handle(args[2])
		indexes, err = variables.GetAs[*types.FTIntegerArray](s.Variables(), hIndex)
		if err != nil {
			return "", err
		}
	} else {
		indexes = types.ArangeFTIntegerArray(values.Length())
	}

	result, err := values.IndexSort(indexes)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, result)

	return Ack, nil
}
