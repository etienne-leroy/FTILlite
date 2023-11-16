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

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandToList = "command_tolist" // command_tolist

type PythonArrayTypeVal interface {
	types.ArrayTypeVal

	PythonString() string
}

func ToPythonList(s SegmentHost, args []string) (string, error) {
	if s.Node().NodeID() != 0 {
		return "", errors.New("command only allowed on coordinator node")
	}

	h := variables.Handle(args[0])

	v, err := variables.GetAs[PythonArrayTypeVal](s.Variables(), h)
	if err != nil {
		return "", err
	}

	return v.PythonString(), nil
}
