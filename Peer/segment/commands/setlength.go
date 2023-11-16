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

const CommandSetLength = "command_setlength" // command_setlength <hTarget :: Handle→[](int64|float64|[]byte)> <hLength :: Handle→[1](int64)>

func SetLength(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hLength := variables.Handle(args[1])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	lengthArray, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hLength)
	if err != nil {
		return "", err
	}

	length, err := lengthArray.Single()
	if err != nil {
		return "", err
	}

	err = target.SetLength(length)
	if err != nil {
		return "", err
	}

	return Ack, nil
}
