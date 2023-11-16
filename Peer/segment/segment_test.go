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
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"strconv"
	"testing"

	"filippo.io/edwards25519"
	"github.com/AUSTRAC/ftillite/Peer/segment/commands"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
	"github.com/stretchr/testify/assert"
)

const pickleTable = `CREATE TABLE pickle (destination character varying, dtype character varying(20), handle character varying(20), 
opcode character varying(10), data bytea, elementindex integer, chunkindex integer, created timestamp without time zone);`

func TestCommandInit(t *testing.T) {
	s := NewTestSegment()
	s.gpuEnabled = true

	AssertCommand(t, s, commands.CommandInit)
	AssertValue(t, s, "0", types.NewFTIntegerArray(0))
	AssertCommandResponse(t, s, commands.CommandInit, []string{}, "0   gpu")
}

func TestCommandInit_NoGPU(t *testing.T) {
	s := NewTestSegment()
	s.gpuEnabled = false

	AssertCommand(t, s, commands.CommandInit)
	AssertValue(t, s, "0", types.NewFTIntegerArray(0))
	AssertCommandResponse(t, s, commands.CommandInit, []string{}, "0   no_gpu")
}

func TestCommandDel(t *testing.T) {
	s := NewTestSegment()
	AssertCommand(t, s, commands.CommandNewilist, "1")

	AssertVariable[*types.FTIntegerArray](t, s, "1")

	AssertCommand(t, s, commands.CommandDel, "1")
	AssertNoVariable(t, s, "1")
}

func TestCommandClearVariableStore(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1")
	AssertVariable[*types.FTIntegerArray](t, s, "1")

	AssertCommand(t, s, commands.CommandNewflist, "2")
	AssertVariable[*types.FTFloatArray](t, s, "2")

	AssertCommand(t, s, commands.CommandNewilist, "3", "3")
	AssertCommand(t, s, commands.CommandNewArray, "4", "b12", "3")

	AssertValue(t, s, "4", types.NewFTBytearrayArrayOrPanic(
		12,
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	))

	AssertCommand(t, s, commands.CommandClearVariableStore)

	count, _, _ := s.variables.Stats()
	if count != 0 {
		t.Fatal("Variable store not cleared.")
	}
}

func TestCommandNewilist(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1")
	AssertVariable[*types.FTIntegerArray](t, s, "1")
}

func TestCommandNewflist(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1")
	AssertVariable[*types.FTFloatArray](t, s, "1")
}

func TestCommandArange(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10")
	AssertCommand(t, s, commands.CommandArange, "2", "1")
	AssertValue(t, s, "2", types.NewFTIntegerArray(0, 1, 2, 3, 4, 5, 6, 7, 8, 9))
}
func TestCommandNewArray_Integer_NoValue(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10")
	AssertCommand(t, s, commands.CommandNewArray, "2", "i", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
}
func TestCommandNewArray_Integer_Value(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10")
	AssertCommand(t, s, commands.CommandNewilist, "2", "42")
	AssertCommand(t, s, commands.CommandNewArray, "3", "i", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(42, 42, 42, 42, 42, 42, 42, 42, 42, 42))
}
func TestCommandNewArray_Float_NoValue(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10")
	AssertCommand(t, s, commands.CommandNewArray, "2", "f", "1")

	AssertValue(t, s, "2", types.NewFTFloatArray(0, 0, 0, 0, 0, 0, 0, 0, 0, 0))
}
func TestCommandNewArray_Float_Value(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10")
	AssertCommand(t, s, commands.CommandNewflist, "2", "42.16")
	AssertCommand(t, s, commands.CommandNewArray, "3", "f", "1", "2")

	AssertValue(t, s, "3", types.NewFTFloatArray(42.16, 42.16, 42.16, 42.16, 42.16, 42.16, 42.16, 42.16, 42.16, 42.16))
}
func TestCommandNewArray_Bytearray(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3")
	AssertCommand(t, s, commands.CommandNewArray, "2", "b12", "1")

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		12,
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	))
}

func TestCommandNewArray_Ed25519Int_NoValue(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3")
	AssertCommand(t, s, commands.CommandNewArray, "2", "I", "1")

	AssertValue(t, s, "2", types.NewFTEd25519IntArrayFromInt64s(0, 0, 0))
}

func TestCommandNewArray_Ed25519Int_Value(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "5")
	AssertCommand(t, s, commands.CommandNewArray, "3", "I", "1", "2")

	AssertValue(t, s, "3", types.NewFTEd25519IntArrayFromInt64s(5, 5, 5))
}

func TestCommandNewArray_Ed25519_NoValue(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3")
	AssertCommand(t, s, commands.CommandNewArray, "2", "E", "1")

	AssertValue(t, s, "2", types.NewEd25519ArrayFromInt64sOrPanic(0, 0, 0))
}
func TestCommandNewArray_Ed25519_Value(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3")
	AssertCommand(t, s, commands.CommandNewilist, "2", "5")
	AssertCommand(t, s, commands.CommandNewArray, "3", "E", "1", "2")

	AssertValue(t, s, "3", types.NewEd25519ArrayFromInt64sOrPanic(5, 5, 5))
}

func TestCommandNewArray_Ed25519_Empty(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "0")
	AssertCommand(t, s, commands.CommandNewArray, "2", "E", "1")

	xs := AssertVariable[*types.Ed25519Array](t, s, "2")
	if !xs.IsEmpty() {
		t.Error("array is not empty")
	}
}

func TestCommandSetItem_Integer_NoKeys(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(5, 4, 6))

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 3, 4, 5))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2")
	AssertValue(t, s, "1", types.NewFTIntegerArray(5, 4, 6))
}
func TestCommandSetItem_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(5, 6))
	s.Variables().Set("3", types.NewFTIntegerArray(1, 3))

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 3, 4, 5))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 5, 3, 6, 5))
}
func TestCommandSetItem_Float_NoKeys(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTFloatArray(5.3, 4.2, 6.1))

	AssertValue(t, s, "1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2")
	AssertValue(t, s, "1", types.NewFTFloatArray(5.3, 4.2, 6.1))
}
func TestCommandSetItem_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTFloatArray(5.3, 4.2))
	s.Variables().Set("3", types.NewFTIntegerArray(1, 3))

	AssertValue(t, s, "1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTFloatArray(1.1, 5.3, 3.3, 4.2, 5.5))
}
func TestCommandSetItem_Bytearray_NoKeys(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 1, 1},
		[]byte{1, 0, 0, 0},
		[]byte{1, 0, 0, 1},
	))
	s.Variables().Set("2", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 1, 0},
		[]byte{1, 0, 1, 0},
	))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2")
	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 1, 0},
		[]byte{1, 0, 1, 0},
	))
}
func TestCommandSetItem_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 1, 1},
		[]byte{1, 0, 0, 0},
		[]byte{1, 0, 0, 1},
	))
	s.Variables().Set("2", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 1, 0},
		[]byte{1, 0, 1, 0},
	))
	s.Variables().Set("3", types.NewFTIntegerArray(1, 2))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 1, 1},
		[]byte{0, 1, 1, 0},
		[]byte{1, 0, 1, 0},
	))
}

func TestCommandSetItem_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(10, 20))
	s.Variables().Set("3", types.NewFTIntegerArray(1, 3))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2", "3")

	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(1, 10, 3, 20, 5))
}

func TestCommandSetItem_Ed25519_NoKeys(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(10, 20))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2")

	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(10, 20))
}

func TestCommandSetItem_Ed25519_EmptyArray_NoKeys(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic())
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(10, 20))

	AssertCommand(t, s, commands.CommandSetItem, "1", "2")

	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(10, 20))
}

func TestCommandDelItem_IntegerArray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(3))

	AssertCommand(t, s, commands.CommandDelItem, "1", "2")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 3, 5))
}

func TestCommandDelItem_MultipleElements_IntegerArray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
	s.Variables().Set("2", types.NewFTIntegerArray(3, 7))

	AssertCommand(t, s, commands.CommandDelItem, "1", "2")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 3, 5, 6, 7, 9, 10))
}

func TestCommandDelItem_FloatArray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01))
	s.Variables().Set("2", types.NewFTIntegerArray(3, 7))

	AssertCommand(t, s, commands.CommandDelItem, "1", "2")

	AssertValue(t, s, "1", types.NewFTFloatArray(1.1, 2.2, 3.3, 5.5, 6.6, 7.7, 9.9, 10.01))
}

func TestCommandDelItem_BytearrayArray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 0, 0, 1},
		[]byte{0, 0, 1, 0},
		[]byte{0, 0, 1, 1},
		[]byte{0, 1, 0, 0},
		[]byte{0, 1, 0, 1},
		[]byte{0, 1, 1, 0},
		[]byte{0, 1, 1, 1},
		[]byte{1, 0, 0, 0},
		[]byte{1, 0, 0, 1},
		[]byte{1, 0, 1, 0},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(3, 7))

	AssertCommand(t, s, commands.CommandDelItem, "1", "2")

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 0, 0, 1},
		[]byte{0, 0, 1, 0},
		[]byte{0, 0, 1, 1},
		[]byte{0, 1, 0, 1},
		[]byte{0, 1, 1, 0},
		[]byte{0, 1, 1, 1},
		[]byte{1, 0, 0, 1},
		[]byte{1, 0, 1, 0},
	))
}

func TestCommandDelItem_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4))
	s.Variables().Set("2", types.NewFTIntegerArray(1))

	AssertCommand(t, s, commands.CommandDelItem, "1", "2")

	AssertValue(t, s, "1", types.NewFTEd25519IntArrayFromInt64s(1, 3, 4))
}

func TestCommandDelItem_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4))
	s.Variables().Set("2", types.NewFTIntegerArray(1))

	AssertCommand(t, s, commands.CommandDelItem, "1", "2")

	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(1, 3, 4))
}

func TestCommandLen_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	AssertCommand(t, s, commands.CommandLen, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(5))
}
func TestCommandLen_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.5, 2.2, 3.6, 4.3, 5.1))
	AssertCommand(t, s, commands.CommandLen, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(5))
}
func TestCommandLen_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		1,
		[]byte{1},
		[]byte{2},
		[]byte{3},
		[]byte{4},
		[]byte{5},
	))
	AssertCommand(t, s, commands.CommandLen, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(5))
}
func TestCommandLen_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	AssertCommand(t, s, commands.CommandLen, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(5))
}
func TestCommandLen_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	AssertCommand(t, s, commands.CommandLen, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(5))
}

func TestCommandListmapContains(t *testing.T) {
	s := NewTestSegment()

	//Last key will be ignored.
	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 1))
	s.Variables().Set("2", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 1.1))
	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")

	s.Variables().Set("4", types.NewFTIntegerArray(2, 17, 3))
	s.Variables().Set("5", types.NewFTFloatArray(2.2, 3.1415, 3.3))

	AssertCommandResponse(t, s, commands.CommandListmapContains, []string{"6", "3", "4", "5"}, "array i 6")
	AssertValue(t, s, "6", types.NewFTIntegerArray(1, 0, 1))
}

