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

const CommandIndex = "command_index" // command_index <hResult :: Handle->[](int64)>

func Index(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, target.Index())

	return fmt.Sprintf("array i %v", hResult), nil
}
