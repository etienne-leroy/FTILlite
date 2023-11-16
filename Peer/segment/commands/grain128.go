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

const CommandGrain128aeadv2 = "command_grain128aeadv2"

func Grain128Aeadv2(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hKey := variables.Handle(args[1])
	hIV := variables.Handle(args[2])
	hSize := variables.Handle(args[3])
	hLength := variables.Handle(args[4])

	key, err := variables.GetAsBytes(s.Variables(), hKey)
	if err != nil {
		return "", err
	}

	iv, err := variables.GetAsBytes(s.Variables(), hIV)
	if err != nil {
		return "", err
	}
	width, err := variables.GetAsInteger(s.Variables(), hSize)
	if err != nil {
		return "", err
	}
	length, err := variables.GetAsInteger(s.Variables(), hLength)
	if err != nil {
		return "", err
	}

	targetBytes, err := crypto.Grain128Aeadv2(key, iv, width, length)
	if err != nil {
		return "", err
	}
	target, err := types.NewFTBytearrayArray(types.CalcBytearrayWidth(targetBytes), targetBytes...)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b%v %v", width, hTarget), nil
}