func TestCommandListmapGetkeys(t *testing.T) {
	s := NewTestSegment()

	//Last key will be ignored.
	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 1))
	s.Variables().Set("2", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 1.1))
	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	expectedResponse := "array i 4 array f 5"
	AssertCommandResponse(t, s, commands.CommandListmapKeys, []string{"4", "5", "3"}, expectedResponse)
	AssertValue(t, s, "4", types.NewFTIntegerArray(1, 2, 3, 4))
	AssertValue(t, s, "5", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4))
}

func TestCommandListmapAddItem_IgnoreError(t *testing.T) {
	s := NewTestSegment()

	//Last key will be ignored - duplicate.
	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 1))
	s.Variables().Set("2", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 1.1))
	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	s.Variables().Set("4", types.NewFTIntegerArray(2, 17, 3, 13))
	s.Variables().Set("5", types.NewFTFloatArray(2.2, 3.1415, 3.3, 1.41421356))
	AssertCommandResponse(t, s, commands.CommandListmapAddItem, []string{"2", "3", "1", "4_5"}, "  array i 2")
	ints := types.NewFTIntegerArray(1, 2, 3, 4, 17, 13)
	floats := types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 3.1415, 1.41421356)

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}
	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error from NewListMapFromArrays")
	AssertValue(t, s, "3", lm)
}

func TestCommandListmapRemoveItem(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	s.Variables().Set("4", types.NewFTIntegerArray(2, 3))
	s.Variables().Set("5", types.NewFTFloatArray(2.2, 3.3))
	AssertCommandResponse(t, s, commands.CommandListmapRemoveItem, []string{"6", "7", "8", "9", "3", "1", "4_5"}, "array i 6 array f 7 array i 8 array i 9")
	ints := types.NewFTIntegerArray(1, 4, 5)
	floats := types.NewFTFloatArray(1.1, 4.4, 5.5)

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}
	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error from NewListMapFromArrays")
	AssertValue(t, s, "3", lm)
	AssertValue(t, s, "6", types.NewFTIntegerArray(4, 5))
	AssertValue(t, s, "7", types.NewFTFloatArray(4.4, 5.5))
	AssertValue(t, s, "8", types.NewFTIntegerArray(3, 4))
	AssertValue(t, s, "9", types.NewFTIntegerArray(1, 2))
}

func TestCommandListmapRemoveItem_Duplicates(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4))
	s.Variables().Set("2", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4))
	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	s.Variables().Set("4", types.NewFTIntegerArray(2, 3))
	s.Variables().Set("5", types.NewFTFloatArray(2.2, 3.3))
	AssertCommandResponse(t, s, commands.CommandListmapRemoveItem, []string{"6", "7", "8", "9", "3", "0", "4_5"}, "array i 6 array f 7 array i 8 array i 9")
	ints := types.NewFTIntegerArray(1, 4)
	floats := types.NewFTFloatArray(1.1, 4.4)

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}
	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error from NewListMapFromArrays")
	AssertValue(t, s, "3", lm)
	AssertValue(t, s, "6", types.NewFTIntegerArray(4))
	AssertValue(t, s, "7", types.NewFTFloatArray(4.4))
	AssertValue(t, s, "8", types.NewFTIntegerArray(3))
	AssertValue(t, s, "9", types.NewFTIntegerArray(1))
}

func TestCommandListmapRemoveItem_Duplicates2(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4))
	s.Variables().Set("2", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4))
	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	s.Variables().Set("4", types.NewFTIntegerArray(2, 3))
	s.Variables().Set("5", types.NewFTFloatArray(2.2, 3.3))
	AssertCommandResponse(t, s, commands.CommandListmapRemoveItem, []string{"6", "7", "8", "9", "3", "0", "4_5"}, "array i 6 array f 7 array i 8 array i 9")
	ints := types.NewFTIntegerArray(1, 4)
	floats := types.NewFTFloatArray(1.1, 4.4)

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}
	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error from NewListMapFromArrays")
	AssertValue(t, s, "3", lm)
}

func TestCommandListmapRemoveItem_RemoveAll(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4))
	s.Variables().Set("2", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4))
	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	s.Variables().Set("4", types.NewFTIntegerArray(1, 2, 3, 4))
	s.Variables().Set("5", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4))
	AssertCommandResponse(t, s, commands.CommandListmapRemoveItem, []string{"6", "7", "8", "9", "3", "0", "4_5"}, "array i 6 array f 7 array i 8 array i 9")
	ints := types.NewFTIntegerArray()
	floats := types.NewFTFloatArray()

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}
	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error from NewListMapFromArrays")
	AssertValue(t, s, "3", lm)
}

func TestCommandListmapIntersectItem(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3))
	s.Variables().Set("2", types.NewFTFloatArray(4.4, 5.5, 6.6))

	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	s.Variables().Set("4", types.NewFTIntegerArray(4, 2, 3))

	AssertCommandResponse(t, s, commands.CommandListmapIntersectItem, []string{"5", "3", "4_2"}, "listmap if 5")
	expectedInts := types.NewFTIntegerArray(2, 3)
	expectedFloats := types.NewFTFloatArray(5.5, 6.6)
	keys := []types.ArrayElementTypeVal{
		expectedInts,
		expectedFloats,
	}
	typecode := []types.TypeCode{"i", "f"}
	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error from NewListMapFromArrays")
	AssertValue(t, s, "5", lm)
}

func TestCommandListmapIntersectItem_NoIntersection(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3))
	s.Variables().Set("2", types.NewFTFloatArray(4.4, 5.5, 6.6))

	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	s.Variables().Set("4", types.NewFTIntegerArray(4, 5, 6))

	AssertCommandResponse(t, s, commands.CommandListmapIntersectItem, []string{"5", "3", "4_2"}, "listmap if 5")
	expectedInts := types.NewFTIntegerArray()
	expectedFloats := types.NewFTFloatArray()
	keys := []types.ArrayElementTypeVal{
		expectedInts,
		expectedFloats,
	}
	typecode := []types.TypeCode{"i", "f"}
	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error from NewListMapFromArrays")
	AssertValue(t, s, "5", lm)
}

func TestCommandListmapCopy(t *testing.T) {
	s := NewTestSegment()

	ints := types.NewFTIntegerArray(1, 2, 3, 4)
	floats := types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4)
	typecode := []types.TypeCode{"i", "f"}
	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}
	expectedLm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error from NewListMapFromArrays")
	s.Variables().Set("1", ints)
	s.Variables().Set("2", floats)
	tcString := "if"
	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")
	AssertCommandResponse(t, s, commands.CommandListmapCopy, []string{"4", "3"}, "listmap if 4")
	AssertValue(t, s, "4", expectedLm)
}

func TestCommandSetLength_Integer_Truncate(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(3))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 3))
}

func TestCommandSetLength_Integer_NoOp(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(5))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
}
func TestCommandSetLength_Integer_Extend(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(10))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 3, 4, 5, 0, 0, 0, 0, 0))
}
func TestCommandSetLength_Float_Truncate(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTIntegerArray(3))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTFloatArray(1.1, 2.2, 3.3))
}
func TestCommandSetLength_Float_NoOp(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTIntegerArray(5))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
}
func TestCommandSetLength_Float_Extend(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTIntegerArray(10))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 0, 0, 0, 0, 0))
}
func TestCommandSetLength_Ed25519Int_Truncate(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(3))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
}
func TestCommandSetLength_Ed25519Int_NoOp(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(5))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
}
func TestCommandSetLength_Ed25519Int_Extend(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(10))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5, 0, 0, 0, 0, 0))
}

func TestCommandSetLength_Ed25519_Truncate(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(3))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3))
}

func TestCommandSetLength_Ed25519_NoOp(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(5))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
}

func TestCommandSetLength_Ed25519_Extend(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(10))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5, 0, 0, 0, 0, 0))
}

func TestCommandSetLength_Ed25519_Empty_Extend(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic())

	s.Variables().Set("2", types.NewFTIntegerArray(10))
	AssertCommand(t, s, commands.CommandSetLength, "1", "2")
	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(0, 0, 0, 0, 0, 0, 0, 0, 0, 0))

	s.Variables().Set("5", types.NewEd25519ArrayFromInt64sOrPanic())
	s.Variables().Set("6", types.NewFTIntegerArray(0))
	AssertCommand(t, s, commands.CommandSetLength, "5", "6")

	s.Variables().Set("7", types.NewEd25519ArrayFromInt64sOrPanic())
	AssertCommand(t, s, commands.CommandEq, "8", "5", "7")
	AssertValue(t, s, "8", types.NewFTIntegerArray())

	s.Variables().Set("9", types.NewEd25519ArrayFromInt64sOrPanic(1))
	AssertEd25519ArrayNotEmpty(t, s, "9")

	AssertCommand(t, s, commands.CommandSetLength, "9", "6")
	AssertEd25519ArrayEmpty(t, s, "9")
}

func TestCommandSetLength_Bytearray_Truncate(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(3))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
	))
}
func TestCommandSetLength_Bytearray_NoOp(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(5))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
}
func TestCommandSetLength_Bytearray_Extend(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(10))

	AssertCommand(t, s, commands.CommandSetLength, "1", "2")

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{0, 0, 0, 0},
		[]byte{0, 0, 0, 0},
		[]byte{0, 0, 0, 0},
		[]byte{0, 0, 0, 0},
		[]byte{0, 0, 0, 0},
	))
}
func TestCommandGetItem_NegativeKey_Error(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, -1))

	AssertCommandFailure(t, s, commands.CommandGetItem, []string{"3", "1", "2"}, "out of range")
}
func TestCommandGetItem_Ed25519_NegativeKey_Error(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, -1))

	AssertCommandFailure(t, s, commands.CommandGetItem, []string{"3", "1", "2"}, "out of range")
}
func TestCommandGetItem_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3))

	AssertCommand(t, s, commands.CommandGetItem, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(2, 4))
}
func TestCommandGetItem_Integer_NoKeys(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))

	AssertCommand(t, s, commands.CommandGetItem, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 2, 3, 4, 5))
}
func TestCommandGetItem_Integer_OutOfRange(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 6))

	AssertCommandFailure(t, s, commands.CommandGetItem, []string{"3", "1", "2"}, "out of range")
}
func TestCommandGetItem_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3))

	AssertCommand(t, s, commands.CommandGetItem, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTFloatArray(2.2, 4.4))
}
func TestCommandGetItem_Float_NoKeys(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))

	AssertCommand(t, s, commands.CommandGetItem, "2", "1")

	AssertValue(t, s, "2", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
}
func TestCommandGetItem_Float_OutOfRange(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 6))

	AssertCommandFailure(t, s, commands.CommandGetItem, []string{"3", "1", "2"}, "out of range")
}
func TestCommandGetItem_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3))

	AssertCommand(t, s, commands.CommandGetItem, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{4, 5, 6, 7},
		[]byte{12, 13, 14, 15},
	))
}
func TestCommandGetItem_Bytearray_NoKeys(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))

	AssertCommand(t, s, commands.CommandGetItem, "2", "1")

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
}
func TestCommandGetItem_Bytearray_OutOfRange(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 6))

	AssertCommandFailure(t, s, commands.CommandGetItem, []string{"3", "1", "2"}, "out of range")
}

