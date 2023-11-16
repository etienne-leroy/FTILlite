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

const CommandArange = "command_arange" // command_arange <hResult :: Handle→[]int64> <hLength :: Handle→[1]int64>  ⤶ <hResult :: Handle>

func Arange(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hLength := variables.Handle(args[1])

	lengthArray, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hLength)
	if err != nil {
		return "", err
	}

	length, err := lengthArray.Single()
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, types.ArangeFTIntegerArray(length))

	return fmt.Sprintf("array i %s", hResult), nil
}
