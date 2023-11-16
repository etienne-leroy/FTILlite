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

	"filippo.io/edwards25519"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const CommandNewArray = "command_newarray" // command_newarray <hResult :: Handle→[](int64|float64|[]byte)> <typecode> <hLength :: Handle→[1]int64> [ <value :: (int64|float64)> ]  ⤶ <hResult :: Handle>

func NewArray(s SegmentHost, args []string) (string, error) {
	hResult := variables.Handle(args[0])

	tc, err := types.ParseTypeCode(args[1])
	if err != nil {
		return "", err
	}

	hLength := variables.Handle(args[2])
	lengthArray, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), hLength)
	if err != nil {
		return "", err
	}
	length, err := lengthArray.Single()
	if err != nil {
		return "", err
	}

	switch tc.GetBase() {
	case types.IntegerB:
		xs := make([]int64, length)
		if len(args) > 3 {
			value, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), variables.Handle(args[3]))
			if err != nil {
				return "", err
			}

			x, err := value.Single()
			if err != nil {
				return "", err
			}

			for i := range xs {
				xs[i] = x
			}
		}
		s.Variables().Set(hResult, types.NewFTIntegerArray(xs...))
	case types.FloatB:
		xs := make([]float64, length)
		if len(args) > 3 {
			value, err := variables.GetAs[*types.FTFloatArray](s.Variables(), variables.Handle(args[3]))
			if err != nil {
				return "", err
			}

			x, err := value.Single()
			if err != nil {
				return "", err
			}

			for i := range xs {
				xs[i] = x
			}
		}

		s.Variables().Set(hResult, types.NewFTFloatArray(xs...))

	case types.BytearrayB:
		xs := make([][]byte, length)
		for i := range xs {
			xs[i] = make([]byte, tc.Length())
		}
		s.Variables().Set(hResult, types.NewFTBytearrayArrayOrPanic(tc.Length(), xs...))

	case types.Ed25519IntB:
		xs := make([]*edwards25519.Scalar, length)

		if len(args) > 3 {
			value, err := s.Variables().Get(variables.Handle(args[3]))
			if err != nil {
				return "", err
			}

			if vIntegerArr, ok := value.(*types.FTIntegerArray); ok {
				x, err := vIntegerArr.Single()
				if err != nil {
					return "", err
				}

				for i := range xs {
					xs[i] = types.Int64ToScalar(x)
				}
			} else if vEd25519IntArr, ok := value.(*types.FTEd25519IntArray); ok {
				x, err := vEd25519IntArr.Single()
				if err != nil {
					return "", err
				}
				for i := range xs {
					y := *x
					xs[i] = &y
				}
			} else {
				return "", errors.New("value must be a singleton array of Integer or Ed25519Int")
			}
		} else {
			for i := range xs {
				xs[i] = edwards25519.NewScalar()
			}
		}

		s.Variables().Set(hResult, types.NewFTEd25519IntArray(xs...))

	case types.Ed25519B:
		if !s.IsGPUAvailable() {
			return "", ErrEd25519Unavailable
		}

		var xs *types.Ed25519Array
		var err error

		if len(args) > 3 {
			var value types.TypeVal
			value, err = s.Variables().Get(variables.Handle(args[3]))
			if err != nil {
				return "", err
			}

			if vIntegerArr, ok := value.(*types.FTIntegerArray); ok {
				x, err := vIntegerArr.Single()
				if err != nil {
					return "", err
				}
				v := types.Int64ToScalar(x)
				xs, err = types.NewEd25519Array(length, v)
				if err != nil {
					return "", err
				}
			} else if vEd25519Arr, ok := value.(*types.Ed25519Array); ok && vEd25519Arr.Length() == 1 {
				xs, err = types.NewEd25519ArrayFromPoint(length, vEd25519Arr, 0)
			} else if vEd25519IntArr, ok := value.(*types.FTEd25519IntArray); ok {
				x, err := vEd25519IntArr.Single()
				if err != nil {
					return "", err
				}
				xs, err = types.NewEd25519Array(length, x)
				if err != nil {
					return "", err
				}
			} else {
				err = errors.New("value must be a singleton array of Integer, Ed25519Int or Ed25519")
			}
		} else {
			xs, err = types.NewEd25519Array(length, nil)
		}

		if err != nil {
			return "", err
		}

		s.Variables().Set(hResult, xs)

	default:
		panic("newarray not implemented for base type: " + tc.GetBase().String())
	}

	return fmt.Sprintf("array %s %s", tc, hResult), nil
}
