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
	"fmt"
	"os"
	"testing"

	"golang.org/x/exp/slices"
)

var GPUHasBeenInitialized = false

func TestMain(m *testing.M) {

	if (*EnableGPUFlag || EnableGPUEnv) && !GPUHasBeenInitialized {
		fmt.Printf("Initializing GPU\n")
		err := InitializeGPU()
		if err != nil {
			panic(err)
		}
		GPUHasBeenInitialized = true
	}

	code := m.Run()
	os.Exit(code)
}

func TestErrorString(t *testing.T) {
	errString := FTError(1).Error()
	if errString != "out of memory" {
		t.Errorf("FTError(1) should return 'out of memory', but got '%v'", errString)
	}
}

func TestArrayInit_NoValue(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519Array(2, nil)
	defer xs.Free()

	if err != nil {
		t.Fatal(err)
	}

	if xs.Length() != 2 {
		t.Errorf("Expected length to be 2")
	}
}

func TestArrayInit_WithValue(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519Array(2, Int64ToScalar(5))
	defer xs.Free()

	if err != nil {
		t.Fatal(err)
	}

	if xs.Length() != 2 {
		t.Errorf("Expected length to be 2")
	}

	ys, err := NewEd25519ArrayFromInt64s(5, 5)
	if err != nil {
		t.Fatal(err)
	}

	if !xs.Equals(ys) {
		t.Error("Array was expected to be {5,5}")
	}
}

func TestArrayInit_ZeroLength(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519Array(0, nil)
	defer xs.Free()

	if err != nil {
		t.Fatal(err)
	}

	if xs.Length() != 0 {
		t.Errorf("Expected length to be 2")
	}
}

func TestArrayToBytes(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519ArrayFromInt64s(5, 10)
	defer xs.Free()
	if err != nil {
		t.Fatal(err)
	}
}

func TestArrayToFromFoldedBytes(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519ArrayFromInt64s(5, 10, 15, 20, 25, 30, 35, 40)
	if err != nil {
		t.Fatal(err)
	}
	defer xs.Free()

	bs, _ := xs.ToFoldedBytes()

	ys, err := NewEd25519ArrayFromFoldedBytes(bs)
	if err != nil {
		t.Fatal(err)
	}
	defer ys.Free()

	if !xs.Equals(ys) {
		t.Fatalf("%v\n\n<>\n\n%v", bs, ys.ToBytes())
	}

}

func TestArrayFromBytes(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519ArrayFromInt64s(5, 10)
	defer xs.Free()
	if err != nil {
		t.Fatal(err)
	}

	bs := xs.ToBytes()

	ys, err := NewEd25519ArrayFromBytes(bs)
	defer ys.Free()
	if err != nil {
		t.Fatal(err)
	}

	if !xs.Equals(ys) {
		t.Fatalf("%v\n\n<>\n\n%v", bs, ys.ToBytes())
	}
}

func TestArrayFromEmptyBytes(t *testing.T) {
	SkipIfGPUUnavailable(t)

	bs := []byte{}

	xs, err := NewEd25519ArrayFromBytes(bs)
	if err != nil {
		t.Fatal(err)
	}
	defer xs.Free()

	if !xs.IsEmpty() {
		t.Errorf("array should be empty")
	}
}
func TestArrayFromEmptyAffineBytes(t *testing.T) {
	SkipIfGPUUnavailable(t)

	bs := []byte{}

	xs, err := NewEd25519ArrayFromAffineBytes(bs)
	defer xs.Free()
	if err != nil {
		t.Fatal(err)
	}

	if !xs.IsEmpty() {
		t.Errorf("array should be empty")
	}
}

func TestArrayFromEmptyFoldedBytes(t *testing.T) {
	SkipIfGPUUnavailable(t)

	bs := []byte{}

	xs, err := NewEd25519ArrayFromFoldedBytes(bs)
	defer xs.Free()
	if err != nil {
		t.Fatal(err)
	}

	if !xs.IsEmpty() {
		t.Errorf("array should be empty")
	}
}

func TestArrayFromInts(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519ArrayFromInt64s(5, 10)
	defer xs.Free()
	if err != nil {
		t.Fatal(err)
	}

	ys, err := NewEd25519ArrayFromInt64s(5, 11)
	defer ys.Free()
	if err != nil {
		t.Fatal(err)
	}

	zs, err := NewEd25519ArrayFromInt64s(5, 10)
	defer zs.Free()
	if err != nil {
		t.Fatal(err)
	}

	if xs.Equals(ys) {
		t.Fatal("xs and ys should not be considered equal")
	}
	if !xs.Equals(zs) {
		t.Fatal("xs and zs should be considered equal")
	}
}

