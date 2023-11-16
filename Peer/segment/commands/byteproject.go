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
	"strconv"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandByteProject = "command_byteproject" // command_byteproject <hResult :: Handle→[]([]byte)> <hSource :: Handle→[]([]byte)> <size :: int64> ( <i :: int64> <j :: int64> )+  ⤶ b<size> <hResult :: Handle>

func ByteProject(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hSource := variables.Handle(args[1])
	strSize := args[2]
	hMappingKeys := variables.Handle(args[3])
	hMappingValues := variables.Handle(args[4])

	source, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hSource)
	if err != nil {
		return "", err
	}

	size, err := strconv.Atoi(strSize)
	if err != nil {
		return "", err
	}

	indexes, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hMappingKeys)
	if err != nil {
		return "", err
	}
	values, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hMappingValues)
	if err != nil {
		return "", err
	}

	result, err := source.Project(int64(size), indexes, values)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, result)

	return fmt.Sprintf("array b%v %v", size, hResult), nil
}
