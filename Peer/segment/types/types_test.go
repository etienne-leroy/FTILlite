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
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

const DeltaSliceAccuracy = 0.0001

func TestFloatArray_Equal(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	xt := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	actual := xs.Equals(xt)
	assert.True(t, actual)
}

func TestFloatArray_GetBinary(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3)
	expected := []byte{0x9a, 0x99, 0x99, 0x99, 0x99, 0x99, 0xf1, 0x3f,
		0x9a, 0x99, 0x99, 0x99, 0x99, 0x99, 0x1, 0x40,
		0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0xa, 0x40}
	actual, err := xs.GetBinaryArray(-1)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloatArray_AsType(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3)
	expectedErr := "conversion not supported: f -> i"
	_, err := xs.AsType("i")
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestFloatArray_Sort(t *testing.T) {
	xs := NewFTFloatArray(10.01, 4.4, 2.2, 3.3, 6.6, 7.7, 5.5, 8.8, 9.9, 1.1)
	expected := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	actual := xs.Sort()
	assert.Equal(t, expected, actual)
}

func TestFloatArray_IndexSort(t *testing.T) {
	xs := NewFTFloatArray(1.1, 1.1, 5.5, 1.1, 2.2)
	index := NewFTIntegerArray(1, 2, 3, 0, 4)
	expected := NewFTIntegerArray(0, 1, 2, 4, 3)
	actual, err := xs.IndexSort(index)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_SetLength(t *testing.T) {
	xs := NewFTFloatArray(1.1, 1.1, 5.5, 2.2)
	expectedLength := int64(10)
	err := xs.SetLength(expectedLength)
	assert.NoError(t, err)
	assert.Equal(t, expectedLength, xs.Length())
}

func TestFloat_Remove(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	index := NewFTIntegerArray(2, 3, 5)
	expected := NewFTFloatArray(1.1, 2.2, 5.5, 7.7, 8.8, 9.9, 10.01)
	err := xs.Remove(index)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_Broadcast(t *testing.T) {
	xs := NewFTFloatArray(1.1)
	err := xs.Broadcast(int64(10))
	expected := NewFTFloatArray(1.1, 1.1, 1.1, 1.1, 1.1, 1.1, 1.1, 1.1, 1.1, 1.1)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_Broadcast_Failure(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2)
	expectedErr := "cannot broadcast array of size 2"
	err := xs.Broadcast(int64(10))
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestFloat_Lookup_NoDefault(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	index := NewFTIntegerArray(1, 2, 9, 11)
	expected := NewFTFloatArray(2.2, 3.3, 10.01, 0)
	actual, err := xs.Lookup(index, nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Lookup_Default(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	index := NewFTIntegerArray(1, 2, 9, 11)
	expected := NewFTFloatArray(2.2, 3.3, 10.01, 99)
	actual, err := xs.Lookup(index, NewFTFloatArray(99))
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Get_NoIndexOrDefault(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	expected := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	actual, err := xs.Get(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Get_NoDefault(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	index := NewFTIntegerArray(1, 2, 9, 11)
	expectedErr := "out of range: 11"
	_, err := xs.Get(index, nil)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestFloat_Get_Default(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	index := NewFTIntegerArray(1, 2, 9, 11)
	expected := NewFTFloatArray(2.2, 3.3, 10.01, 99)
	actual, err := xs.Lookup(index, NewFTFloatArray(99))
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Set(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	index := NewFTIntegerArray(1, 2, 9)
	values := NewFTFloatArray(9.99, 8.88, 7.77)
	expected := NewFTFloatArray(1.1, 9.99, 8.88, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 7.77)
	err := xs.Set(index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_Index(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	expected := ArangeFTIntegerArray(int64(10))
	actual := xs.Index()
	assert.Equal(t, expected, actual)
}

func TestFloat_Contains(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7, 8.8, 9.9, 10.01)
	values := NewFTFloatArray(1.1, 5.5, 99.3, 10.01, 6)
	expected := NewFTIntegerArray(1, 1, 0, 1, 0)
	actual, err := xs.Contains(values)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_ReduceSum(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	values := NewFTFloatArray(5.5, 12.4, 2.2)
	index := NewFTIntegerArray(2, 2, 3)
	expected := NewFTFloatArray(1, 2, 17.9, 2.2, 5, 6, 7)
	err := xs.ReduceSum(index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_ReduceISum(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	values := NewFTFloatArray(5.5, 12.4, 2.2)
	index := NewFTIntegerArray(2, 2, 3)
	expected := NewFTFloatArray(1, 2, 20.9, 6.2, 5, 6, 7)
	err := xs.ReduceISum(index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_ReduceMin(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	values := NewFTFloatArray(5.5, 12.4, 2.2)
	index := NewFTIntegerArray(2, 2, 3)
	expected := NewFTFloatArray(1, 2, 5.5, 2.2, 5, 6, 7)
	err := xs.ReduceMin(index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_ReduceIMin(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 2.0, 5.0, 6.0, 7.0)
	values := NewFTFloatArray(5.5, 12.4, 2.2)
	index := NewFTIntegerArray(2, 2, 3)
	expected := NewFTFloatArray(1, 2, 3.0, 2.0, 5, 6, 7)
	err := xs.ReduceIMin(index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_ReduceMax(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	values := NewFTFloatArray(5.5, 12.4, 2.2)
	index := NewFTIntegerArray(2, 2, 3)
	expected := NewFTFloatArray(1, 2, 12.4, 2.2, 5, 6, 7)
	err := xs.ReduceMax(index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_ReduceIMax(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	values := NewFTFloatArray(5.5, 12.4, 2.2)
	index := NewFTIntegerArray(2, 2, 3)
	expected := NewFTFloatArray(1.0, 2.0, 12.4, 4.0, 5.0, 6.0, 7.0)
	err := xs.ReduceIMax(index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestFloat_Cumsum(t *testing.T) {
	xs := NewFTFloatArray(1.1, 2.1, 3.3, 4.4, 5.5, 6.6, 7.7)
	expected := NewFTFloatArray(1.1, 3.2, 6.5, 10.9, 16.4, 23, 30.7)
	actual, err := xs.CumSum()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Mux(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	condition := NewFTIntegerArray(1, 0, 1, 2, 0, 0, 3)            // Cond
	falseVal := NewFTFloatArray(1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7) // if false
	expected := NewFTFloatArray(1.0, 2.2, 3.0, 4.0, 5.5, 6.6, 7.0)
	actual, err := xs.Mux(condition, falseVal)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_NewRandom(t *testing.T) {
	min := float64(2)
	max := float64(10)
	length := int64(5)
	xs, err := NewRandomFTFloatArray(min, max, length)
	assert.NoError(t, err)
	expectedTypeCode := TypeCode("f")
	assert.Equal(t, expectedTypeCode, xs.TypeCode())
	assert.Equal(t, length, xs.Length())
	for _, e := range xs.array {
		assert.LessOrEqual(t, min, e)
		assert.GreaterOrEqual(t, max, e)
	}
}

func TestFloat_NewFromBytes(t *testing.T) {
	b := []byte{
		0, 0, 0, 0, 0, 0, 240, 63,
		0, 0, 0, 0, 0, 0, 0, 64,
		0, 0, 0, 0, 0, 0, 8, 64,
		0, 0, 0, 0, 0, 0, 16, 64,
		0, 0, 0, 0, 0, 0, 20, 64,
		0, 0, 0, 0, 0, 0, 24, 64,
		0, 0, 0, 0, 0, 0, 28, 64,
	}
	xs, err := NewFTFloatArrayFromBytes(b)
	assert.NoError(t, err)
	expected := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	assert.Equal(t, expected, xs)
}

func TestFloat_Eq(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	xt := NewFTFloatArray(1.0, 2.0, 1.0, 4.1, 5.0, 6.0, 77.0)
	expected := NewFTIntegerArray(1, 1, 0, 0, 1, 1, 0)
	actual, err := xs.Eq(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Ne(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	xt := NewFTFloatArray(1.0, 2.0, 1.0, 4.1, 5.0, 6.0, 77.0)
	expected := NewFTIntegerArray(0, 0, 1, 1, 0, 0, 1)
	actual, err := xs.Ne(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Gt(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	xt := NewFTFloatArray(0.1, 2.0, 3.1, 4.1, 5.0, 5.9, 77.0)
	expected := NewFTIntegerArray(1, 0, 0, 0, 0, 1, 0)
	actual, err := xs.Gt(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Lt(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	xt := NewFTFloatArray(0.1, 2.0, 3.1, 4.1, 5.0, 5.9, 77.0)
	expected := NewFTIntegerArray(0, 0, 1, 1, 0, 0, 1)
	actual, err := xs.Lt(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Ge(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	xt := NewFTFloatArray(0.1, 2.0, 3.1, 4.1, 5.0, 5.9, 77.0)
	expected := NewFTIntegerArray(1, 1, 0, 0, 1, 1, 0)
	actual, err := xs.Ge(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Le(t *testing.T) {
	xs := NewFTFloatArray(1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	xt := NewFTFloatArray(0.1, 2.0, 3.1, 4.1, 5.0, 5.9, 77.0)
	expected := NewFTIntegerArray(0, 1, 1, 1, 1, 0, 1)
	actual, err := xs.Le(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Neg(t *testing.T) {
	xs := NewFTFloatArray(1.0, -2.0, 3.0, -4.0, 5.0, -6.0, 7.0)
	expected := NewFTFloatArray(-1.0, 2.0, -3.0, 4.0, -5.0, 6.0, -7.0)
	actual, err := xs.Neg()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Abs(t *testing.T) {
	xs := NewFTFloatArray(0.0, 1.0, -2.0, 3.0, -4.0, 5.0, -6.0, 7.0)
	expected := NewFTFloatArray(0.0, 1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0)
	actual, err := xs.Abs()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Floor(t *testing.T) {
	xs := NewFTFloatArray(0.1, 2.0, 3.1, -4.1, 5.0, 5.9, 77.0)
	expected := NewFTIntegerArray(0, 2, 3, -5, 5, 5, 77)
	actual, err := xs.Floor()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Ceil(t *testing.T) {
	xs := NewFTFloatArray(-0.1, 0.1, 2.0, -3.1, 4.1, 5.0, 5.9, 77.1)
	expected := NewFTIntegerArray(0, 1, 2, -3, 5, 5, 6, 78)
	actual, err := xs.Ceil()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Round(t *testing.T) {
	xs := NewFTFloatArray(-0.1, 0.1, 2.0, -3.1, 4.1, 5.0, 5.9, 77.1, 5.5)
	expected := NewFTIntegerArray(0, 0, 2, -3, 4, 5, 6, 77, 6)
	actual, err := xs.Round()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Trunc(t *testing.T) {
	xs := NewFTFloatArray(-0.1, 0.1, 2.0, -3.1, 4.1, 5.0, 5.9, 77.1, 5.5)
	expected := NewFTFloatArray(0, 0, 2, -3, 4, 5, 5, 77, 5)
	actual, err := xs.Trunc()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestFloat_Add(t *testing.T) {
	addend1 := NewFTFloatArray(-0.1, 0.1, 2.0, -3.1, 4.1, 5.0, 5.9, 77.1, 5.5)
	addend2 := NewFTFloatArray(-0.2, 7.1, 2.0, 5.2, 4.1, -8.2, 0.01, -77.1, 5.5)
	expected := NewFTFloatArray(-0.3, 7.2, 4.0, 2.1, 8.2, -3.2, 5.91, 0, 11.0)
	actual, err := addend1.Add(addend2)
	assert.NoError(t, err)
	actualFTFloat, err := asFTFloatArray(actual)
	assert.NoError(t, err)
	// Use InDeltaSlice instead of equals to handle floating point errors
	assert.InDeltaSlice(t, expected.array, actualFTFloat.array, DeltaSliceAccuracy)
}

func TestFloat_Sub(t *testing.T) {
	minuend := NewFTFloatArray(-0.1, 0.1, 2.0, -3.1, 4.1, 5.0, 5.9, 77.1, 5.5)
	subtrahend := NewFTFloatArray(-0.2, 7.1, 2.0, 5.2, 4.1, -8.2, 0.01, -77.1, 5.5)
	expected := NewFTFloatArray(0.1, -7.0, 0.0, -8.3, 0.0, 13.2, 5.89, 154.2, 0.0)
	actual, err := minuend.Sub(subtrahend)
	assert.NoError(t, err)
	actualFTFloat, err := asFTFloatArray(actual)
	assert.NoError(t, err)
	// Use InDeltaSlice instead of equals to handle floating point errors
	assert.InDeltaSlice(t, expected.array, actualFTFloat.array, DeltaSliceAccuracy)
}

func TestFloat_Mul(t *testing.T) {
	multiplier := NewFTFloatArray(-0.1, 0.1, 2.0, -3.1, 4.1, 5.0, 5.9, 77.1, 5.5)
	multiplicand := NewFTFloatArray(-0.2, 7.1, 2.0, 5.2, 4.1, -8.2, 0.01, -77.1, 5.5)
	expectedProduct := NewFTFloatArray(0.02, 0.71, 4.0, -16.12, 16.81, -41.0, 0.059, -5944.41, 30.25)
	actualProduct, err := multiplier.Mul(multiplicand)
	assert.NoError(t, err)
	actualProductFTFloat, err := asFTFloatArray(actualProduct)
	assert.NoError(t, err)
	assert.InDeltaSlice(t, expectedProduct.array, actualProductFTFloat.array, DeltaSliceAccuracy)
}

func TestFloat_TrueDiv(t *testing.T) {
	dividend := NewFTFloatArray(-0.3, 7.1, 2.0, 10.0, 4.1, -8.2, 0.01, 5.5)
	divisor := NewFTFloatArray(-0.1, 0.1, 2.0, 4.0, 4.1, 4.1, 10.0, -5.5)
	expectedQuotient := NewFTFloatArray(3.0, 71.0, 1.0, 2.5, 1, -2.0, 0.001, -1.0)
	actualQuotient, err := dividend.TrueDiv(divisor)
	assert.NoError(t, err)
	actualQuotientFTFloat, err := asFTFloatArray(actualQuotient)
	assert.NoError(t, err)
	assert.InDeltaSlice(t, expectedQuotient.array, actualQuotientFTFloat.array, DeltaSliceAccuracy)
}

func TestFloat_Pow(t *testing.T) {
	bases := NewFTFloatArray(2.0, 3.5, -3.0, -2.0, 17.0, 2.0)
	powers := NewFTIntegerArray(4, 2, 3, 4, 0, -2)
	expected := NewFTFloatArray(16.0, 12.25, -27.0, 16.0, 1.0, 0.25)
	actual, err := bases.Pow(powers)
	assert.NoError(t, err)
	actualFTFloat, err := asFTFloatArray(actual)
	assert.NoError(t, err)
	assert.InDeltaSlice(t, expected.array, actualFTFloat.array, DeltaSliceAccuracy)
}

func TestFloat_Pow_NotEnoughPowers(t *testing.T) {
	bases := NewFTFloatArray(2.0, 3.5, -3.0, -2.0, 17.0, 2.0)
	powers := NewFTIntegerArray(4, 2, 3, 4, 0)
	_, err := bases.Pow(powers)
	expectedErr := "not enough values in other array"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestFloat_Pow_TooManyPowers(t *testing.T) {
	t.Skip()
	bases := NewFTFloatArray(2.0, 3.5, -3.0, -2.0, 17.0, 2.0)
	powers := NewFTIntegerArray(4, 2, 3, 4, 0, -2, 1)
	_, err := bases.Pow(powers)
	expectedErr := "not enough values in other array"
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err.Error())
}

func TestFloat_Exp(t *testing.T) {
	powers := NewFTFloatArray(0.0, 1.0, 2.0, -1.0, -2.0, 0.5, -0.5)
	expected := NewFTFloatArray(1.0, 2.7182818284, 7.3890560989, 0.3678794411, 0.1353352832, 1.6487212707, 0.6065306597)
	actual, err := powers.Exp()
	assert.NoError(t, err)
	actualFTFloat, err := asFTFloatArray(actual)
	assert.NoError(t, err)
	assert.InDeltaSlice(t, expected.array, actualFTFloat.array, DeltaSliceAccuracy)
}

func TestFloat_Log(t *testing.T) {
	xs := NewFTFloatArray(0.1, 1.0, 10.0)
	expected := NewFTFloatArray(-2.3025850930, 0.0, 2.3025850930)
	actual, err := xs.Log()
	assert.NoError(t, err)
	actualFTFloat, err := asFTFloatArray(actual)
	assert.NoError(t, err)
	assert.InDeltaSlice(t, expected.array, actualFTFloat.array, DeltaSliceAccuracy)
}

func TestFloat_Log_ErrorNegative(t *testing.T) {
	xs := NewFTFloatArray(0.1, 1.0, -10.0)
	expectedErr := "cannot take log of a negative number"
	_, err := xs.Log()
	assert.Error(t, err, "expected error not thrown")
	assert.Equal(t, expectedErr, err.Error())
}

func TestFloat_Sin(t *testing.T) {
	// 0°, 45°, 60°, 90°
	xs := NewFTFloatArray(0.0, 0.785398163, 1.047197551, 1.570796327)
	expected := NewFTFloatArray(0.0, 0.707106781, 0.866025404, 1.0)
	actual, err := xs.Sin()
	assert.NoError(t, err)
	actualFTFloat, err := asFTFloatArray(actual)
	assert.NoError(t, err)
	assert.InDeltaSlice(t, expected.array, actualFTFloat.array, DeltaSliceAccuracy)
}

func TestFloat_Cos(t *testing.T) {
	// 0°, 45°, 60°, 90°
	xs := NewFTFloatArray(0.0, 0.785398163, 1.047197551, 1.570796327)
	expected := NewFTFloatArray(1.0, 0.707106781, 0.5, 0.0)
	actual, err := xs.Cos()
	assert.NoError(t, err)
	actualFTFloat, err := asFTFloatArray(actual)
	assert.NoError(t, err)
	assert.InDeltaSlice(t, expected.array, actualFTFloat.array, DeltaSliceAccuracy)

}
func TestEd25519IntToFromBytes(t *testing.T) {
	a := NewFTEd25519IntArrayFromInt64s(25, 30, 50, 60, 70, 85, 90, 100)
	xs, err := a.GetBinaryArray(0)
	if err != nil {
		t.Error(err)
	}
	b, err := NewFTEd25519IntArrayFromBytes(xs)
	if err != nil {
		t.Error(err)
	}
	if len(a.Values()) != len(b.Values()) {
		t.Error("wrong length")
	}
	if !a.Equals(b) {
		t.Error("Not equal")
	}
}

func TestConvertToAndFromBytearray(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)

	bs, err := xs.GetBinaryArray(0)
	if err != nil {
		t.Fail()
	}

	result, err := NewFTBytearrayArrayFromBytes(bs, 4)
	assert.NoError(t, err)

	if !result.Equals(xs) {
		t.Fail()
	}
}

func TestConvertToAndFromByteArrayWithDifferentLength(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23},
	)

	bs, err := xs.GetBinaryArray(0)
	if err != nil {
		t.Fail()
	}

	result, err := NewFTBytearrayArrayFromBytes(bs, 8)
	assert.NoError(t, err)
	if !result.Equals(xs) {
		t.Fail()
	}
}

func TestByteArrayArray_DebugString(t *testing.T) {
	// Returns the type with its length and memory usage
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23},
	)
	expectedDebugString := "BytearrayArray(Length=3,Memory=24 B)"
	actualDebugString := xs.DebugString()
	assert.Equal(t, expectedDebugString, actualDebugString)
}

func TestByteArrayArray_PythonString(t *testing.T) {
	// Returns "bytearraylist <space delimited list of elements>""
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23},
	)
	expectedPythonString := "bytearraylist 0 1 2 3 4 5 6 7,8 9 10 11 12 13 14 15,16 17 18 19 20 21 22 23"
	actualPythonString := xs.PythonString()
	assert.Equal(t, expectedPythonString, actualPythonString)
}

func TestByteArrayArray_AsType_ToByteArrayArray_Success(t *testing.T) {
	// Converts to a bytearrayarray of larger width
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	expected := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 0, 0, 0, 0},
		[]byte{4, 5, 6, 7, 0, 0, 0, 0},
		[]byte{8, 9, 10, 11, 0, 0, 0, 0},
		[]byte{12, 13, 14, 15, 0, 0, 0, 0},
		[]byte{16, 17, 18, 19, 0, 0, 0, 0},
		[]byte{20, 21, 22, 23, 0, 0, 0, 0},
	)

	actual, err := xs.AsType("b8")
	assert.Equal(t, expected, actual)
	assert.NoError(t, err)
}

func TestByteArrayArray_AsType_ToByteArrayArray_Failure(t *testing.T) {
	// Throws Error - cannot convert to bytearrayarray of narrower width
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 0, 0, 0, 0},
		[]byte{4, 5, 6, 7, 0, 0, 0, 0},
		[]byte{8, 9, 10, 11, 0, 0, 0, 0},
		[]byte{12, 13, 14, 15, 0, 0, 0, 0},
		[]byte{16, 17, 18, 19, 0, 0, 0, 0},
		[]byte{20, 21, 22, 23, 0, 0, 0, 0},
	)
	expectedError := "can not convert to a smaller bytearray: b8 -> b4"
	_, err := xs.AsType("b4")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err.Error())

}
func TestByteArrayArray_AsType_ToIntArray_Success(t *testing.T) {
	// Converts to int array
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 2, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 3, 0, 0, 0, 0, 0},
	)
	expected := NewFTIntegerArray(1, 512, 196608)

	actual, err := xs.AsType("i")
	assert.Equal(t, expected, actual)
	assert.NoError(t, err)
}

func TestByteArrayArray_AsType_ToIntArray_Failure(t *testing.T) {
	// Throws error - cannot convert to int array if not b8.
	xs := NewFTBytearrayArrayOrPanic(
		9,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 2, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 3, 0, 0, 0, 0, 0, 0},
	)

	_, err := xs.AsType("i")
	expectedError := "can only convert from b8 to integer"
	assert.Error(t, err)
	assert.Equal(t, expectedError, err.Error())
}

func TestByteArrayArray_AsType_ToFloat_Failure(t *testing.T) {
	// Throws error - cannot convert to float.
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 2, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 3, 0, 0, 0, 0, 0},
	)
	expectedError := "conversion not supported: b8 -> f"
	_, err := xs.AsType("f")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err.Error())
}

func TestByteArrayArray_AsType_ToEd25519Int_Success(t *testing.T) {
	// Converts to Ed25519IntArray
	xs := NewFTBytearrayArrayOrPanic(
		32,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	)

	expected := NewFTEd25519IntArrayFromInt64s(1, 2)

	actual, err := xs.AsType("I")
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_AsType_ToEd25519Int_Failure(t *testing.T) {
	// Throws error - can only convert if bytearrayarray is 32 elements wide.
	xs := NewFTBytearrayArrayOrPanic(
		31,
		[]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	)
	expectedError := "can only convert from b32 to Ed25519Int"
	_, err := xs.AsType("I")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err.Error())
}

func TestByteArrayArray_SetLength_Larger(t *testing.T) {
	// Sets the length of the bytearrayarray, appending zero value elements
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23},
	)
	lengthToSet := int64(5)
	err := xs.SetLength(int64(lengthToSet))
	expected := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0},
		[]byte{0, 0, 0, 0, 0, 0, 0, 0},
	)

	assert.NoError(t, err)
	assert.Equal(t, lengthToSet, xs.Length())
	assert.Equal(t, expected, xs)
}

func TestByteArrayArray_SetLength_Smaller(t *testing.T) {
	// Sets the length of the bytearrayarray and truncates
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23},
	)
	lengthToSet := int64(2)
	err := xs.SetLength(int64(lengthToSet))

	expected := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
	)

	assert.NoError(t, err)
	assert.Equal(t, lengthToSet, xs.Length())
	assert.Equal(t, expected, xs)
}

func TestByteArrayArray_Remove(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	indexesToRemove := NewFTIntegerArray(2, 3)
	expectedLength := int64(4)

	err := xs.Remove(indexesToRemove)
	assert.NoError(t, err)
	assert.Equal(t, expectedLength, xs.Length())
}

func TestByteArrayArray_Broadcast_Success(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
	)
	expectedLength := int64(10)
	err := xs.Broadcast(expectedLength)

	assert.NoError(t, err)
	assert.Equal(t, expectedLength, xs.Length())
	assert.Equal(t, xs.array[0], xs.array[1])
	assert.Equal(t, xs.array[0], xs.array[len(xs.array)-1])
}

func TestByteArrayArray_Broadcast_Failure(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
	)
	expectedLength := int64(10)
	expectedError := "cannot broadcast array of size 2"
	err := xs.Broadcast(expectedLength)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err.Error())
}

func TestByteArrayArray_Lookup(t *testing.T) {
	// Returns a bytearrayarray with the elements from original array at specified indexes.
	// If index out of range then returns the specified default value instead of that element.
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	defaultValue := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{1, 1, 1, 1},
	)
	indexesToLookup := NewFTIntegerArray(2, 3, 27)
	expected := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{1, 1, 1, 1},
	)
	actual, err := xs.Lookup(indexesToLookup, defaultValue)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Get_NoIndex(t *testing.T) {
	// Should return a copy of xs.
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	actual, err := xs.Get(nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, xs, actual)
}

func TestByteArrayArray_Get_NoDefault(t *testing.T) {
	// Should return an array with 2 elements at index 2,3.
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	indexes := NewFTIntegerArray(2, 3)
	expected := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
	)
	actual, err := xs.Get(indexes, nil)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Get_IndexAndDefault(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	indexes := NewFTIntegerArray(2, 3, 27)
	defaultValue := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{9, 9, 9, 9},
	)
	expected := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{9, 9, 9, 9},
	)
	actual, err := xs.Get(indexes, defaultValue)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Get_IndexOutOfRange_NoDefault(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	indexes := NewFTIntegerArray(2, 3, 27)
	expectedError := "out of range: 27"
	_, err := xs.Get(indexes, nil)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err.Error())
}

func TestByteArrayArray_Set_Success(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	indexes := NewFTIntegerArray(2, 3)
	values := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{1, 1, 1, 1},
		[]byte{2, 2, 2, 2},
	)
	err := xs.Set(indexes, values)
	expected := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{1, 1, 1, 1},
		[]byte{2, 2, 2, 2},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestByteArrayArray_Set_Failure(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	indexes := NewFTIntegerArray(2, 3)
	values := NewFTIntegerArray(1, 2, 3)
	expectedError := "values is not an BytearrayArray"
	err := xs.Set(indexes, values)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err.Error())
}

func TestByteArrayArray_Index(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	expected := NewFTIntegerArray(0, 1, 2, 3, 4, 5)
	actual := xs.Index()
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Contains(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
		[]byte{20, 21, 22, 23},
	)
	values := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{4, 5, 6, 7},
		[]byte{99, 99, 99, 99},
		[]byte{0, 1, 2, 3},
		[]byte{3, 1, 2, 3},
	)
	actual, err := xs.Contains(values)
	expected := NewFTIntegerArray(1, 0, 1, 0)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_ReduceSum(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)
	indexes := NewFTIntegerArray(2, 2, 3)
	values := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{3, 5, 7, 6},
		[]byte{1, 1, 2, 3},
		[]byte{4, 4, 4, 4},
	)
	expected := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{3, 5, 7, 7},
		[]byte{4, 4, 4, 4},
		[]byte{16, 17, 18, 19},
	)
	err := xs.ReduceSum(indexes, values)

	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestByteArrayArray_ReduceISum(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)
	indexes := NewFTIntegerArray(2, 2, 3)
	values := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{3, 5, 7, 6},
		[]byte{1, 1, 2, 3},
		[]byte{4, 4, 4, 4},
	)
	expected := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{11, 13, 15, 15},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)
	err := xs.ReduceISum(indexes, values)

	assert.NoError(t, err)
	assert.Equal(t, expected, xs)
}

func TestByteArrayArray_Cumsum(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{1, 10, 1, 1},
		[]byte{1, 11, 2, 2},
		[]byte{0, 15, 3, 4},
		[]byte{1, 17, 4, 8},
	)

	expected := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{1, 10, 1, 1},
		[]byte{1, 11, 3, 3},
		[]byte{1, 15, 3, 7},
		[]byte{1, 31, 7, 15},
	)
	actual, err := xs.CumSum()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Mux(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic( // if true
		4,
		[]byte{12, 0, 7, 5},
		[]byte{13, 1, 6, 4},
		[]byte{4, 5, 7, 9},
		[]byte{15, 7, 4, 3},
	)
	condition := NewFTIntegerArray(1, 0, 1, 2)
	falseVal := NewFTBytearrayArrayOrPanic( // if false
		4,
		[]byte{1, 10, 1, 1},
		[]byte{1, 11, 2, 2},
		[]byte{0, 15, 3, 4},
		[]byte{1, 17, 4, 8},
	)
	expected := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{12, 0, 7, 5},
		[]byte{1, 11, 2, 2},
		[]byte{4, 5, 7, 9},
		[]byte{15, 7, 4, 3},
	)
	actual, err := xs.Mux(condition, falseVal)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Concat_DifferentLengths(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
	)
	xt := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)
	expected := "arrays have different lengths"
	_, err := xs.Concat(xt)
	assert.Error(t, err)
	assert.Equal(t, expected, err.Error())
}

func TestByteArrayArray_Concat_DifferentWidths(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
	)
	xt := NewFTBytearrayArrayOrPanic(
		5,
		[]byte{8, 9, 10, 11, 12},
		[]byte{13, 14, 15, 16, 17},
	)
	expected := NewFTBytearrayArrayOrPanic(
		9,
		[]byte{0, 1, 2, 3, 8, 9, 10, 11, 12},
		[]byte{4, 5, 6, 7, 13, 14, 15, 16, 17},
	)
	actual, err := xs.Concat(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Concat_Success(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
	)
	xt := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
	)
	expected := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 8, 9, 10, 11},
		[]byte{4, 5, 6, 7, 12, 13, 14, 15},
	)

	actual, err := xs.Concat(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Project_CompleteMap(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)
	// 1 -> 0
	// 0 -> 2
	// 2 -> 1
	index := NewFTIntegerArray(0, 2, 1)
	values := NewFTIntegerArray(1, 0, 2)
	expected := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{1, 2, 0},
		[]byte{5, 6, 4},
		[]byte{9, 10, 8},
		[]byte{13, 14, 12},
		[]byte{17, 18, 16},
	)
	actual, err := xs.Project(int64(3), index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Project_PartialMap(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)
	// 1 -> 0
	// 0 -> 2
	index := NewFTIntegerArray(0, 2)
	values := NewFTIntegerArray(1, 0)
	expected := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{1, 0, 0},
		[]byte{5, 0, 4},
		[]byte{9, 0, 8},
		[]byte{13, 0, 12},
		[]byte{17, 0, 16},
	)
	actual, err := xs.Project(int64(3), index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Eq(t *testing.T) {
	b1 := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)
	b2 := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 99, 11},
		[]byte{12, 13, 14, 15},
		[]byte{0, 0, 1, 1},
	)
	expected := NewFTIntegerArray(1, 1, 0, 1, 0)
	actual, err := b1.Eq(b2)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Ne(t *testing.T) {
	b1 := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 10, 11},
		[]byte{12, 13, 14, 15},
		[]byte{16, 17, 18, 19},
	)
	b2 := NewFTBytearrayArrayOrPanic(
		4,
		[]byte{0, 1, 2, 3},
		[]byte{4, 5, 6, 7},
		[]byte{8, 9, 99, 11},
		[]byte{12, 13, 14, 15},
		[]byte{0, 0, 1, 1},
	)
	expected := NewFTIntegerArray(0, 0, 1, 0, 1)
	actual, err := b1.Ne(b2)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_LShift(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0000_0000, 0b0000_1101},
		[]byte{0b0000_0000, 0b0000_1101, 0b0000_1101},
		[]byte{0b1100_0000, 0b1100_1101, 0b1100_1101},
	)
	shift := NewFTIntegerArray(10, 9, 2)
	expected := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
		[]byte{0b0000_0011, 0b0011_0111, 0b0011_0100},
	)
	actual, err := xs.LShift(shift)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_RShift(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b1111_0000},
	)
	shift := NewFTIntegerArray(10, 9, 5)
	expected := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0000_0000, 0b0000_1101},
		[]byte{0b0000_0000, 0b0000_1101, 0b0000_1101},
		[]byte{0b0000_0000, 0b1101_0000, 0b1101_0111},
	)
	actual, err := xs.RShift(shift)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_And(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100},
	)
	xt := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0011_0100},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
	)
	expected := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
	)
	actual, err := xs.And(xt)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Or(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100},
	)
	xt := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0011_0100},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
	)
	expected := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0011_0100},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100},
	)
	actual, err := xs.Or(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Xor(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100},
	)
	xt := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0011_0100},
		[]byte{0b0001_1010, 0b0001_1010, 0b0000_0000},
	)
	expected := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0000_0000, 0b0011_0100},
		[]byte{0b0000_0000, 0b0000_0000, 0b0011_0100},
	)
	actual, err := xs.Xor(xt)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestByteArrayArray_Invert(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b0000_0000, 0b0011_0100, 0b0000_0000},
		[]byte{0b0001_1010, 0b0001_1010, 0b0011_0100},
	)
	expected := NewFTBytearrayArrayOrPanic(
		3,
		[]byte{0b1111_1111, 0b11001011, 0b1111_1111},
		[]byte{0b1110_0101, 0b1110_0101, 0b1100_1011},
	)
	actual, err := xs.Invert()
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestExtractSingleton_IntegerArray(t *testing.T) {
	xs := NewFTIntegerArray(5)
	x, err := xs.Single()

	assert.NoError(t, err)
	if x != 5 {
		t.Error("x should have been 5")
	}
}