func TestArrayFromInt64s(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519ArrayFromInt64s(5, 10)
	defer xs.Free()
	if err != nil {
		t.Fatal(err)
	}

	ys, err := NewEd25519ArrayFromInt64s(5, 11)
	defer ys.Free()
	if err != nil {
		t.Fatal(err)
	}

	zs, err := NewEd25519ArrayFromInt64s(5, 10)
	defer zs.Free()
	if err != nil {
		t.Fatal(err)
	}

	if xs.Equals(ys) {
		t.Fatal("xs and ys should not be considered equal")
	}
	if !xs.Equals(zs) {
		t.Fatal("xs and zs should be considered equal")
	}
}

func TestIndex(t *testing.T) {
	SkipIfGPUUnavailable(t)

	values := NewEd25519ArrayFromInt64sOrPanic(0, 1, 2, 0, 1, 2)

	expected := []int64{1, 2, 4, 5}
	actual := values.Index()

	if !slices.Equal(actual.array, expected) {
		t.Fatalf("index: %v != %v", actual, expected)
	}
}

func TestReduceISum(t *testing.T) {
	SkipIfGPUUnavailable(t)

	arr, err := NewEd25519ArrayFromInt64s(5, 8, 4, 5, 50)
	if err != nil {
		t.Fatal(err)
	}
	defer arr.Free()
	arr2, err := NewEd25519ArrayFromInt64s(10, 20, 30)
	if err != nil {
		t.Fatal(err)
	}
	defer arr2.Free()

	keys := NewFTIntegerArray(2, 2, 3)

	arr.ReduceISum(keys, arr2)

	arr3, err := NewEd25519ArrayFromInt64s(5, 8, 34, 35, 50)
	if err != nil {
		t.Fatal(err)
	}
	defer arr3.Free()

	if !arr3.Equals(arr) {
		t.Error("arrays not equal")
	}
}

func TestPairwiseEquals(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519ArrayFromInt64s(5, 10)
	defer xs.Free()
	if err != nil {
		t.Fatal(err)
	}

	ys, err := NewEd25519ArrayFromInt64s(5, 10)
	defer ys.Free()
	if err != nil {
		t.Fatal(err)
	}

	results, err := xs.Eq(ys)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(results.Values(), []int64{1, 1}) {
		t.Errorf("Expected [1,1] but got %v", results)
	}
}

func TestPairwiseNotEquals(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519ArrayFromInt64s(5, 10)
	defer xs.Free()
	if err != nil {
		t.Fatal(err)
	}

	ys, err := NewEd25519ArrayFromInt64s(5, 11)
	defer ys.Free()
	if err != nil {
		t.Fatal(err)
	}

	results, err := xs.Ne(ys)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(results.Values(), []int64{0, 1}) {
		t.Errorf("Expected {1,0} but got %v", results)
	}
}

func TestReduceSumAssignToSubsetIssue(t *testing.T) {
	SkipIfGPUUnavailable(t)

	xs, err := NewEd25519Array(5, nil)
	if err != nil {
		t.Error(err)
	}
	defer xs.Free()

	ys, err := NewEd25519Array(30, nil)
	if err != nil {
		t.Error(err)
	}
	defer ys.Free()

	is := &FTIntegerArray{make([]int64, 30)}

	err = xs.ReduceSum(is, ys)
	if err != nil {
		t.Error(err)
	}
}

func TestAssignToSubsetIssue(t *testing.T) {
	SkipIfGPUUnavailable(t)

	is := make([]int64, 30)
	is[29] = 1

	xs, err := NewEd25519Array(5, nil)
	if err != nil {
		t.Error(err)
	}
	defer xs.Free()

	os := make([]int64, 30)
	for i := range os {
		os[i] = 1
	}
	os[29] = 2

	ones, err := NewEd25519ArrayFromInt64s(os...)
	if err != nil {
		t.Error(err)
	}
	defer ones.Free()

	// Assign 1 to index 0, 30 times
	err = xs.AssignToSubset(is, ones)
	if err != nil {
		t.Error(err)
	}

	comp, err := NewEd25519ArrayFromInt64s(1, 2, 0, 0, 0)
	if err != nil {
		t.Error(err)
	}

	if !xs.Equals(comp) {
		t.Error("result should be 1,2,0,0,0")
	}

}
