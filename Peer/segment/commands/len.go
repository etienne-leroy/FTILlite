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
	"errors"
	"fmt"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandLen = "command_len"     // command_len
const CommandPyLen = "command_pylen" // command_pylen

func Len(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, types.NewFTIntegerArray(target.Length()))

	return fmt.Sprintf("array i %s", hResult), err
}

func PyLen(s SegmentHost, args []string) (string, error) {
	if s.Node().NodeID() != 0 {
		return "", errors.New("command only allowed on coordinator node")
	}

	h := variables.Handle(args[0])

	v, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), h)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("int %d", v.Length()), nil
}