func TestCommandGetItem_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))

	s.Variables().Set("2", types.NewFTIntegerArray(1, 3))
	AssertCommand(t, s, commands.CommandGetItem, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTEd25519IntArrayFromInt64s(2, 4))
}
func TestCommandGetItem_Ed25519Int_NoKeys(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	AssertCommand(t, s, commands.CommandGetItem, "2", "1")

	AssertValue(t, s, "2", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
}
func TestCommandGetItem_Ed25519Int_OutOfRange(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 6))

	AssertCommandFailure(t, s, commands.CommandGetItem, []string{"3", "1", "2"}, "out of range")
}

func TestCommandGetItem_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))

	s.Variables().Set("0", types.NewFTIntegerArray(1, 3))
	AssertCommand(t, s, commands.CommandGetItem, "2", "1", "0")

	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic(2, 4))
	AssertCommand(t, s, commands.CommandEq, "4", "2", "3")

	AssertValue(t, s, "4", types.NewFTIntegerArray(1, 1))
}
func TestCommandGetItem_Ed25519_NoKeys(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))

	AssertCommand(t, s, commands.CommandGetItem, "2", "1")

	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	AssertCommand(t, s, commands.CommandEq, "4", "2", "3")

	AssertValue(t, s, "4", types.NewFTIntegerArray(1, 1, 1, 1, 1))
}
func TestCommandGetItem_Ed25519_OutOfRange(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 6))

	AssertCommandFailure(t, s, commands.CommandGetItem, []string{"3", "1", "2"}, "out of range")
}

func TestCommandLookup_Integer_NoDefault(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))

	AssertCommand(t, s, commands.CommandLookup, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(2, 4, 0))
}
func TestCommandLookup_Integer_WithDefault(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))
	s.Variables().Set("3", types.NewFTIntegerArray(42))

	AssertCommand(t, s, commands.CommandLookup, "4", "1", "2", "3")

	AssertValue(t, s, "4", types.NewFTIntegerArray(2, 4, 42))
}
func TestCommandLookup_Float_NoDefault(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))

	AssertCommand(t, s, commands.CommandLookup, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTFloatArray(2.2, 4.4, 0))
}
func TestCommandLookup_Float_WithDefault(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))
	s.Variables().Set("3", types.NewFTFloatArray(42.2))

	AssertCommand(t, s, commands.CommandLookup, "4", "1", "2", "3")

	AssertValue(t, s, "4", types.NewFTFloatArray(2.2, 4.4, 42.2))
}
func TestCommandLookup_Bytearray_NoDefault(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))

	AssertCommand(t, s, commands.CommandLookup, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{4, 5, 6, 7},
		[]byte{12, 13, 14, 15},
		[]byte{0, 0, 0, 0},
	))
}
func TestCommandLookup_Bytearray_WithDefault(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))
	s.Variables().Set("3", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{42, 43, 44, 45},
	))

	AssertCommand(t, s, commands.CommandLookup, "4", "1", "2", "3")

	AssertValue(t, s, "4", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{4, 5, 6, 7},
		[]byte{12, 13, 14, 15},
		[]byte{42, 43, 44, 45},
	))
}

func TestCommandLookup_Ed25519Int_NoDefault(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))

	AssertCommand(t, s, commands.CommandLookup, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTEd25519IntArrayFromInt64s(2, 4, 0))
}
func TestCommandLookup_Ed25519Int_WithDefault(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))
	s.Variables().Set("3", types.NewFTEd25519IntArrayFromInt64s(42))

	AssertCommand(t, s, commands.CommandLookup, "4", "1", "2", "3")
	AssertValue(t, s, "4", types.NewFTEd25519IntArrayFromInt64s(2, 4, 42))
}

func TestCommandLookup_Ed25519_NoDefault(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))

	AssertCommand(t, s, commands.CommandLookup, "3", "1", "2")
	AssertValue(t, s, "3", types.NewEd25519ArrayFromInt64sOrPanic(2, 4, 0))
}
func TestCommandLookup_Ed25519_WithDefault(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 3, 6))
	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic(42))

	AssertCommand(t, s, commands.CommandLookup, "4", "1", "2", "3")
	AssertValue(t, s, "4", types.NewEd25519ArrayFromInt64sOrPanic(2, 4, 42))
}

func TestCommandLookup_Ed25519_AllOutOfRange(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(7, 8, 9))
	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic(42))

	AssertCommand(t, s, commands.CommandLookup, "4", "1", "2", "3")
	AssertValue(t, s, "4", types.NewEd25519ArrayFromInt64sOrPanic(42, 42, 42))
}

func TestCommandIndex_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(0, 2, 1, 0, 0, 3, 0, 4, 5, 0))

	AssertCommand(t, s, commands.CommandIndex, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 2, 5, 7, 8))
}
func TestCommandIndex_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(0, 2.1, 1.1, 0, 1, 0, 0, 4.5, 5.1, 0))

	AssertCommand(t, s, commands.CommandIndex, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 2, 4, 7, 8))
}
func TestCommandIndex_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 0, 4, 0, 6, 7))

	AssertCommand(t, s, commands.CommandIndex, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(0, 1, 3, 5, 6))
}
func TestCommandIndex_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 0, 4, 0, 6, 7))

	AssertCommand(t, s, commands.CommandIndex, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(0, 1, 3, 5, 6))
}
func TestCommandIndex_Ed25519_Empty(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic())

	AssertCommand(t, s, commands.CommandIndex, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray())
}

func TestCommandIndex_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		5,
		[]byte{0, 1, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0},
		[]byte{1, 2, 3, 4, 5},
		[]byte{0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 1},
	))

	AssertCommand(t, s, commands.CommandIndex, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(0, 2, 4))
}

func TestCommandConcat(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTBytearrayArrayOrPanic(
		2,
		[]byte{4, 5},
		[]byte{8, 9},
		[]byte{12, 13},
		[]byte{16, 17},
		[]byte{20, 21},
	))

	AssertCommandResponse(t, s, commands.CommandConcat, []string{"3", "1", "2"}, "array b6 3")

	AssertValue(t, s, "3", types.NewFTBytearrayArrayOrPanic(
		6,
		[]byte{0, 1, 2, 3, 4, 5},
		[]byte{4, 5, 6, 7, 8, 9},
		[]byte{8, 9, 10, 11, 12, 13},
		[]byte{12, 13, 14, 15, 16, 17},
		[]byte{16, 17, 18, 19, 20, 21},
	))
}

func TestCommandConcat2(t *testing.T) {
	s := NewTestSegment()

	xs, err := types.NewFTIntegerArray(1, 2, 3, 4, 5).AsType("b8")
	assert.NoError(t, err, "error from AsType")
	ys, err := types.NewFTIntegerArray(1, 1, 1, 1, 1).AsType("b8")
	assert.NoError(t, err, "error from AsType")

	s.Variables().Set("1", xs)
	s.Variables().Set("2", ys)
	AssertCommandResponse(t, s, commands.CommandConcat, []string{"3", "1", "2"}, "array b16 3")

	zs, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), "3")
	assert.NoError(t, err, "error from GetAs")

	assert.Equal(t, int64(16), zs.Width(), "width not expected value")
	assert.Equal(t, int64(5), zs.Length(), "length not expected value")
}

func TestCommandByteproject_CompleteMap(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))

	// 1 -> 0
	// 0 -> 2
	// 2 -> 1
	s.Variables().Set("2", types.NewFTIntegerArray(0, 2, 1))
	s.Variables().Set("3", types.NewFTIntegerArray(1, 0, 2))

	AssertCommandResponse(t, s, commands.CommandByteProject, []string{"4", "1", "3", "2", "3"}, "array b3 4")

	AssertValue(t, s, "4", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{1, 2, 0},
		[]byte{5, 6, 4},
		[]byte{9, 10, 8},
		[]byte{13, 14, 12},
		[]byte{17, 18, 16},
	))
}
func TestCommandByteproject_PartialMap(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))

	// 1 -> 0
	// 0 -> 2
	s.Variables().Set("2", types.NewFTIntegerArray(0, 2))
	s.Variables().Set("3", types.NewFTIntegerArray(1, 0))

	AssertCommandResponse(t, s, commands.CommandByteProject, []string{"4", "1", "3", "2", "3"}, "array b3 4")

	AssertValue(t, s, "4", types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{1, 0, 0},
		[]byte{5, 0, 4},
		[]byte{9, 0, 8},
		[]byte{13, 0, 12},
		[]byte{17, 0, 16},
	))
}

func TestCommandSerialiseAndDeserialise_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(0, 2, 1, 0, 0, 3, 0, 4, 5, 0))
	s.Variables().Set("3", types.NewFTIntegerArray())

	AssertCommand(t, s, commands.CommandSerialise, "2", "1")
	AssertCommand(t, s, commands.CommandDeserialise, "3", "2")

	AssertVariable[*types.FTBytearrayArray](t, s, "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 2, 1, 0, 0, 3, 0, 4, 5, 0))
}

func TestCommandSerialiseAndDeserialise_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(0, 2.1, 1.1, 0, 1, 0, 0, 4.5, 5.1, 0))
	s.Variables().Set("3", types.NewFTFloatArray())

	AssertCommand(t, s, commands.CommandSerialise, "2", "1")
	AssertCommand(t, s, commands.CommandDeserialise, "3", "2")

	AssertVariable[*types.FTBytearrayArray](t, s, "2")
	AssertValue(t, s, "3", types.NewFTFloatArray(0, 2.1, 1.1, 0, 1, 0, 0, 4.5, 5.1, 0))
}

func TestCommandSerialiseAndDeserialise_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(0, 2, 1, 0, 0, 3, 0, 4, 5, 0))
	s.Variables().Set("3", types.NewFTEd25519IntArrayFromInt64s())

	AssertCommand(t, s, commands.CommandSerialise, "2", "1")
	AssertCommand(t, s, commands.CommandDeserialise, "3", "2")

	AssertVariable[*types.FTBytearrayArray](t, s, "2")
	AssertValue(t, s, "3", types.NewFTEd25519IntArrayFromInt64s(0, 2, 1, 0, 0, 3, 0, 4, 5, 0))
}

