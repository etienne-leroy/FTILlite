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

const CommandDelItem = "command_delitem"

func DelItem(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hIndices := variables.Handle(args[1])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	indices, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hIndices)
	if err != nil {
		return "", err
	}

	err = target.Remove(indices)
	if err != nil {
		return "", err
	}

	return Ack, nil
}
