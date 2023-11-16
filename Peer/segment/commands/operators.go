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
	CommandEq       = "command_eq"       // command_eq <hResult :: Handle→[]int64> <hLHS :: Handle→[](int64|float64|[]byte)> <hRHS :: Handle→[](int64|float64|[]byte)>
	CommandNe       = "command_ne"       // command_ne <hResult :: Handle→[]int64> <hLHS :: Handle→[](int64|float64|[]byte)> <hRHS :: Handle→[](int64|float64|[]byte)>
	CommandLt       = "command_lt"       // command_lt <hResult :: Handle→[]int64> <hLHS :: Handle→[](int64|float64|[]byte)> <hRHS :: Handle→[](int64|float64|[]byte)>
	CommandGt       = "command_gt"       // command_gt <hResult :: Handle→[]int64> <hLHS :: Handle→[](int64|float64|[]byte)> <hRHS :: Handle→[](int64|float64|[]byte)>
	CommandLe       = "command_le"       // command_le <hResult :: Handle→[]int64> <hLHS :: Handle→[](int64|float64|[]byte)> <hRHS :: Handle→[](int64|float64|[]byte)>
	CommandGe       = "command_ge"       // command_ge <hResult :: Handle→[]int64> <hLHS :: Handle→[](int64|float64|[]byte)> <hRHS :: Handle→[](int64|float64|[]byte)>
	CommandNeg      = "command_neg"      // command_neg <hResult :: Handle→[](int64|float64|*big.Int|Ed25519)> <hSource :: Handle→[](int64|float64|*big.Int|Ed25519)>
	CommandAbs      = "command_abs"      // command_abs <hResult :: Handle→[](int64|float64)> <hSource :: Handle→[](int64|float64)>
	CommandFloor    = "command_floor"    // command_floor <hResult :: Handle→[]float64> <hSource :: Handle→[]float64>
	CommandCeil     = "command_ceil"     // command_ceil <hResult :: Handle→[]float64> <hSource :: Handle→[]float64>
	CommandRound    = "command_round"    // command_round <hResult :: Handle→[]float64> <hSource :: Handle→[]float64>
	CommandTrunc    = "command_trunc"    // command_trunc <hResult :: Handle→[]float64> <hSource :: Handle→[]float64>
	CommandAdd      = "command_add"      // command_add <hResult :: Handle→[](int64|float64)> <hLHS :: Handle→[](int64|float64)> <hRHS :: Handle→[](int64|float64)>
	CommandSub      = "command_sub"      // command_sub <hResult :: Handle→[](int64|float64)> <hLHS :: Handle→[](int64|float64)> <hRHS :: Handle→[](int64|float64)>
	CommandMul      = "command_mul"      // command_mul <hResult :: Handle→[](int64|float64)> <hLHS :: Handle→[](int64|float64)> <hRHS :: Handle→[](int64|float64)>
	CommandFloorDiv = "command_floordiv" // command_floordiv <hResult :: Handle→[]int64> <hLHS :: Handle→[]int64> <hRHS :: Handle→[]int64>
	CommandTrueDiv  = "command_truediv"  // command_truediv <hResult :: Handle→[]float64> <hLHS :: Handle→[](int64|float64)> <hRHS :: Handle→[](int64|float64)>
	CommandMod      = "command_mod"      // command_mod <hResult :: Handle→[]int64> <hLHS :: Handle→[]int64> <hRHS :: Handle→[]int64>
	CommandDivMod   = "command_divmod"   // command_divmod <hQuotientResult :: Handle→[]int64> <hRemainderResult :: Handle→[]int64> <hLHS :: Handle→[]int64> <hRHS :: Handle→[]int64>
	CommandPow      = "command_pow"      // command_pow <hResult :: Handle→[](int64|float64)> <hLHS :: Handle→[](int64|float64)> <hRHS :: Handle→[](int64|float64)>
	CommandLShift   = "command_lshift"   // command_lshift <hResult :: Handle→[](int64|[]byte])> <hLHS :: Handle→[](int64|[]byte)> <hRHS :: Handle→[]int64>
	CommandRShift   = "command_rshift"   // command_rshift <hResult :: Handle→[](int64|[]byte])> <hLHS :: Handle→[](int64|[]byte)> <hRHS :: Handle→[]int64>
	CommandAnd      = "command_and"      // command_and <hResult :: Handle→[](int64|[]byte])> <hLHS :: Handle→[](int64|[]byte)> <hRHS :: Handle→[](int64|[]byte)>
	CommandOr       = "command_or"       // command_or <hResult :: Handle→[](int64|[]byte])> <hLHS :: Handle→[](int64|[]byte)> <hRHS :: Handle→[](int64|[]byte)>
	CommandXor      = "command_xor"      // command_xor <hResult :: Handle→[](int64|[]byte])> <hLHS :: Handle→[](int64|[]byte)> <hRHS :: Handle→[](int64|[]byte)>
	CommandInvert   = "command_invert"   // command_invert <hResult :: Handle→[](int64|[]byte])> <hSource :: Handle→[](int64|[]byte)>
	CommandNearest  = "command_nearest"  // command_nearest <hResult :: Handle→[]int64> <hSource :: Handle→[]float64>
	CommandExp      = "command_exp"      // command_exp <hResult :: Handle→[]float64> <hSource :: Handle→[]float64>
	CommandLog      = "command_log"      // command_log <hResult :: Handle→[]float64> <hSource :: Handle→[]float64>
	CommandSin      = "command_sin"      // command_sin <hResult :: Handle→[]float64> <hSource :: Handle→[]float64>
	CommandCos      = "command_cos"      // command_cos <hResult :: Handle→[]float64> <hSource :: Handle→[]float64>
)

