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

const CommandVerify = "command_verify"   // command_verify
const CommandNonZero = "command_nonzero" // command_nonzero <hValues :: Handle→[](int64)>

func Verify(s SegmentHost, args []string) (string, error) {
	hValues := variables.Handle(args[0])

	values, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hValues)
	if err != nil {
		return "", err
	}

	if values.NonZero() {
		return "bool 1", nil
	}

	return "bool 0", nil
}
