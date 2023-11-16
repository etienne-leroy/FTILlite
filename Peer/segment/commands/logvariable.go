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
	"strings"

	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandLogVariable = "command_log_variable"

func LogVariable(s SegmentHost, args []string) (string, error) {
	label := args[0]

	label = strings.ReplaceAll(label, "~", " ")

	for _, v := range args[1:] {
		x, err := s.Variables().Get(variables.Handle(v))

		if err == nil {
			s.Log("%v: %v = %v\n", label, v, x)
		} else {
			s.Log("%v: %v not found\n", label, v)
		}
	}

	return Ack, nil
}