var Eq = binaryOperator(func(as, bs types.ArrayTypeVal) (types.ArrayTypeVal, error) { return as.Eq(bs) })
var Ne = binaryOperator(func(as, bs types.ArrayTypeVal) (types.ArrayTypeVal, error) { return as.Ne(bs) })
var Gt = binaryOperator(func(as, bs types.ArrayComparableTypeVal) (types.ArrayTypeVal, error) { return as.Gt(bs) })
var Lt = binaryOperator(func(as, bs types.ArrayComparableTypeVal) (types.ArrayTypeVal, error) { return as.Lt(bs) })
var Ge = binaryOperator(func(as, bs types.ArrayComparableTypeVal) (types.ArrayTypeVal, error) { return as.Ge(bs) })
var Le = binaryOperator(func(as, bs types.ArrayComparableTypeVal) (types.ArrayTypeVal, error) { return as.Le(bs) })

var Neg = unaryOperator(func(as types.ArrayNegTypeVal) (types.ArrayNegTypeVal, error) { return as.Neg() })
var Abs = unaryOperator(func(as types.ArrayAbsTypeVal) (types.ArrayAbsTypeVal, error) { return as.Abs() })

var Floor = unaryOperator(func(as *types.FTFloatArray) (*types.FTIntegerArray, error) { return as.Floor() })
var Ceil = unaryOperator(func(as *types.FTFloatArray) (*types.FTIntegerArray, error) { return as.Ceil() })
var Round = unaryOperator(func(as *types.FTFloatArray) (*types.FTIntegerArray, error) { return as.Round() })
var Trunc = unaryOperator(func(as *types.FTFloatArray) (*types.FTFloatArray, error) { return as.Trunc() })

var Add = binaryOperator(func(as, bs types.ArrayAddSubMulTypeVal) (types.ArrayTypeVal, error) { return as.Add(bs) })
var Sub = binaryOperator(func(as, bs types.ArrayAddSubMulTypeVal) (types.ArrayTypeVal, error) { return as.Sub(bs) })
var Mul = binaryOperator(func(as, bs types.ArrayAddSubMulTypeVal) (types.ArrayTypeVal, error) { return as.Mul(bs) })

var FloorDiv = binaryOperator(func(as, bs types.ArrayFloorDivTypeVal) (types.ArrayTypeVal, error) { return as.FloorDiv(bs) })
var TrueDiv = binaryOperator(func(as, bs types.ArrayTrueDivTypeVal) (types.ArrayTypeVal, error) { return as.TrueDiv(bs) })
var Mod = binaryOperator(func(as, bs *types.FTIntegerArray) (*types.FTIntegerArray, error) { return as.Mod(bs) })
var DivMod = binaryOperator2(func(as, bs *types.FTIntegerArray) (*types.FTIntegerArray, *types.FTIntegerArray, error) {
	return as.DivMod(bs)
})
var Pow = binaryOperator(func(as, bs types.ArrayPowTypeVal) (types.ArrayTypeVal, error) { return as.Pow(bs) })