func TestCommandSerialiseAndDeserialise_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(0, 2, 1, 0, 0, 3, 0, 4, 5, 0))
	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic())

	AssertCommand(t, s, commands.CommandSerialise, "2", "1")
	AssertCommand(t, s, commands.CommandDeserialise, "3", "2")

	AssertVariable[*types.FTBytearrayArray](t, s, "2")
	AssertValue(t, s, "3", types.NewEd25519ArrayFromInt64sOrPanic(0, 2, 1, 0, 0, 3, 0, 4, 5, 0))
}

func TestCommandReduceSum_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTIntegerArray(5, 11, 2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceSum, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 16, 2, 5, 6, 7))
}
func TestCommandReduceSum_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0))
	s.Variables().Set("2", types.NewFTFloatArray(5.5, 12.4, 2.2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceSum, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTFloatArray(1, 2, 17.9, 2.2, 5, 6, 7))
}
func TestCommandReduceSum_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(5, 11, 2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceSum, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 16, 2, 5, 6, 7))
}

func TestCommandReduceSum_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(5, 8, 1, 2, 50))

	s.Variables().Set("2", types.NewFTIntegerArray(2, 2, 3, 3))
	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic(10, 5, 20, 5))

	AssertCommand(t, s, commands.CommandReduceSum, "1", "3", "2")

	s.Variables().Set("4", types.NewEd25519ArrayFromInt64sOrPanic(5, 8, 15, 25, 50))

	AssertCommand(t, s, commands.CommandEq, "5", "1", "4")

	AssertValue(t, s, "5", types.NewFTIntegerArray(1, 1, 1, 1, 1))
}

func TestCommandReduceSum_Ed25519_Issue96(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(0, 0, 0, 0, 0, 0, 0, 0, 0, 0))

	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(0, 1, 0, 0))
	s.Variables().Set("3", types.NewFTIntegerArray(3, 7, 8, 5))

	AssertCommand(t, s, commands.CommandReduceISum, "1", "2", "3")

	s.Variables().Set("4", types.NewEd25519ArrayFromInt64sOrPanic(0, 0, 0, 0, 0, 0, 0, 1, 0, 0))
	AssertCommand(t, s, commands.CommandEq, "5", "1", "4")

	xs, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), "5")
	if err != nil {
		t.Error(err)
	}
	for i, v := range xs.Values() {
		if v != 1 {
			t.Errorf("arrays not equal, different value at index %v", i)
		}
	}
}

func TestCommandReduceISum_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(5, 8, 30, 20, 50))

	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(10, 20, 30, 5))

	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3, 3))

	AssertCommand(t, s, commands.CommandReduceISum, "1", "2", "3")

	s.Variables().Set("4", types.NewEd25519ArrayFromInt64sOrPanic(5, 8, 60, 55, 50))

	AssertCommand(t, s, commands.CommandEq, "5", "1", "4")

	AssertValue(t, s, "5", types.NewFTIntegerArray(1, 1, 1, 1, 1))

}

func TestCommandContains_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTIntegerArray(3, 5, 8, 2))

	AssertCommand(t, s, commands.CommandContains, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 1, 0, 1))
}
func TestCommandContains_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7))
	s.Variables().Set("2", types.NewFTFloatArray(3.3, 5.5, 7, 2.2))

	AssertCommand(t, s, commands.CommandContains, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 1, 0, 1))
}

func TestCommandContains_Bytearray(t *testing.T) {
	s := NewTestSegment()
	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
	s.Variables().Set("2", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{0, 1, 2, 3},
		[]byte{16, 17, 11, 19},
	))

	AssertCommand(t, s, commands.CommandContains, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 1, 1, 0))
}

func TestCommandContains_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(3, 5, 8, 2))

	AssertCommand(t, s, commands.CommandContains, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 1, 0, 1))
}
func TestCommandContains_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(3, 5, 8, 2))

	AssertCommand(t, s, commands.CommandContains, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(1, 1, 0, 1))
}

func TestCommandContains_Ed25519_Empty(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic())
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(3, 5, 8, 2))

	AssertCommand(t, s, commands.CommandContains, "3", "1", "2")
	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 0, 0, 0))
}

func TestCommandCumSum_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5, 6, 7))

	AssertCommand(t, s, commands.CommandCumSum, "2", "1")
	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 3, 6, 10, 15, 21, 28))
}
func TestCommandCumSum_Float64(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 2.1, 3.3, 4.4, 5.5, 6.6, 7.7))

	AssertCommand(t, s, commands.CommandCumSum, "2", "1")
	AssertValue(t, s, "2", types.NewFTFloatArray(1.1, 3.2, 6.5, 10.9, 16.4, 23, 30.7))
}
func TestCommandCumsum_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{1, 10, 1, 1},
		[]byte{1, 11, 2, 2},
		[]byte{0, 15, 3, 4},
		[]byte{1, 17, 4, 8},
	))

	AssertCommand(t, s, commands.CommandCumSum, "2", "1")
	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{1, 10, 1, 1},
		[]byte{1, 11, 3, 3},
		[]byte{1, 15, 3, 7},
		[]byte{1, 31, 7, 15},
	))
}
func TestCommandCumSum_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5, 6, 7))

	AssertCommand(t, s, commands.CommandCumSum, "2", "1")
	AssertValue(t, s, "2", types.NewFTEd25519IntArrayFromInt64s(1, 3, 6, 10, 15, 21, 28))
}
func TestCommandCumSum_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5, 6, 7))

	AssertCommand(t, s, commands.CommandCumSum, "2", "1")
	AssertValue(t, s, "2", types.NewEd25519ArrayFromInt64sOrPanic(1, 3, 6, 10, 15, 21, 28))
}

func TestCommandCumsum_Ed25519_Empty(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic())

	AssertCommand(t, s, commands.CommandCumSum, "2", "1")
	AssertValue(t, s, "2", types.NewEd25519ArrayFromInt64sOrPanic())
}

func TestCommandMux_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 0, 1, 2, 0, 0, 3))        // Cond
	s.Variables().Set("2", types.NewFTIntegerArray(10, 20, 30, 40, 50, 60, 70)) // if true
	s.Variables().Set("3", types.NewFTIntegerArray(11, 22, 33, 44, 55, 66, 77)) // if false

	AssertCommand(t, s, commands.CommandMux, "4", "1", "2", "3")
	AssertValue(t, s, "4", types.NewFTIntegerArray(10, 22, 30, 40, 55, 66, 70))
}
func TestCommandMux_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 0, 1, 2, 0, 0, 3))             // Cond
	s.Variables().Set("2", types.NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)) // if true
	s.Variables().Set("3", types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7)) // if false

	AssertCommand(t, s, commands.CommandMux, "4", "1", "2", "3")
	AssertValue(t, s, "4", types.NewFTFloatArray(1.0, 2.2, 3.0, 4.0, 5.5, 6.6, 7.0))
}

func TestCommandMux_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 0, 1, 2, 0, 0, 3))                     // Cond
	s.Variables().Set("2", types.NewFTEd25519IntArrayFromInt64s(10, 20, 30, 40, 50, 60, 70)) // if true
	s.Variables().Set("3", types.NewFTEd25519IntArrayFromInt64s(11, 22, 33, 44, 55, 66, 77)) // if false

	AssertCommand(t, s, commands.CommandMux, "4", "1", "2", "3")
	AssertValue(t, s, "4", types.NewFTEd25519IntArrayFromInt64s(10, 22, 30, 40, 55, 66, 70))
}
func TestCommandMux_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 0, 1, 2, 0, 0, 3))                       // Cond
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic(10, 20, 30, 40, 50, 60, 70)) // if true
	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic(11, 22, 33, 44, 55, 66, 77)) // if false

	AssertCommand(t, s, commands.CommandMux, "4", "1", "2", "3")

	AssertValue(t, s, "4", types.NewEd25519ArrayFromInt64sOrPanic(10, 22, 30, 40, 55, 66, 70))
}

func TestCommandMux_Ed25519_Empty(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray()) // Cond
	s.Variables().Set("2", types.NewEd25519ArrayFromInt64sOrPanic())
	s.Variables().Set("3", types.NewEd25519ArrayFromInt64sOrPanic())

	AssertCommand(t, s, commands.CommandMux, "4", "1", "2", "3")

	AssertValue(t, s, "4", types.NewEd25519ArrayFromInt64sOrPanic())
}

func TestCommandMux_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 0, 1, 2)) // Cond
	s.Variables().Set("2", types.NewFTBytearrayArrayOrPanic(    // if true
		4,
		[]byte{12, 0, 7, 5},
		[]byte{13, 1, 6, 4},
		[]byte{4, 5, 7, 9},
		[]byte{15, 7, 4, 3},
	))
	s.Variables().Set("3", types.NewFTBytearrayArrayOrPanic( // if false
		4,
		[]byte{1, 10, 1, 1},
		[]byte{1, 11, 2, 2},
		[]byte{0, 15, 3, 4},
		[]byte{1, 17, 4, 8},
	))

	AssertCommand(t, s, commands.CommandMux, "4", "1", "2", "3")

	AssertValue(t, s, "4", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{12, 0, 7, 5},
		[]byte{1, 11, 2, 2},
		[]byte{4, 5, 7, 9},
		[]byte{15, 7, 4, 3},
	))
}

func TestCommandReduceSum_ByteArrayArray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))

	s.Variables().Set("2", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{3, 5, 7, 6},
		[]byte{1, 1, 2, 3},
		[]byte{4, 4, 4, 4},
	))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceSum, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{3, 5, 7, 7},
		[]byte{4, 4, 4, 4},
		[]byte{16, 17, 18, 19},
	))
}

func TestCommandReduceISum_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTIntegerArray(5, 11, 2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceISum, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 19, 6, 5, 6, 7))
}
func TestCommandReduceISum_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0))
	s.Variables().Set("2", types.NewFTFloatArray(5.5, 12.4, 2.2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceISum, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTFloatArray(1, 2, 20.9, 6.2, 5, 6, 7))
}

func TestCommandReduceISum_ByteArrayArray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))

	s.Variables().Set("2", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{3, 5, 7, 6},
		[]byte{1, 1, 2, 3},
		[]byte{4, 4, 4, 4},
	))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceISum, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{11, 13, 15, 15},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	))
}

func TestCommandReduceMax_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 12, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTIntegerArray(5, 11, 2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceMax, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 11, 2, 5, 6, 7))
}

func TestCommandReduceMax_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0))
	s.Variables().Set("2", types.NewFTFloatArray(5.5, 12.4, 2.2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceMax, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTFloatArray(1, 2, 12.4, 2.2, 5, 6, 7))
}

func TestCommandReduceIMax_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 12, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTIntegerArray(5, 11, 2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceIMax, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 12, 4, 5, 6, 7))
}

func TestCommandReduceIMax_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0))
	s.Variables().Set("2", types.NewFTFloatArray(5.5, 12.4, 2.2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceIMax, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTFloatArray(1, 2, 12.4, 4.0, 5, 6, 7))
}

func TestCommandReduceMin_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTIntegerArray(5, 11, 2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceMin, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 5, 2, 5, 6, 7))
}

