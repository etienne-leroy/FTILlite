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

const (
	CommandCalcBroadcastLength = "command_calc_broadcast_length" // command_calc_broadcast_length
	CommandBroadcastValue      = "command_broadcast_value"       // command_broadcast_value
)

func CalcBroadcastLength(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hLengths := args[1:]

	// If all parameters are of length 1, the return value is 1.
	// Otherwise, the return value is the maximum among all parameters that are not of length 1.
	maxLength := int64(-1)

	for _, h := range hLengths {
		tmpLength, err := variables.GetAsInteger(s.Variables(), variables.Handle(h))
		if err != nil {
			return "", err
		}

		if tmpLength != 1 && tmpLength > maxLength {
			maxLength = tmpLength
		}
	}

	if maxLength == -1 {
		maxLength = 1
	}

	s.Variables().Set(hResult, types.NewFTIntegerArray(maxLength))

	return fmt.Sprintf("array i %s", hResult), nil
}

func BroadcastValue(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hLen := variables.Handle(args[1])

	newLength, err := variables.GetAsInteger(s.Variables(), hLen)
	if err != nil {
		return "", err
	} else if newLength == 1 || newLength < 0 {
		return "", fmt.Errorf("invalid broadcast length %d", newLength)
	}

	v, err := variables.GetAs[types.ArrayTypeVal](s.Variables(), hTarget)
	if err != nil {
		return "", err
	}

	err = v.Broadcast(newLength)
	if err != nil {
		return "", err
	}

	return Ack, nil
}
