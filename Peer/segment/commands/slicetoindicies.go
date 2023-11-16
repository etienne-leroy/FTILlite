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
	"fmt"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandSliceToIndices = "command_slice_to_indices" // command_slicetoindices

func SliceToIndices(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	// Get the array of interest
	valTarget, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	// Get start indice
	arrLength := valTarget.Length()

	start, err := variables.GetAsInteger(s.Variables(), variables.Handle(args[2]))
	if err != nil {
		return "", err
	}

	// Get stop indice if provided, otherwise assume it is the length of the array
	stop := arrLength
	if len(args) >= 4 {
		stop, err = variables.GetAsInteger(s.Variables(), variables.Handle(args[3]))
		if err != nil {
			return "", err
		}
	}

	// Handle negative indices i.e. -1
	if start < 0 {
		start = arrLength + start
	}

	if stop < 0 {
		stop = arrLength + stop
	}

	// Return empty int array if start and stop equal
	if start == stop {
		s.Variables().Set(hResult, types.NewFTIntegerArray())
		return fmt.Sprintf("array i %s", hResult), nil
	}

	// Check start and stop don't exceed bounds of array
	if start < 0 || stop < 0 || start >= stop || start > arrLength || stop > arrLength {
		return "", fmt.Errorf("invalid slice values, start: %d stop: %d", start, stop)
	}

	keyLength := stop - start
	indices := make([]int64, keyLength)
	for i := int64(0); i < keyLength; i++ {
		indices[i] = int64(i + start)
	}
	s.Variables().Set(hResult, types.NewFTIntegerArray(indices...))

	return fmt.Sprintf("array i %s", hResult), nil
}