func TestCommandReduceMin_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0))
	s.Variables().Set("2", types.NewFTFloatArray(5.5, 12.4, 2.2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceMin, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTFloatArray(1, 2, 5.5, 2.2, 5, 6, 7))
}

func TestCommandReduceIMin_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5, 6, 7))
	s.Variables().Set("2", types.NewFTIntegerArray(5, 11, 2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceIMin, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTIntegerArray(1, 2, 3, 2, 5, 6, 7))
}

func TestCommandReduceIMin_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.0, 2.0, 3.0, 2.0, 5.0, 6.0, 7.0))
	s.Variables().Set("2", types.NewFTFloatArray(5.5, 12.4, 2.2))
	s.Variables().Set("3", types.NewFTIntegerArray(2, 2, 3))

	AssertCommand(t, s, commands.CommandReduceIMin, "1", "2", "3")

	AssertValue(t, s, "1", types.NewFTFloatArray(1, 2, 3.0, 2.0, 5, 6, 7))
}

func TestCommandSorted_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(7, 3, 2, 4, 6, 5))

	AssertCommand(t, s, commands.CommandSorted, "2", "1")

	AssertValue(t, s, "1", types.NewFTIntegerArray(7, 3, 2, 4, 6, 5))
	AssertValue(t, s, "2", types.NewFTIntegerArray(2, 3, 4, 5, 6, 7))
}

func TestCommandSorted_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(7.7, 3.3, 2.2, 4.4, 6.6, 5.5))

	AssertCommand(t, s, commands.CommandSorted, "2", "1")

	AssertValue(t, s, "1", types.NewFTFloatArray(7.7, 3.3, 2.2, 4.4, 6.6, 5.5))
	AssertValue(t, s, "2", types.NewFTFloatArray(2.2, 3.3, 4.4, 5.5, 6.6, 7.7))
}

func TestCommandIndexSorted_Integer(t *testing.T) {
	//TODO: Change this one to no index provided
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 1, 5, 6, 2))
	s.Variables().Set("2", types.NewFTIntegerArray(0, 1, 2, 3, 4))

	AssertCommand(t, s, commands.CommandIndexSorted, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 4, 2, 3))
}

func TestCommandIndexSorted_Integer2(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 1, 5, 1, 2))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 2, 3, 0, 4))

	AssertCommand(t, s, commands.CommandIndexSorted, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 2, 4, 3))
}

func TestCommandIndexSorted_Integer_NoIndexProvided(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 1, 5, 6, 2))

	AssertCommand(t, s, commands.CommandIndexSorted, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(0, 1, 4, 2, 3))
}

func TestCommandIndexSorted_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 1.1, 5.5, 6.6, 2.2))

	AssertCommand(t, s, commands.CommandIndexSorted, "2", "1")

	AssertValue(t, s, "2", types.NewFTIntegerArray(0, 1, 4, 2, 3))
}

func TestCommandIndexSorted_Float2(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.1, 1.1, 5.5, 1.1, 2.2))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 2, 3, 0, 4))

	AssertCommand(t, s, commands.CommandIndexSorted, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray(0, 1, 2, 4, 3))
}

func TestCommandVerify(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(0, 2, 1, 0, 0, 3, 0, 4, 5, 0))
	AssertCommandResponse(t, s, commands.CommandVerify, []string{"1"}, "bool 0")

	s.Variables().Set("2", types.NewFTIntegerArray(-5, 10, 11, 25, 36))
	AssertCommandResponse(t, s, commands.CommandVerify, []string{"2"}, "bool 1")

}
func TestCommandNonZero_IntegerTrue(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 1, 5, 1, 2))
	response := "bool 1"
	AssertCommandResponse(t, s, commands.CommandNonZero, []string{"1"}, response)
}

func TestCommandNonZero_IntegerPartialZeroes(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 1, 0, 0, 2))
	response := "bool 0"
	AssertCommandResponse(t, s, commands.CommandNonZero, []string{"1"}, response)
}

func TestCommandNonZero_IntegerZeroes(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(0, 0, 0, 0, 0))
	response := "bool 0"
	AssertCommandResponse(t, s, commands.CommandNonZero, []string{"1"}, response)
}

func TestCommandRandomArray_Integer(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "-20")   // min
	AssertCommand(t, s, commands.CommandNewilist, "2", "20")    // max
	AssertCommand(t, s, commands.CommandNewilist, "3", "10000") // count
	AssertCommand(t, s, commands.CommandRandomArray, "4", "i", "3", "1", "2")

	xs, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), variables.Handle("4"))
	assert.NoError(t, err, "error calling GetAs")

	dis := make([]int64, 41)
	for _, x := range xs.Values() {
		dis[int(x)+20]++
	}

	for i, x := range dis {
		if x == 0 {
			t.Fatalf("there was no %v generated", i-20)
		}
	}
}
func TestCommandRandomArray_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "-20")   // min
	AssertCommand(t, s, commands.CommandNewflist, "2", "20")    // max
	AssertCommand(t, s, commands.CommandNewilist, "3", "10000") // count
	AssertCommand(t, s, commands.CommandRandomArray, "4", "f", "3", "1", "2")

	xs, err := variables.GetAs[*types.FTFloatArray](s.Variables(), variables.Handle("4"))
	assert.NoError(t, err, "error calling GetAs")

	dis := make([]int64, 41)

	for _, x := range xs.Values() {
		dis[int(x)+20]++
	}

	// Skipping min and max exactly as the probbility that they get a hit is extremely low
	for i := 1; i < len(dis)-1; i++ {
		if dis[i] == 0 {
			t.Fatalf("there was no %v generated", i-20)
		}
	}
}
func TestCommandRandomArray_Bytearray(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10") // count
	AssertCommand(t, s, commands.CommandRandomArray, "2", "b32", "1")

	xs, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), variables.Handle("2"))
	assert.NoError(t, err, "error calling GetAs")

	for i, x := range xs.Values() {
		if len(x) != 32 {
			t.Fatalf("length of bytearray at index #%v was not 32, but %v", i, len(x))
		}

		hasNonZero := false

		for _, b := range x {
			if b != 0 {
				hasNonZero = true
				break
			}
		}

		if !hasNonZero {
			t.Fatalf("Random bytearray only had zeroes")
		}
	}
}

func TestCommandRandomArray_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "10") // count
	AssertCommand(t, s, commands.CommandRandomArray, "2", "I", "1")

	xs, err := variables.GetAs[*types.FTEd25519IntArray](s.Variables(), variables.Handle("2"))
	assert.NoError(t, err, "error calling GetAs")

	zero := edwards25519.NewScalar()

	for _, x := range xs.Values() {
		if x.Equal(zero) == 1 {
			t.Fatalf("no zeroes should be genereted when nonzero is True")
		}
	}
}

func TestCommandRandomPerm(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "20") // length
	AssertCommand(t, s, commands.CommandNewilist, "2", "40") // N
	AssertCommand(t, s, commands.CommandRandomPerm, "3", "1", "2")

	xs, err := variables.GetAs[*types.FTIntegerArray](s.Variables(), variables.Handle("3"))
	assert.NoError(t, err, "error calling GetAs")

	// TODO: add seed to keep consistent results
	assert.Equal(t, int64(20), xs.Length(), "unexpected length")
	for i, x := range xs.Values() {
		if x >= 40 {
			t.Fatalf("there was no %v generated", i)
		}
	}
}

func TestCommandAsType_Int_Float(t *testing.T) {

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "f")
	AssertValue(t, s, "2", types.NewFTFloatArray(1.0, 2.0, 3.0))
}
func TestCommandAsType_Int_Int(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "2", "3")
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "i")
	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 2, 3))
}
func TestCommandAsType_Int_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "2", "-15")
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "I")

	// L = 2**252 + 27742317777372353535851937790883648493
	// -15 ≡ 7237005577332262213973186563042994240857116359379907606001950938285454250974 (mod L)
	b, _ := big.NewInt(0).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250974", 10)

	AssertValue(t, s, "2", types.NewFTEd25519IntArray(
		types.Int64ToScalar(1),
		types.Int64ToScalar(2),
		types.BigIntToScalar(b),
	))
}
func TestCommandAsType_Int_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1")
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "E")

	es, err := types.NewEd25519ArrayFromInt64s()
	if err != nil {
		t.Fatal(err)
	}

	AssertValue(t, s, "2", es)
}
func TestCommandAsType_Int_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 234567, 41223123113, -1, -3456, -41223123113))
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "b8")

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		8,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0},
		[]byte{71, 148, 3, 0, 0, 0, 0, 0},
		[]byte{169, 240, 22, 153, 9, 0, 0, 0},
		[]byte{255, 255, 255, 255, 255, 255, 255, 255},
		[]byte{128, 242, 255, 255, 255, 255, 255, 255},
		[]byte{87, 15, 233, 102, 246, 255, 255, 255},
	))
}

func TestCommandAsType_Int_Bytearray_Long(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 234567, 41223123113, -1, -3456, -41223123113))
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "b16")

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		16,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{71, 148, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{169, 240, 22, 153, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		[]byte{128, 242, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		[]byte{87, 15, 233, 102, 246, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
	))
}

func TestCommandAsType_Ed25519Int_Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "i")
	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 2, 3))
}
func TestCommandAsType_Ed25519Int_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3))
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "E")

	AssertValue(t, s, "2", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3))
}
func TestCommandAsType_Ed25519Int_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 234567, 412231231133))
	AssertCommand(t, s, commands.CommandAsType, "2", "1", "b32")

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		32,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{71, 148, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{157, 102, 229, 250, 95, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	))
}

func TestCommandAsType_Bytearray_Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		8,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0},
		[]byte{71, 148, 3, 0, 0, 0, 0, 0},
		[]byte{157, 102, 229, 250, 95, 0, 0, 0},
	))

	AssertCommand(t, s, commands.CommandAsType, "2", "1", "i")

	AssertValue(t, s, "2", types.NewFTIntegerArray(1, 234567, 412231231133))
}
func TestCommandAsType_Bytearray_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		32,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{71, 148, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{157, 102, 229, 250, 95, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	))

	AssertCommand(t, s, commands.CommandAsType, "2", "1", "I")
	AssertValue(t, s, "2", types.NewFTEd25519IntArrayFromInt64s(1, 234567, 412231231133))
}

func TestCommandAuxDBRead_Int(t *testing.T) {

	db, err := CreateTmpTable(`CREATE TABLE test (col1 int, col2 int); 
	INSERT INTO test (col1, col2) VALUES (2,-55);
	INSERT INTO test (col1, col2) VALUES (-55,234234);`)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandAuxDbRead, "1", "2", "i", "i", `
	SELECT col1, col2 FROM test
	`)
	AssertValue(t, s, "1", types.NewFTIntegerArray(2, -55))
	AssertValue(t, s, "2", types.NewFTIntegerArray(-55, 234234))
}

