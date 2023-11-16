// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package segment

import (
	"math/big"
	"testing"

	"github.com/AUSTRAC/ftillite/Peer/segment/commands"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
)

func TestCommandOpEq_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "3", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "2", "3")
	AssertCommand(t, s, commands.CommandEq, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 0, 1))
}
func TestCommandOpEq_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewflist, "2", "0.5", "2", "1.5")
	AssertCommand(t, s, commands.CommandEq, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 0))
}
func TestCommandOpEq_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(1, 4, 3))

	AssertCommand(t, s, commands.CommandEq, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 0, 1))
}
func TestCommandOpEq_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(4, []byte{0, 1, 3, 4}, []byte{0, 1, 3, 4}, []byte{0, 1, 3, 4}))
	s.SetVariable("2", types.NewFTBytearrayArrayOrPanic(4, []byte{0, 1, 3, 4}, []byte{1, 2, 3, 4}, []byte{0, 1, 3, 4}))

	AssertCommand(t, s, commands.CommandEq, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 0, 1))
}
func TestCommandOpEq_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.SetVariable("1", types.NewEd25519ArrayFromInt64sOrPanic(10, 20, 30))
	s.SetVariable("2", types.NewEd25519ArrayFromInt64sOrPanic(10, 20, 30))
	s.SetVariable("3", types.NewEd25519ArrayFromInt64sOrPanic(10, 40, 30))

	AssertCommand(t, s, commands.CommandEq, "4", "1", "2")
	AssertValue(t, s, "4", types.NewFTIntegerArray(1, 1, 1))

	AssertCommand(t, s, commands.CommandEq, "5", "1", "3")
	AssertValue(t, s, "5", types.NewFTIntegerArray(1, 0, 1))
}

func TestCommandOpNe_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "3", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNe, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 0))
}
func TestCommandOpNe_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewflist, "2", "0.5", "2", "1.5")
	AssertCommand(t, s, commands.CommandNe, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 0, 1))
}
func TestCommandOpNe_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(4, []byte{0, 1, 3, 4}, []byte{0, 1, 3, 4}, []byte{0, 1, 3, 4}))
	s.SetVariable("2", types.NewFTBytearrayArrayOrPanic(4, []byte{0, 1, 3, 4}, []byte{1, 2, 3, 4}, []byte{0, 1, 3, 4}))

	AssertCommand(t, s, commands.CommandNe, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 0))
}
func TestCommandOpNe_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(1, 4, 3))

	AssertCommand(t, s, commands.CommandNe, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 0))
}
func TestCommandOpNe_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(10, 20, 30))
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(10, 20, 30))
	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic(10, 40, 30))

	AssertCommand(t, s, commands.CommandNe, "4", "1", "2")
	AssertValue(t, s, "4", types.NewFTIntegerArray(0, 0, 0))

	AssertCommand(t, s, commands.CommandNe, "5", "1", "3")
	AssertValue(t, s, "5", types.NewFTIntegerArray(0, 1, 0))
}

func TestCommandOpLt_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "3", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "4", "2")
	AssertCommand(t, s, commands.CommandLt, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 0))
}
func TestCommandOpLt_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewflist, "2", "1", "3", "2")
	AssertCommand(t, s, commands.CommandLt, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 0))
}

func TestCommandOpGt_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "3", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "4", "2")
	AssertCommand(t, s, commands.CommandGt, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 0, 1))
}
func TestCommandOpGt_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewflist, "2", "1", "3", "2")

	AssertCommand(t, s, commands.CommandGt, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 0, 1))
}

func TestCommandOpLe_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "3", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "4", "2")
	AssertCommand(t, s, commands.CommandLe, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 1, 0))
}
func TestCommandOpLe_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewflist, "2", "1", "3", "2")
	AssertCommand(t, s, commands.CommandLe, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 1, 0))
}

func TestCommandOpGe_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "3", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "4", "2")
	AssertCommand(t, s, commands.CommandGe, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 0, 1))
}
func TestCommandOpGe_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewflist, "2", "1", "3", "2")
	AssertCommand(t, s, commands.CommandGe, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 0, 1))
}

