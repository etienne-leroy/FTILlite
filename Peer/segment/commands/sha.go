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

	"github.com/AUSTRAC/ftillite/Peer/segment/crypto"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandSHA3256 = "command_sha3_256"

func Sha3_256(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hSource := variables.Handle(args[1])

	source, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hSource)
	if err != nil {
		return "", err
	}

	targetBytes := crypto.Sha3256Sum(source.Values())
	target, err := types.NewFTBytearrayArray(32, targetBytes...)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b32 %v", hTarget), nil
}