var LShift = binaryOperator(func(as types.ArrayBitwiseTypeVal, bs types.ArrayTypeVal) (types.ArrayTypeVal, error) {
	return as.LShift(bs)
})
var RShift = binaryOperator(func(as types.ArrayBitwiseTypeVal, bs types.ArrayTypeVal) (types.ArrayTypeVal, error) {
	return as.RShift(bs)
})
var And = binaryOperator(func(as, bs types.ArrayBitwiseTypeVal) (types.ArrayTypeVal, error) { return as.And(bs) })
var Or = binaryOperator(func(as, bs types.ArrayBitwiseTypeVal) (types.ArrayTypeVal, error) { return as.Or(bs) })
var Xor = binaryOperator(func(as, bs types.ArrayBitwiseTypeVal) (types.ArrayTypeVal, error) { return as.Xor(bs) })
var Invert = unaryOperator(func(as types.ArrayBitwiseTypeVal) (types.ArrayTypeVal, error) { return as.Invert() })

var Nearest = unaryOperator(func(as *types.FTIntegerArray) (*types.FTFloatArray, error) { return as.Nearest() })
var Exp = unaryOperator(func(as *types.FTFloatArray) (*types.FTFloatArray, error) { return as.Exp() })
var Log = unaryOperator(func(as *types.FTFloatArray) (*types.FTFloatArray, error) { return as.Log() })
var Sin = unaryOperator(func(as *types.FTFloatArray) (*types.FTFloatArray, error) { return as.Sin() })
var Cos = unaryOperator(func(as *types.FTFloatArray) (*types.FTFloatArray, error) { return as.Cos() })

func unaryOperator[T types.ArrayTypeVal, U types.ArrayTypeVal](f func(as T) (U, error)) CommandFunc {
	return func(s SegmentHost, args []string) (string, error) {
		hResult := variables.Handle(args[0])
		hValue := variables.Handle(args[1])

		value, err := variables.GetAs[T](s.Variables(), hValue)

		if err != nil {
			return "", err
		}

		result, err := f(value)
		if err != nil {
			return "", err
		}

		s.Variables().Set(hResult, result)

		return fmt.Sprintf("array %s %s", result.TypeCode(), hResult), err
	}
}

func binaryOperator[T types.ArrayTypeVal, U types.ArrayTypeVal, V types.ArrayTypeVal](f func(as T, bs U) (V, error)) CommandFunc {
	return func(s SegmentHost, args []string) (string, error) {
		hResult := variables.Handle(args[0])
		hAs := variables.Handle(args[1])
		hBs := variables.Handle(args[2])

		as, err := variables.GetAs[T](s.Variables(), hAs)

		if err != nil {
			return "", err
		}
		bs, err := variables.GetAs[U](s.Variables(), hBs)
		if err != nil {
			return "", err
		}

		result, err := f(as, bs)
		if err != nil {
			return "", err
		}

		s.Variables().Set(hResult, result)

		return fmt.Sprintf("array %s %s", result.TypeCode(), hResult), err
	}
}

func binaryOperator2[T types.ArrayTypeVal, U types.ArrayTypeVal, V types.ArrayTypeVal, W types.ArrayTypeVal](f func(as T, bs U) (V, W, error)) CommandFunc {
	return func(s SegmentHost, args []string) (string, error) {
		hResultU := variables.Handle(args[0])
		hResultV := variables.Handle(args[1])
		hAs := variables.Handle(args[2])
		hBs := variables.Handle(args[3])

		as, err := variables.GetAs[T](s.Variables(), hAs)
		if err != nil {
			return "", err
		}

		bs, err := variables.GetAs[U](s.Variables(), hBs)
		if err != nil {
			return "", err
		}

		us, vs, err := f(as, bs)
		if err != nil {
			return "", err
		}

		s.Variables().Set(hResultU, us)
		s.Variables().Set(hResultV, vs)

		return fmt.Sprintf("array %s %s array %s %s", us.TypeCode(), hResultU, vs.TypeCode(), hResultV), err
	}
}
