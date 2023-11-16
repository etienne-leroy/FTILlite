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

const CommandContains = "command_contains" // command_contains

func Contains(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hTarget := variables.Handle(args[1])
	hValues := variables.Handle(args[2])

	target, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	values, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hValues)
	if err != nil {
		return "", err
	}

	if target.TypeCode() != values.TypeCode() {
		return "", errors.New("target and values are of a different type")
	}

	result, err := target.Contains(values)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, result)

	return fmt.Sprintf("array i %s", hResult), nil
}