func TestCommandOpNeg_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "3", "-3")
	AssertCommand(t, s, commands.CommandNeg, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(-1, -3, 3))
}
func TestCommandOpNeg_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "3", "-3")
	AssertCommand(t, s, commands.CommandNeg, "2", "1")

	AssertValue(t, s, "2", types.NewFTFloatArray(-1, -3, 3))
}
func TestCommandOpNeg_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 3, -3))
	AssertCommand(t, s, commands.CommandNeg, "2", "1")

	AssertValue(t, s, "2", types.NewFTEd25519IntArrayFromInt64s(-1, -3, 3))
}
func TestCommandOpNeg_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	original := types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3)
	s.Variables().Set("1", original)

	AssertCommand(t, s, commands.CommandNeg, "2", "1")
	AssertValueNot(t, s, "2", original)

	AssertCommand(t, s, commands.CommandNeg, "3", "2")
	AssertValue(t, s, "3", original)
}
func TestCommandOpAbs_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "-1", "3", "-3")
	AssertCommand(t, s, commands.CommandAbs, "2", "1")
	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 3, 3))
}
func TestCommandOpAbs_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "-1.5", "3.5", "-3.5")
	AssertCommand(t, s, commands.CommandAbs, "2", "1")
	AssertValue(t, s, "2", types.NewFTFloatArray(1.5, 3.5, 3.5))
}

func TestCommandOpFloor_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1.0", "1.3", "1.7")
	AssertCommand(t, s, commands.CommandFloor, "2", "1")
	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 1, 1))
}
func TestCommandOpCeil_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1.0", "1.3", "1.7")
	AssertCommand(t, s, commands.CommandCeil, "2", "1")
	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 2, 2))
}
func TestCommandOpRound_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1.0", "1.3", "1.7")
	AssertCommand(t, s, commands.CommandRound, "2", "1")
	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 1, 2))
}
func TestCommandOpTrunc_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "-1.0", "2.3", "14.7")
	AssertCommand(t, s, commands.CommandTrunc, "2", "1")
	AssertValue(t, s, "2", types.NewFTFloatArray(1, -1, 2, 14))
}

func TestCommandOpAdd_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "2", "3")

	AssertCommand(t, s, commands.CommandAdd, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(2, 4, 6))
}
func TestCommandOpAdd_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "0.5", "1", "1.5")
	AssertCommand(t, s, commands.CommandNewflist, "2", "0.5", "1", "1.5")

	AssertCommand(t, s, commands.CommandAdd, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTFloatArray(1, 2, 3))
}
func TestCommandOpAdd_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))

	AssertCommand(t, s, commands.CommandAdd, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTEd25519IntArrayFromInt64s(2, 4, 6))
}
func TestCommandOpAdd_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(10, 20, 30))
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(20, 30, 40))

	AssertCommand(t, s, commands.CommandAdd, "3", "1", "2")

	AssertValue(t, s, "3", types.NewEd25519ArrayFromInt64sOrPanic(30, 50, 70))
}

func TestCommandOpSub_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "4", "5")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "2", "3")
	AssertCommand(t, s, commands.CommandSub, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(2, 2, 2))
}
func TestCommandOpSub_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewflist, "2", "0.5", "1", "1.5")
	AssertCommand(t, s, commands.CommandSub, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTFloatArray(0.5, 1, 1.5))
}
func TestCommandOpSub_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(3, 4, 5))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
	AssertCommand(t, s, commands.CommandSub, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTEd25519IntArrayFromInt64s(2, 2, 2))
}
func TestCommandOpSub_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(20, 30, 40))
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(10, 20, 20))
	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic(10, 10, 20))

	AssertCommand(t, s, commands.CommandSub, "4", "1", "2")
	AssertCommand(t, s, commands.CommandEq, "5", "3", "4")

	AssertValue(t, s, "5", types.NewFTIntegerArray(1, 1, 1))
}

func TestCommandOpMul_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "4", "5")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "2", "3")
	AssertCommand(t, s, commands.CommandMul, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(3, 8, 15))
}
func TestCommandOpMul_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewflist, "2", "0.5", "1", "1.5")
	AssertCommand(t, s, commands.CommandMul, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTFloatArray(0.5, 2, 4.5))
}
func TestCommandOpMul_Ed25519Int(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(3, 4, 5))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
	AssertCommand(t, s, commands.CommandMul, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTEd25519IntArrayFromInt64s(3, 8, 15))

	s.Variables().Set("4", types.NewFTEd25519IntArrayFromInt64s(3, 4, 5))
	s.Variables().Set("5", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3))
	AssertCommand(t, s, commands.CommandMul, "6", "4", "5")

	AssertValue(t, s, "6", types.NewEd25519ArrayFromInt64sOrPanic(3, 8, 15))
}
func TestCommandOpMul_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(2, 3, 4))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
	AssertCommand(t, s, commands.CommandMul, "3", "1", "2")

	AssertValue(t, s, "3", types.NewEd25519ArrayFromInt64sOrPanic(2, 6, 12))
}

