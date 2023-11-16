// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package types

import (
	"flag"
	"os"
	"strconv"
)

// TODO: use a package to handle flags, environment variables and config files
var EnableGPUEnv, _ = strconv.ParseBool(os.Getenv("gpu"))

// Must stay as *bool here to allow the Go initilisation code to update the value after
// this line has been executed.
var EnableGPUFlag = flag.Bool("gpu", false, "enable testing of Ed25519/GPU code")

type Helper interface {
	Helper()
	Skip(args ...any)
}

func SkipIfGPUUnavailable(t Helper) {
	t.Helper()

	if !*EnableGPUFlag && !EnableGPUEnv {
		t.Skip("skipping due to the GPU being unavailable")
	}
}

func BuildListMap() (*ListMap, []ArrayElementTypeVal) {
	ints := NewFTIntegerArray(1, 2, 3)
	floats := NewFTFloatArray(4.4, 5.5, 6.6)
	bytearrays := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{1, 2, 3},
		[]byte{4, 5, 6},
		[]byte{7, 8, 9},
	)
	ed25519Ints := NewFTEd25519IntArrayFromInt64s(5, 5, 5)

	typecode := []TypeCode{"i", "f", "b3", "I"}

	keys := []ArrayElementTypeVal{
		ints,
		floats,
		bytearrays,
		ed25519Ints,
	}

	lm, _ := NewListMapFromArrays(typecode, keys, "any")
	return lm, keys
}

func BuildListMapNoEncryption() (*ListMap, []ArrayElementTypeVal) {
	ints := NewFTIntegerArray(1, 2, 3)
	floats := NewFTFloatArray(4.4, 5.5, 6.6)
	bytearrays := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{1, 2, 3},
		[]byte{4, 5, 6},
		[]byte{7, 8, 9},
	)

	typecode := []TypeCode{"i", "f", "b3"}

	keys := []ArrayElementTypeVal{
		ints,
		floats,
		bytearrays,
	}

	lm, _ := NewListMapFromArrays(typecode, keys, "any")
	return lm, keys
}

func BuildListMapNoEncryptionEmpty() (*ListMap, []ArrayElementTypeVal) {
	ints := NewFTIntegerArray()
	floats := NewFTFloatArray()
	bytearrays := NewFTBytearrayArrayOrPanic(3)

	typecode := []TypeCode{"i", "f", "b3"}

	keys := []ArrayElementTypeVal{
		ints,
		floats,
		bytearrays,
	}

	lm, _ := NewListMapFromArrays(typecode, keys, "any")
	return lm, keys
}