func TestCommandAuxDBRead_Float(t *testing.T) {

	db, err := CreateTmpTable(`CREATE TABLE test (col1 float, col2 float); 
	INSERT INTO test (col1, col2) VALUES (5.6,-6.83);
	INSERT INTO test (col1, col2) VALUES (22222.77777,1009.766);`)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandAuxDbRead, "1", "2", "f", "f", `
	SELECT col1, col2 FROM test
	`)
	AssertValue(t, s, "1", types.NewFTFloatArray(5.6, 22222.77777))
	AssertValue(t, s, "2", types.NewFTFloatArray(-6.83, 1009.766))
}

func TestCommandAuxDBRead_Bytearray(t *testing.T) {

	db, err := CreateTmpTable(`CREATE TABLE test (col1 bytea, col2 bytea); 
	INSERT INTO test (col1, col2) VALUES ('abc','def');
	INSERT INTO test (col1, col2) VALUES ('foo001','bar002');`)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandAuxDbRead, "1", "2", "b32", "b32", `
	SELECT col1, col2 FROM test
	`)

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(32, StrToByteSlice("abc", 32), StrToByteSlice("foo001", 32)))
	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(32, StrToByteSlice("def", 32), StrToByteSlice("bar002", 32)))
}

func TestCommandAuxDBWrite_Int(t *testing.T) {
	db, err := CreateTmpTable(`CREATE TABLE test (col1 int, col2 int);`)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "-20", "10")
	AssertCommand(t, s, commands.CommandNewilist, "2", "100", "5")

	AssertCommand(t, s, commands.CommandAuxDbWrite, "test", "col1", "col2", "1", "2")

	AssertCommand(t, s, commands.CommandAuxDbRead, "3", "4", "i", "i", "SELECT col1, col2 FROM test")
	AssertValue(t, s, "3", types.NewFTIntegerArray(-20, 10))
	AssertValue(t, s, "4", types.NewFTIntegerArray(100, 5))
}

func TestCommandAuxDBWrite_IntToWrongColumnName(t *testing.T) {
	db, err := CreateTmpTable(`CREATE TABLE test (col1 int, col2 int);`)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "-20", "10")
	AssertCommand(t, s, commands.CommandNewilist, "2", "100", "5")
	expectedError := "table test has no column named wrongcolumnname"
	AssertCommandFailure(t, s, commands.CommandAuxDbWrite, []string{"test", "wrongcolumnname", "col2", "1", "2"}, expectedError)
}

func TestCommandAuxDBWrite_Float(t *testing.T) {
	db, err := CreateTmpTable(`CREATE TABLE test (col1 float, col2 float);`)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1002.3", "-23.99999")
	AssertCommand(t, s, commands.CommandNewflist, "2", "0.42323", "5222.2321234")

	AssertCommand(t, s, commands.CommandAuxDbWrite, "test", "col1", "col2", "1", "2")

	AssertCommand(t, s, commands.CommandAuxDbRead, "3", "4", "f", "f", "SELECT col1, col2 FROM test")
	AssertValue(t, s, "3", types.NewFTFloatArray(1002.3, -23.99999))
	AssertValue(t, s, "4", types.NewFTFloatArray(0.42323, 5222.2321234))
}

func TestCommandAuxDBWrite_FloatReadIntFail(t *testing.T) {
	db, err := CreateTmpTable(`CREATE TABLE test (col1 float, col2 float);`)
	if err != nil {
		t.Fatal(err)
		//panic(err)
	}
	defer db.Close()
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1002.3", "-23.99999")
	AssertCommand(t, s, commands.CommandNewflist, "2", "0.42323", "5222.2321234")

	AssertCommand(t, s, commands.CommandAuxDbWrite, "test", "col1", "col2", "1", "2")
	expectedError := "sql: Scan error on column index 0, name \"col1\": converting driver.Value type float64 (\"1002.3\") to a int64: invalid syntax"
	AssertCommandFailure(t, s, commands.CommandAuxDbRead, []string{"3", "4", "i", "i", "SELECT col1, col2 FROM test"}, expectedError)
}

func TestCommandAuxDBWrite_Bytearray(t *testing.T) {
	// TODO: add expected fail cases
	db, err := CreateTmpTable(`CREATE TABLE test (col1 bytea);`)
	if err != nil {
		t.Fatal(err)
		//panic(err)
	}
	defer db.Close()

	s := NewTestSegment()
	AssertCommand(t, s, commands.CommandNewilist, "1", "10") // count
	AssertCommand(t, s, commands.CommandRandomArray, "2", "b32", "1")

	AssertCommand(t, s, commands.CommandAuxDbWrite, "test", "col1", "2")

	AssertCommand(t, s, commands.CommandAuxDbRead, "3", "b32", "SELECT col1 FROM test")
	bArrOriginal, _ := s.GetVariable("2")
	bArrDB, _ := s.GetVariable("3")
	assert.True(t, bArrOriginal.Equals(bArrDB), "arrays not equal")

}

func TestCommand_DeleteFile(t *testing.T) {
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	destination := "test"

	s := NewTestSegment()
	AssertCommand(t, s, commands.CommandNewilist, "1", "-10", "20")
	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "1", "array")
	AssertCommand(t, s, commands.CommandFinishSave, destination)
	s.variables.Clear() // Empty variable store to simulate a restart.
	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "2", "1", "i")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)
	AssertValue(t, s, "2", types.NewFTIntegerArray(-10, 20))

	AssertCommand(t, s, commands.CommandDelFile, destination)
}

func TestCommand_SaveLoad_IntegerArray(t *testing.T) {
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	destination := "test"

	s := NewTestSegment()
	AssertCommand(t, s, commands.CommandNewilist, "1", "-10", "20")
	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "1", "array")
	AssertCommand(t, s, commands.CommandFinishSave, destination)
	s.variables.Clear() // Empty variable store to simulate a restart.
	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "2", "1", "i")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)

	AssertValue(t, s, "2", types.NewFTIntegerArray(-10, 20))
}

func TestCommand_SaveLoad_LargeIntegerArray(t *testing.T) {
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	destination := "test"

	s := NewTestSegment()
	//AssertCommand(t, s, CommandNewilist, []string{"hl", "100000000"})
	AssertCommand(t, s, commands.CommandNewilist, "hl", "10000")
	AssertCommand(t, s, commands.CommandArange, "1", "hl")
	v, _ := s.Variables().Get("1")

	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "1", "array")
	AssertCommand(t, s, commands.CommandFinishSave, destination)

	s.Variables().Clear() // Empty variable store to simulate a restart.
	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "2", "1", "i")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)

	AssertValue(t, s, "2", v)

}

func TestCommand_SaveAndLoad_FloatArray(t *testing.T) {
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	destination := "test"

	s := NewTestSegment()
	AssertCommand(t, s, commands.CommandNewflist, "1", "-10.5", "26.314")
	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "1", "array")
	AssertCommand(t, s, commands.CommandFinishSave, destination)
	s.variables.Clear() // Empty variable store to simulate a restart.
	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "2", "1", "f")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)

	AssertValue(t, s, "2", types.NewFTFloatArray(-10.5, 26.314))
}

func TestCommand_SaveAndLoad_ByteArrayArray(t *testing.T) {
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	destination := "test"

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3")
	AssertCommand(t, s, commands.CommandNewArray, "5", "b12", "1")
	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "5", "array")
	AssertCommand(t, s, commands.CommandFinishSave, destination)
	s.variables.Clear() // Empty variable store to simulate a restart.
	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "6", "5", "b12")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)

	AssertValue(t, s, "6", types.NewFTBytearrayArrayOrPanic(
		12,
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	))
}

func TestCommand_SaveAndLoad_Listmap(t *testing.T) {
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	destination := "test"

	s := NewTestSegment()

	lm, _ := types.BuildListMapNoEncryption()
	s.Variables().Set("1", lm)

	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "1", "listmap")
	AssertCommand(t, s, commands.CommandFinishSave, destination)
	s.Variables().Clear() // Empty variable store to simulate a restart.

	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "2", "1", "listmap", "i", "f", "b3")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)
	actual, err := s.Variables().Get("2")
	if err != nil {
		t.Fail()
	}
	assert.True(t, lm.Equals(actual), "listmap not equal to expected")

}

func TestCommand_SaveAndLoad_Listmap_Empty(t *testing.T) {
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	destination := "test"

	s := NewTestSegment()

	lm, _ := types.BuildListMapNoEncryptionEmpty()
	s.Variables().Set("1", lm)

	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "1", "listmap")
	AssertCommand(t, s, commands.CommandFinishSave, destination)
	s.Variables().Clear() // Empty variable store to simulate a restart.

	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "2", "1", "listmap", "i", "f", "b3")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)
	actual, err := s.Variables().Get("2")
	if err != nil {
		t.Fail()
	}
	assert.True(t, lm.Equals(actual), "listmap not equal to expected")

}

func TestCommand_SaveAndLoad_ListmapWithChunking(t *testing.T) {
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	destination := "test"

	s := NewTestSegment()
	s.dbChunkSize = 2

	lm, _ := types.BuildListMapNoEncryption()
	s.SetVariable("1", lm)

	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "1", "listmap")
	AssertCommand(t, s, commands.CommandFinishSave, destination)
	s.Variables().Clear() // Empty variable store to simulate a restart.

	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "2", "1", "listmap", "i", "f", "b3")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)
	actual, err := s.Variables().Get("2")
	if err != nil {
		t.Fail()
	}
	assert.True(t, lm.Equals(actual), "listmap not equal to expected")

}

func TestNewListmap(t *testing.T) {
	s := NewTestSegment()

	ints := types.NewFTIntegerArray(1, 2, 3, 4, 1)
	floats := types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 1.1)
	tcString := "if"

	s.Variables().Set("1", ints)
	s.Variables().Set("2", floats)
	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}

	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error in NewListMapFromArrays")

	AssertCommand(t, s, commands.CommandNewListmap, "3", tcString, "any", "1", "2")

	AssertValue(t, s, "3", lm)
}

func TestListmapGetKeys(t *testing.T) {
	//Last key will be ignored.
	ints := types.NewFTIntegerArray(1, 2, 3, 4, 1)
	floats := types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 1.1)

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}

	expectedKey1 := types.NewKey([]interface{}{int64(1), float64(1.1)})
	expectedKey2 := types.NewKey([]interface{}{int64(2), float64(2.2)})
	expectedKey3 := types.NewKey([]interface{}{int64(3), float64(3.3)})
	expectedKey4 := types.NewKey([]interface{}{int64(4), float64(4.4)})

	expected := []types.Key{expectedKey1, expectedKey2, expectedKey3, expectedKey4}
	lm, _ := types.NewListMapFromArrays(typecode, keys, "any")

	actual := lm.GetKeys(true)
	assert.Equal(t, expected, actual, "listmap not equal to expected")
}