func TestCommandOpFloordiv_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "4", "5", "-3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "2", "2", "2")

	AssertCommand(t, s, commands.CommandFloorDiv, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 2, 2, -2))
}
func TestCommandOpFloorDiv_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(4, 5, 5, -15, -15))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(2, 2, 5, 5, -15))

	AssertCommand(t, s, commands.CommandFloorDiv, "3", "1", "2")

	// L = 2**252 + 27742317777372353535851937790883648493
	// -15 ≡ 7237005577332262213973186563042994240857116359379907606001950938285454250974 (mod L)
	b, _ := big.NewInt(0).SetString("1447401115466452442794637312608598848171423271875981521200390187657090850194", 10)

	AssertValue(t, s, "3", types.NewFTEd25519IntArray(
		types.Int64ToScalar(2),
		types.Int64ToScalar(2),
		types.Int64ToScalar(1),
		types.BigIntToScalar(b),
		types.Int64ToScalar(1),
	))
}
func TestCommandOpTrueDiv_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(4, 5, 5, -15))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(2, 2, 5, 5))

	AssertCommand(t, s, commands.CommandTrueDiv, "3", "1", "2")

	/*
		5 / 2 explained:

		The operations are done modulo L = 2**252 + 27742317777372353535851937790883648493 (which is a value coming
		from the elliptic curve and is a constant from the point of view of the cryptosystem; the curve is chosen
		specifically so that L is prime), so the result will always be in the range [0, L). The value that we want
		to call "a half" is some value s which satisfies 2*s = 1 (mod L) which happens to be
		3618502788666131106986593281521497120428558179689953803000975469142727125495.

		This s can be calculated with the "extended greatest common divisor" algorithm (XGCD) which, given
		numbers a and b gives you three integers d, s and t which satisfy d = a*s + b*t where d will be
		GCD(a, b). In this case we use a = 2 and b = L, and their GCD has to be 1 because L is prime.
		So XGCD gives us s and t satisfying 1 = 2*s + L*t which is precisely the condition above that
		2*s = 1 (mod L), so s = half modulo this L.

		Then you get 5/2 by taking our halfs = 3618502788666131106986593281521497120428558179689953803000975469142727125495,
		multiplying by 5 and reducing modulo L again, which gives the final result
		3618502788666131106986593281521497120428558179689953803000975469142727125497.

	*/

	r2, _ := big.NewInt(0).SetString("3618502788666131106986593281521497120428558179689953803000975469142727125497", 10)

	AssertValue(t, s, "3", types.NewFTEd25519IntArray(
		types.Int64ToScalar(2),
		types.BigIntToScalar(r2),
		types.Int64ToScalar(1),
		types.Int64ToScalar(-3),
	))
}

func TestCommandOpTruediv_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "4", "5")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "2", "2")

	AssertCommand(t, s, commands.CommandTrueDiv, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTFloatArray(1.5, 2, 2.5))
}
func TestCommandOpTruediv_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "3", "4", "5")
	AssertCommand(t, s, commands.CommandNewflist, "2", "2", "2", "2")

	AssertCommand(t, s, commands.CommandTrueDiv, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTFloatArray(1.5, 2, 2.5))
}
func TestCommandOpMod_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10", "10", "10", "20")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "3", "4", "-15")

	AssertCommand(t, s, commands.CommandMod, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 2, -10))
}
func TestCommandOpDivmod_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10", "10", "10", "20")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "3", "4", "-15")

	AssertCommand(t, s, commands.CommandDivMod, "3", "4", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(5, 3, 2, -2))
	AssertValue(t, s, "4", types.NewFTIntegerArray(0, 1, 2, -10))
}
func TestCommandOpPow_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "4", "5", "5", "1", "100", "-2")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "2", "2", "-1", "-2", "0", "3")

	AssertCommand(t, s, commands.CommandPow, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(9, 16, 25, 0, 1, 1, -8))
}
func TestCommandOpPow_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "3.5", "4.5", "5.5", "2.5", "-2.5")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "3", "4", "-2", "-3")

	AssertCommand(t, s, commands.CommandPow, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTFloatArray(12.25, 91.125, 915.0625, 0.16, -0.064))
}
func TestCommandOpPow_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(2, 1, 2, 3, 4))
	AssertCommand(t, s, commands.CommandNewilist, "2", "-2", "-1", "0", "1", "2")

	AssertCommand(t, s, commands.CommandPow, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTEd25519IntArrayFromInt64s(0, 1, 1, 3, 16))
}

