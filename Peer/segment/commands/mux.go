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

const CommandMux = "command_mux" // command_mux <hResult :: Handle→[]> <hCond :: Handle→[]Int64> <hIfTrue :: Handle→[]> <hIfFalse :: Handle→[]>

func Mux(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hCond := variables.Handle(args[1])
	hIfTrue := variables.Handle(args[2])
	hIfFalse := variables.Handle(args[3])

	cond, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hCond)
	if err != nil {
		return "", err
	}

	ifTrue, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hIfTrue)
	if err != nil {
		return "", err
	}
	ifFalse, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hIfFalse)
	if err != nil {
		return "", err
	}

	if ifTrue.TypeCode() != ifFalse.TypeCode() {
		return "", fmt.Errorf("iftrue and iffalse are not of the same type: %v and %v", ifTrue.TypeCode(), ifFalse.TypeCode())
	}

	results, err := ifTrue.Mux(cond, ifFalse)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, results)

	return fmt.Sprintf("array %s %s", results.TypeCode(), hResult), nil
}