func TestListmapContains(t *testing.T) {
	//Last key will be ignored.
	ints := types.NewFTIntegerArray(1, 2, 3, 4, 1)
	floats := types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 1.1)

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}

	lm, _ := types.NewListMapFromArrays(typecode, keys, "any")
	expected := types.NewFTIntegerArray(1, 0, 1)

	findKey1 := types.NewKey([]interface{}{int64(2), float64(2.2)})
	findKey2 := types.NewKey([]interface{}{int64(17), float64(3.1415)})
	findKey3 := types.NewKey([]interface{}{int64(3), float64(3.3)})
	findKeys, err := types.KeysToArrayTypeVals([]types.Key{findKey1, findKey2, findKey3}, []types.TypeCode{"i", "f"})
	assert.NoError(t, err, "error in KeysToArrayTypeVals")

	actual := lm.Contains(findKeys)
	assert.Equal(t, expected, actual, "listmap not equal to expected")
}

func TestListmapAddItem_IgnoreError(t *testing.T) {
	//Last key will be ignored - duplicate.
	ints := types.NewFTIntegerArray(1, 2, 3, 4, 1)
	floats := types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 1.1)

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}
	addKey1 := types.NewKey([]interface{}{int64(2), float64(2.2)}) //Duplicate
	addKey2 := types.NewKey([]interface{}{int64(17), float64(3.1415)})
	addKey3 := types.NewKey([]interface{}{int64(3), float64(3.3)}) //Duplicate
	addKey4 := types.NewKey([]interface{}{int64(13), float64(1.41421356)})
	addKeys, err := types.KeysToArrayTypeVals([]types.Key{addKey1, addKey2, addKey3, addKey4}, typecode)
	assert.NoError(t, err, "error in KeysToArrayTypeVals")

	// Expect to get back array of keys that were added.
	expectedLm := []types.ArrayElementTypeVal{
		types.NewFTIntegerArray(17, 13),
		types.NewFTFloatArray(3.1415, 1.41421356),
	}
	// Expect to get back integer array with indices of additional keys.
	expectedResult := types.NewFTIntegerArray(4, 5)

	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err)

	actual, actualResult, err := lm.AddItems(addKeys, true)
	assert.NoError(t, err, "error in AddItems")
	assert.Equal(t, expectedLm, actual, "listmap not equal to expected")
	assert.Equal(t, expectedResult, actualResult, "unexpected result")
}

func TestListmapAddItem_DontIgnoreError(t *testing.T) {
	//Last key will be ignored - duplicate.
	ints := types.NewFTIntegerArray(1, 2, 3, 4, 1)
	floats := types.NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 1.1)

	typecode := []types.TypeCode{"i", "f"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
	}
	addKey1 := types.NewKey([]interface{}{int64(2), float64(2.2)}) //Duplicate
	addKey2 := types.NewKey([]interface{}{int64(17), float64(3.1415)})
	addKey3 := types.NewKey([]interface{}{int64(3), float64(3.3)}) //Duplicate
	addKey4 := types.NewKey([]interface{}{int64(13), float64(1.41421356)})
	addKeys, err := types.KeysToArrayTypeVals([]types.Key{addKey1, addKey2, addKey3, addKey4}, typecode)
	assert.NoError(t, err, "error in KeysToArrayTypeVals")

	lm, err := types.NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err, "error in NewListMapFromArrays")
	expectedErr := "key: {[2 2.2]} already exists in listmap"

	_, _, err = lm.AddItems(addKeys, false)
	assert.Error(t, err, "error in AddItems")
	assert.Equal(t, expectedErr, err.Error(), "not expected error")
}

func TestListmapKeysUnique(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 1))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 2, 3, 4, 1))
	s.Variables().Set("3", types.NewFTIntegerArray(1, 2, 3, 4, 5))

	AssertCommandResponse(t, s, commands.CommandListmapKeysUnique, []string{"1", "2", "3"}, "bool 1")

	s.Variables().Set("4", types.NewFTIntegerArray(1, 2, 3, 4, 1))
	AssertCommandResponse(t, s, commands.CommandListmapKeysUnique, []string{"1", "2", "4"}, "bool 0")
}

func TestListmapGetItem(t *testing.T) {
	s := NewTestSegment()

	ints := types.NewFTIntegerArray(1, 2, 3)
	floats := types.NewFTFloatArray(4.4, 5.5, 6.6)
	bytearrays := types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{1, 2, 3},
		[]byte{4, 5, 6},
		[]byte{7, 8, 9},
	)

	typecode := []types.TypeCode{"i", "f", "b3"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
		bytearrays,
	}

	lm, _ := types.NewListMapFromArrays(typecode, keys, "any")

	s.SetVariable("1", lm)
	s.SetVariable("2", ints)
	s.SetVariable("3", floats)
	s.SetVariable("4", bytearrays)
	AssertCommandResponse(t, s, commands.CommandListmapGetItem, []string{"5", "1", "0", "2", "3", "4"}, "array i 5")

	AssertValue(t, s, "5", types.NewFTIntegerArray(0, 1, 2))
}

func TestListmapSetItem(t *testing.T) {
	s := NewTestSegment()

	ints := types.NewFTIntegerArray(1, 2, 3)
	floats := types.NewFTFloatArray(4.4, 5.5, 6.6)
	bytearrays := types.NewFTBytearrayArrayOrPanic(
		3,
		[]byte{1, 2, 3},
		[]byte{4, 5, 6},
		[]byte{7, 8, 9},
	)

	typecode := []types.TypeCode{"i", "f", "b3"}

	keys := []types.ArrayElementTypeVal{
		ints,
		floats,
		bytearrays,
	}

	lm, _ := types.NewListMapFromArrays(typecode, keys, "any")
	s.SetVariable("1", lm)
	AssertCommand(t, s, commands.CommandListmapSetItems, "2", "1")
	AssertValue(t, s, "2", lm)
}

func BenchmarkListmapKeysUnique(b *testing.B) {
	log.SetOutput(io.Discard)

	for x := 0; x < 10; x++ {
		size := int(math.Pow(10, float64(x)))

		b.Run(fmt.Sprintf("Array size: %v", size), func(b *testing.B) {
			b.StopTimer()
			s := NewTestSegment()
			count := 5
			handles := make([]string, 5)

			for i := 0; i < count; i++ {
				xs := make([]int64, size)
				for j := 0; j < size; j++ {
					xs[j] = int64(j)
				}

				s.Variables().Set(variables.Handle(strconv.Itoa(i)), types.NewFTIntegerArray(xs...))
				handles[i] = strconv.Itoa(i)
			}

			b.StartTimer()

			for n := 0; n < b.N; n++ {

				response, err := s.RunCommand(commands.CommandListmapKeysUnique, handles)
				if err != nil {
					b.Error(err)
				}
				if response != "bool 1" {
					b.Error("response must be bool 1")
				}
			}
		})

	}
}

func TestCommand_SaveAndLoad_ED25519Integer(t *testing.T) {
	//t.Skip()
	db, err := CreateTmpTable(pickleTable)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	destination := "test"

	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "3")
	AssertCommand(t, s, commands.CommandNewArray, "2", "I", "1")
	AssertCommand(t, s, commands.CommandStartSave, destination)
	AssertCommand(t, s, commands.CommandSave, "2", "array")
	AssertCommand(t, s, commands.CommandFinishSave, destination)
	s.variables.Clear() // Empty variable store to simulate a restart.
	AssertCommand(t, s, commands.CommandStartLoad, destination)
	AssertCommand(t, s, commands.CommandLoad, "4", "2", "I")
	AssertCommand(t, s, commands.CommandFinishLoad, destination)

	AssertValue(t, s, "4", types.NewFTEd25519IntArrayFromInt64s(0, 0, 0))
}

func TestCommandPyLen_Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3))
	AssertCommandResponse(t, s, commands.CommandPyLen, []string{"1"}, "int 3")
}

func TestCommandPyLen_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(1.0, 2.10, 113.7, -34.22))
	AssertCommandResponse(t, s, commands.CommandPyLen, []string{"1"}, "int 4")
}

func TestCommandPyLen_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{12, 0, 7, 5},
		[]byte{13, 1, 6, 4},
		[]byte{4, 5, 7, 9},
		[]byte{15, 7, 4, 3},
	))
	AssertCommandResponse(t, s, commands.CommandPyLen, []string{"1"}, "int 4")
}

func TestCommandPyLen_Ed25519int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(1, 2, 3, 4, 5))
	AssertCommandResponse(t, s, commands.CommandPyLen, []string{"1"}, "int 5")
}

func TestCommandPyLen_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5, 6, 7))
	AssertCommandResponse(t, s, commands.CommandPyLen, []string{"1"}, "int 7")
}

func TestCommandCalcBroadcastLen(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(5))
	s.Variables().Set("2", types.NewFTIntegerArray(4))
	s.Variables().Set("3", types.NewFTIntegerArray(100))

	AssertCommand(t, s, commands.CommandCalcBroadcastLength, "4", "1", "2", "3")

	AssertValue(t, s, "4", types.NewFTIntegerArray(100))
}
func TestCommandCalcBroadcastLen_SpecChange(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(0))
	s.Variables().Set("2", types.NewFTIntegerArray(1))
	s.Variables().Set("3", types.NewFTIntegerArray(1))
	s.Variables().Set("4", types.NewFTIntegerArray(2))

	AssertCommand(t, s, commands.CommandCalcBroadcastLength, "5", "2", "3") // == 1 : If all parameters are of length 1, the return value is 1
	AssertValue(t, s, "5", types.NewFTIntegerArray(1))

	AssertCommand(t, s, commands.CommandCalcBroadcastLength, "6", "1", "2", "3") // == 0 : Otherwise, the return value is the maximum among all parameters that are not of length 1.
	AssertValue(t, s, "6", types.NewFTIntegerArray(0))

	AssertCommand(t, s, commands.CommandCalcBroadcastLength, "7", "1", "2", "3", "4") // == 2 : Otherwise, the return value is the maximum among all parameters that are not of length 1.
	AssertValue(t, s, "7", types.NewFTIntegerArray(2))
}