func TestCommandOpLShift_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "2", "3")

	AssertCommand(t, s, commands.CommandLShift, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(2, 8, 24))
}
func TestCommandOpLShift_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0000_0000, 0b0000_1101},
		[]byte{0b0000_0000, 0b0000_1101, 0b0000_1101},
		[]byte{0b1100_0000, 0b1100_1101, 0b1100_1101},
	))
	AssertCommand(t, s, commands.CommandNewilist, "2", "10", "9", "2")
	AssertCommand(t, s, commands.CommandLShift, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
		[]byte{0b0000_0011, 0b0011_0111, 0b0011_0100},
	))
}

func TestCommandOpRShift_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "2", "8", "24")
	AssertCommand(t, s, commands.CommandNewilist, "2", "1", "2", "3")

	AssertCommand(t, s, commands.CommandRShift, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 2, 3))
}
func TestCommandOpRShift_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b1111_0000},
	))
	AssertCommand(t, s, commands.CommandNewilist, "2", "10", "9", "5")

	AssertCommand(t, s, commands.CommandRShift, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0000_0000, 0b0000_1101},
		[]byte{0b0000_0000, 0b0000_1101, 0b0000_1101},
		[]byte{0b0000_0000, 0b1101_0000, 0b1101_0111},
	))
}
func TestCommandOpAnd_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "8", "5")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "2", "4")

	AssertCommand(t, s, commands.CommandAnd, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(2, 0, 4))
}
func TestCommandOpAnd_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100}))

	s.SetVariable("2", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0011_0100},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000}))

	AssertCommand(t, s, commands.CommandAnd, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
	))
}
func TestCommandOpOr_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "8", "5")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "2", "4")

	AssertCommand(t, s, commands.CommandOr, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(3, 10, 5))
}
func TestCommandOpOr_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100}))

	s.SetVariable("2", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0011_0100},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000}))

	AssertCommand(t, s, commands.CommandOr, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0011_0100},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100},
	))
}
func TestCommandOpXor_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "8", "5")
	AssertCommand(t, s, commands.CommandNewilist, "2", "2", "2", "4")

	AssertCommand(t, s, commands.CommandXor, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 10, 1))
}
func TestCommandOpXor_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100}))

	s.SetVariable("2", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0011_0100},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000}))

	AssertCommand(t, s, commands.CommandXor, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0000_0000, 0b0011_0100},
		[]byte{0b0000_0000, 0b0000_0000, 0b0011_0100},
	))
}
func TestCommandOpInvert_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3", "8", "5")
	AssertCommand(t, s, commands.CommandInvert, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(-4, -9, -6))
}
func TestCommandOpInvert_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100}))

	AssertCommand(t, s, commands.CommandInvert, "2", "1")

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b1111_1111, 0b11001011, 0b1111_1111},
		[]byte{0b1110_0101, 0b1110_0101, 0b1100_1011},
	))
}
func TestCommandOpNearest_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandNearest, "2", "1")
	AssertValue(t, s, "2", types.NewFTFloatArray(1, 2, 3))
}
func TestCommandOpExp_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandExp, "2", "1")
	AssertValue(t, s, "2", types.NewFTFloatArray(2.718281828459045, 7.38905609893065, 20.085536923187668))
}
func TestCommandOpLog_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandLog, "2", "1")
	AssertValue(t, s, "2", types.NewFTFloatArray(0, 0.6931471805599453, 1.0986122886681096))
}
func TestCommandOpSin_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandSin, "2", "1")
	AssertValue(t, s, "2", types.NewFTFloatArray(0.8414709848078965, 0.9092974268256816, 0.1411200080598672))
}
func TestCommandOpCos_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandCos, "2", "1")
	AssertValue(t, s, "2", types.NewFTFloatArray(0.5403023058681398, -0.4161468365471424, -0.9899924966004454))
}