func TestCopy_Integer(t *testing.T) {
	xs := NewFTIntegerArray(1, 2, 3, 4, 5)
	ys, err := xs.Clone()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(xs.Values(), ys.(*FTIntegerArray).Values()) {
		t.Error("Copy should be copying all elements")
	}
}

func TestCopy_Bytearray(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23},
	)
	ys, err := xs.Clone()
	if err != nil {
		t.Fatal(err)
	}
	yys := ys.(*FTBytearrayArray)

	for i := range xs.Values() {
		if &xs.Values()[i] == &yys.Values()[i] {
			t.Error("Copy should be performing a deep copy")
		}

		if !slices.Equal(xs.Values()[i], yys.Values()[i]) {
			t.Error("Copy should be copying all elements")
		}

	}
}
func TestGet_Integer(t *testing.T) {
	def := NewFTIntegerArray(8)

	xs := NewFTIntegerArray(1, 2, 3, 4, 5)
	ys := NewFTIntegerArray(2, 4, 8)

	zs, err := xs.Get(NewFTIntegerArray(1, 3, 99), def)
	if err != nil {
		t.Error(err)
	}
	if !ys.Equals(zs) {
		t.Error("Get did not get correct values")
	}
}

func TestGet_ByteArray(t *testing.T) {
	xs := NewFTBytearrayArrayOrPanic(
		8,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7},
		[]byte{8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23},
	)

	ys, err := xs.Get(NewFTIntegerArray(1), nil)
	assert.NoError(t, err)

	bs := ys.(*FTBytearrayArray).Values()[0]

	if !slices.Equal(bs, []byte{8, 9, 10, 11, 12, 13, 14, 15}) {
		t.Error("byte arrays were not equal")
	}

	if &bs == &xs.Values()[0] {
		t.Error("Get should return a copy of the array")
	}
}

func TestByteArrayArray_Project_EmptyArray(t *testing.T) {
	xs, err := NewFTBytearrayArrayFromBytes(nil, 8)
	assert.NoError(t, err)
	index := NewFTIntegerArray(0, 2)
	values := NewFTIntegerArray(1, 0)
	expected, err := NewFTBytearrayArrayFromBytes(nil, 3)
	assert.NoError(t, err)
	actual, err := xs.Project(3, index, values)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestShiftLeft_ByteArray(t *testing.T) {
	xs := []byte{8, 9, 10, 11, 12, 13, 14, 15}
	ys := ShiftLeft(xs, 5)

	assert.Equal(t, ys, []byte{1, 33, 65, 97, 129, 161, 193, 224})
}

func TestShiftRight_ByteArray(t *testing.T) {
	xs := []byte{8, 9, 10, 11, 12, 13, 14, 15}
	ys := ShiftLeft(xs, -5) // Negative value indicates right shift

	assert.Equal(t, ys, []byte{0, 64, 72, 80, 88, 96, 104, 112})
}
