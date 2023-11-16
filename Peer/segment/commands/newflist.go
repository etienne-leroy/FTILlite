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

const CommandNewflist = "command_newflist" // command_newflist <hResult :: Handle→[]float64> <float64...>  ⤶ <hResult :: Handle>

func Newflist(s SegmentHost, args []string) (string, error) {
	h := variables.Handle(args[0])
	values := args[1:]

	xs := make([]float64, len(values))
	for i := range values {
		x, err := strconv.ParseFloat(values[i], 64)
		if err != nil {
			return "", err
		}
		xs[i] = x
	}

	s.Variables().Set(h, types.NewFTFloatArray(xs...))

	return fmt.Sprintf("array f %v", h), nil
}
