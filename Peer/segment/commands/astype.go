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

const CommandAsType = "command_astype" // command_astype

func AsType(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hCurrent := variables.Handle(args[1])
	tc := types.TypeCode(args[2])

	value, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hCurrent)
	if err != nil {
		return "", err
	}

	target, err := value.AsType(tc)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array %s %s", tc, hTarget), nil
}
