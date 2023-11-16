// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package commands

import "github.com/AUSTRAC/ftillite/Peer/segment/variables"

const CommandEqualInt = "command_equalint" // command_equalint

func EqualInt(s SegmentHost, args []string) (string, error) {
	h1 := variables.Handle(args[0])
	h2 := variables.Handle(args[1])

	var1, err := s.Variables().Get(h1)
	if err != nil {
		return "", err
	}

	var2, err := s.Variables().Get(h2)
	if err != nil {
		return "", err
	}

	if var1.Equals(var2) {
		return "bool 1", nil
	}
	return "bool 0", nil
}
