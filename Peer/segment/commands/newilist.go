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

const CommandNewilist = "command_newilist" // command_newilist <hResult :: Handle→[]int64> <int64...> ⤶ <hResult :: Handle>

func Newilist(s SegmentHost, args []string) (string, error) {
	h := variables.Handle(args[0])
	values := args[1:]

	xs := make([]int64, len(values))
	for i := range values {
		x, err := strconv.ParseInt(values[i], 10, 64)
		if err != nil {
			return "", err
		}
		xs[i] = x
	}

	s.Variables().Set(h, types.NewFTIntegerArray(xs...))

	return fmt.Sprintf("array i %v", h), nil
}
