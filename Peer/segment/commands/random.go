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
	CommandRandomArray = "command_randomarray" /* command_randomarray <hResult :: Handle→[]int64> i <hLength :: Handle→[1]int64> <hMin :: Handle→[1]int64> <hMax :: Handle→[1]int64>
												  command_randomarray <hResult :: Handle→[]float64> f <hLength :: Handle→[1]int64> <hMin :: Handle→[1]float64> <hMax :: Handle→[1]float64>
	                                              command_randomarray <hResult :: Handle→[][N]byte> bN <hLength :: Handle→[1]int64>
	*/
	CommandRandomPerm = "command_randomperm" // command_randperm
)

func RandomArray(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hLength := variables.Handle(args[2])

	lengthArray, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hLength)
	if err != nil {
		return "", err
	}
	length, err := lengthArray.Single()
	if err != nil {
		return "", err
	}

	tc, err := types.ParseTypeCode(args[1])
	if err != nil {
		return "", err
	}

	switch tc.GetBase() {
	case types.IntegerB:
		minArray, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), variables.Handle(args[3]))
		if err != nil {
			return "", err
		}
		min, err := minArray.Single()
		if err != nil {
			return "", err
		}
		maxArray, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), variables.Handle(args[4]))
		if err != nil {
			return "", err
		}
		max, err := maxArray.Single()
		if err != nil {
			return "", err
		}

		xs, err := types.NewRandomFTIntegerArray(min, max, length)
		if err != nil {
			return "", err
		}

		s.Variables().Set(hResult, xs)

	case types.FloatB:
		minArray, err := variables.GetAs[*types.FTFloatArray](s.Variables(), variables.Handle(args[3]))
		if err != nil {
			return "", err
		}
		min, err := minArray.Single()
		if err != nil {
			return "", err
		}
		maxArray, err := variables.GetAs[*types.FTFloatArray](s.Variables(), variables.Handle(args[4]))
		if err != nil {
			return "", err
		}
		max, err := maxArray.Single()
		if err != nil {
			return "", err
		}

		xs, err := types.NewRandomFTFloatArray(min, max, length)
		if err != nil {
			return "", err
		}

		s.Variables().Set(hResult, xs)

	case types.Ed25519IntB:

		nonZero := true
		if len(args) > 3 {
			nonZero = args[3] == "True"
		}

		xs, err := types.NewRandomFTEd25519IntArray(nonZero, length)
		if err != nil {
			return "", err
		}

		s.Variables().Set(hResult, xs)

	case types.BytearrayB:
		size := tc.Length()

		xs, err := types.NewRandomFTBytearrayArray(int64(size), length)
		if err != nil {
			return "", err
		}

		s.Variables().Set(hResult, xs)

	default:
		return "", fmt.Errorf("randomarray not supported on %v", tc)
	}

	return fmt.Sprintf("array %s %s", tc, hResult), nil
}

func RandomPerm(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])
	hLength := variables.Handle(args[1])
	hN := variables.Handle(args[2])

	lengthArray, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hLength)
	if err != nil {
		return "", err
	}
	length, err := lengthArray.Single()
	if err != nil {
		return "", err
	}

	nArray, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hN)
	if err != nil {
		return "", err
	}
	n, err := nArray.Single()
	if err != nil {
		return "", err
	}

	if n < length {
		return "", fmt.Errorf("n must not be less than length, n: %d, length: %d", n, length)
	}

	xs, err := types.NewRandomPermFTIntegerArray(length, n)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hResult, xs)

	return fmt.Sprintf("array i %s", hResult), nil
}
