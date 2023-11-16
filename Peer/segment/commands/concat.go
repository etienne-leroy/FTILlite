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

const CommandConcat = "command_concat" // command_concat <hResult :: Handle→[]([]byte)> <hTarget :: Handle→[]([]byte)> <hValues :: Handle→[]([]byte)> ⤶ b<size> <hResult :: Handle>

func Concat(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])
	hValues := variables.Handle(args[2])

	target, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}
	values, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hValues)
	if err != nil {
		return "", err
	}

	result, err := target.Concat(values)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, result)

	return fmt.Sprintf("array b%v %v", result.Width(), hResult), nil
}