func TestCommandBroadcastValue_Integer(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(5))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommand(t, s, commands.CommandBroadcastValue, "1", "2")

	AssertValue(t, s, "1", types.NewFTIntegerArray(5, 5, 5, 5))
}
func TestCommandBroadcastValue_Integer_ErrorWhenNotSingleton(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(5, 3, 2))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommandFailure(t, s, commands.CommandBroadcastValue, []string{"1", "2"}, "cannot broadcast array of size 3")
}
func TestCommandBroadcastValue_Float(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(5))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommand(t, s, commands.CommandBroadcastValue, "1", "2")

	AssertValue(t, s, "1", types.NewFTFloatArray(5, 5, 5, 5))
}
func TestCommandBroadcastValue_Float_ErrorWhenNotSingleton(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(5, 4, 3))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommandFailure(t, s, commands.CommandBroadcastValue, []string{"1", "2"}, "cannot broadcast array of size 3")
}
func TestCommandBroadcastValue_Bytearray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 0, 1},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommand(t, s, commands.CommandBroadcastValue, "1", "2")

	AssertValue(t, s, "1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 0, 1},
		[]byte{0, 1, 0, 1},
		[]byte{0, 1, 0, 1},
		[]byte{0, 1, 0, 1},
	))
}
func TestCommandBroadcastValue_Bytearray_ErrorWhenNotSingleton(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 0, 1},
		[]byte{1, 1, 1, 1},
	))
	s.Variables().Set("2", types.NewFTIntegerArray(4))
	AssertCommandFailure(t, s, commands.CommandBroadcastValue, []string{"1", "2"}, "cannot broadcast array of size 2")
}
func TestCommandBroadcastValue_Ed25519Int(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(5))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommand(t, s, commands.CommandBroadcastValue, "1", "2")
	AssertValue(t, s, "1", types.NewFTEd25519IntArrayFromInt64s(5, 5, 5, 5))
}
func TestCommandBroadcastValue_Ed25519Int_ErrorWhenNotSingleton(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTEd25519IntArrayFromInt64s(5, 6, 7))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommandFailure(t, s, commands.CommandBroadcastValue, []string{"1", "2"}, "cannot broadcast array of size 3")
}
func TestCommandBroadcastValue_Ed25519(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(5))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommand(t, s, commands.CommandBroadcastValue, "1", "2")

	AssertValue(t, s, "1", types.NewEd25519ArrayFromInt64sOrPanic(5, 5, 5, 5))
}
func TestCommandBroadcastValue_Ed25519_ErrorWhenNotSingleton(t *testing.T) {
	types.SkipIfGPUUnavailable(t)
	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(5, 2))
	s.Variables().Set("2", types.NewFTIntegerArray(4))

	AssertCommandFailure(t, s, commands.CommandBroadcastValue, []string{"1", "2"}, "only singleton arrays can be broadcasted")
}
func TestCommandToList_Int(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1", "2", "3")
	res, err := s.RunCommand(commands.CommandToList, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "intlist 1 2 3", res, "values should be the same")
}

func TestCommandToList_Float(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewflist, "1", "1.6", "-2.2", "300.15")
	res, err := s.RunCommand(commands.CommandToList, []string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "floatlist 1.6 -2.2 300.15", res, "values should be the same")
}

func TestCommandEqualInt(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("2", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	s.Variables().Set("3", types.NewFTIntegerArray(6, 7, 8, 9, 0))

	AssertCommandResponse(t, s, commands.CommandEqualInt, []string{"1", "2"}, "bool 1")
	AssertCommandResponse(t, s, commands.CommandEqualInt, []string{"1", "3"}, "bool 0")
}

func TestCommandSliceToIndices(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(0))
	s.Variables().Set("2", types.NewFTIntegerArray(3))
	s.Variables().Set("3", types.NewFTFloatArray(1.6, -2.2, 300.15, 5.0, -8.2))

	AssertCommand(t, s, commands.CommandSliceToIndices, "4", "3", "1", "2")

	AssertValue(t, s, "4", types.NewFTIntegerArray(0, 1, 2))

	s.Variables().Set("5", types.NewFTIntegerArray(-1))
	AssertCommand(t, s, commands.CommandSliceToIndices, "6", "3", "1", "5")
	AssertValue(t, s, "6", types.NewFTIntegerArray(0, 1, 2, 3))
}
func TestCommandEmptySliceToIndices(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray())
	s.Variables().Set("2", types.NewFTIntegerArray(0))

	AssertCommand(t, s, commands.CommandSliceToIndices, "3", "1", "2")

	AssertValue(t, s, "3", types.NewFTIntegerArray())
}

func TestCommandEdFolded_EdFoldedProject(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))

	AssertCommand(t, s, commands.CommandEdFolded, "2", "1") // Ed25519 -> Bytearray

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		32,
		[]byte{88, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102},
		[]byte{201, 163, 248, 106, 174, 70, 95, 14, 86, 81, 56, 100, 81, 15, 57, 151, 86, 31, 162, 201, 232, 94, 162, 29, 194, 41, 35, 9, 243, 205, 96, 34},
		[]byte{212, 180, 245, 120, 72, 104, 195, 2, 4, 3, 36, 103, 23, 236, 22, 159, 247, 158, 38, 96, 142, 161, 38, 161, 171, 105, 238, 119, 209, 177, 103, 18},
		[]byte{47, 17, 50, 202, 97, 171, 56, 223, 240, 15, 47, 234, 50, 40, 242, 76, 108, 113, 213, 128, 133, 184, 14, 71, 225, 149, 21, 203, 39, 232, 208, 71},
		[]byte{237, 200, 118, 214, 131, 31, 210, 16, 93, 11, 67, 137, 202, 46, 40, 49, 102, 70, 146, 137, 20, 110, 44, 224, 111, 174, 254, 152, 178, 37, 72, 223},
	))

	AssertCommand(t, s, commands.CommandEdFoldedProject, "3", "2") // Bytearray -> Ed25519
	AssertValue(t, s, "3", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
}

func TestCommandEdAffine_EdAffineProject(t *testing.T) {
	types.SkipIfGPUUnavailable(t)

	s := NewTestSegment()

	s.Variables().Set("1", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))

	AssertCommand(t, s, commands.CommandEdAffine, "2", "1") // Ed25519 -> Bytearray

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		64,
		[]byte{26, 213, 37, 143, 96, 45, 86, 201, 178, 167, 37, 149, 96, 199, 44, 105, 92, 220, 214, 253, 49, 226, 164, 192, 254, 83, 110, 205, 211, 54, 105, 33, 88, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102, 102},
		[]byte{14, 206, 67, 40, 78, 161, 197, 131, 95, 164, 215, 21, 69, 142, 13, 8, 172, 231, 51, 24, 125, 59, 4, 61, 108, 4, 90, 159, 76, 56, 171, 54, 201, 163, 248, 106, 174, 70, 95, 14, 86, 81, 56, 100, 81, 15, 57, 151, 86, 31, 162, 201, 232, 94, 162, 29, 194, 41, 35, 9, 243, 205, 96, 34},
		[]byte{92, 226, 248, 211, 95, 72, 98, 172, 134, 72, 98, 129, 25, 152, 67, 99, 58, 200, 218, 62, 116, 174, 244, 31, 73, 143, 146, 34, 74, 156, 174, 103, 212, 180, 245, 120, 72, 104, 195, 2, 4, 3, 36, 103, 23, 236, 22, 159, 247, 158, 38, 96, 142, 161, 38, 161, 171, 105, 238, 119, 209, 177, 103, 18},
		[]byte{112, 248, 201, 196, 87, 166, 58, 73, 71, 21, 206, 147, 193, 158, 115, 26, 249, 32, 53, 122, 184, 212, 37, 131, 70, 241, 207, 86, 219, 168, 61, 32, 47, 17, 50, 202, 97, 171, 56, 223, 240, 15, 47, 234, 50, 40, 242, 76, 108, 113, 213, 128, 133, 184, 14, 71, 225, 149, 21, 203, 39, 232, 208, 71},
		[]byte{51, 242, 46, 50, 192, 156, 64, 145, 165, 225, 27, 62, 249, 25, 40, 92, 222, 165, 45, 209, 247, 124, 239, 252, 123, 88, 227, 173, 62, 167, 253, 73, 237, 200, 118, 214, 131, 31, 210, 16, 93, 11, 67, 137, 202, 46, 40, 49, 102, 70, 146, 137, 20, 110, 44, 224, 111, 174, 254, 152, 178, 37, 72, 95},
	))

	AssertCommand(t, s, commands.CommandEdAffineProject, "3", "2") // Bytearray -> Ed25519
	AssertValue(t, s, "3", types.NewEd25519ArrayFromInt64sOrPanic(1, 2, 3, 4, 5))
}

func TestParseMessage(t *testing.T) {
	b := []byte("{\"command\": \"command_newilist 0 0\", \"response_required\": \"True\"}")

	expectedName := "command_newilist"
	expectedArgs := []string{"0", "0"}
	expectedResponseRequired := true

	actualName, actualArgs, actualResponseRequired, err := parseMessage(b)
	assert.NoError(t, err, "error in parseMessage")
	assert.Equal(t, expectedName, actualName, "unexpected name")
	assert.Equal(t, expectedArgs, actualArgs, "unexpected args")
	assert.Equal(t, expectedResponseRequired, actualResponseRequired, "unexpected response")
}

func TestVariablesGetAsInteger(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(12))
	expected := int64(12)
	actual, err := variables.GetAsInteger(s.Variables(), "1")
	assert.NoError(t, err, "error in GetAsInteger")
	assert.Equal(t, expected, actual, "unexpected value")
}
func TestVariablesGetAsInteger_NotSingletonArray(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTIntegerArray(1, 2, 3, 4, 5))
	expectedErr := "not a singleton array"
	_, err := variables.GetAsInteger(s.Variables(), "1")
	assert.Error(t, err, "no error in GetAsInteger")
	assert.Equal(t, expectedErr, err.Error(), "not the expected error")
}

func TestVariablesGetAsInteger_NotInteger(t *testing.T) {
	s := NewTestSegment()

	s.Variables().Set("1", types.NewFTFloatArray(3.141592654))
	expectedErr := "GetAs: interface conversion: types.TypeVal is *types.FTFloatArray, not *types.FTIntegerArray"
	_, err := variables.GetAsInteger(s.Variables(), "1")
	assert.Error(t, err, "error in GetAsInteger")
	assert.Equal(t, expectedErr, err.Error(), "not the expected error")
}

func TestVariableGetAsFloat(t *testing.T) {
	s := NewTestSegment()
	s.Variables().Set("1", types.NewFTFloatArray(2.7182818285))
	expected := float64(2.7182818285)
	actual, err := variables.GetAsFloat(s.Variables(), "1")
	assert.NoError(t, err, "error in GetAsFloat")
	assert.Equal(t, expected, actual)
}

func TestVariableGetAsEd25519Int(t *testing.T) {
	s := NewTestSegment()
	s.Variables().Set("1", types.NewFTEd25519IntArray(types.Int64ToScalar(2)))
	expected := types.Int64ToScalar(2)
	actual, err := variables.GetAsEd25519Integer(s.Variables(), "1")
	assert.NoError(t, err, "error in GetAsEd25519Integer")
	assert.Equal(t, expected, actual, "unexpected value returned")
}

func TestVariableGetAsBytes(t *testing.T) {
	s := NewTestSegment()
	s.Variables().Set("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
	))
	expected := []byte{0, 1, 2, 3}
	actual, err := variables.GetAsBytes(s.Variables(), "1")
	assert.NoError(t, err, "error in GetAsBytes")
	assert.Equal(t, expected, actual, "unexpected value returned")
}
