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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func TestNewListMapFromArrays(t *testing.T) {
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

	keys1 := []ArrayElementTypeVal{
		ints,
		floats,
		bytearrays,
		ed25519Ints,
	}

	lm1, err := NewListMapFromArrays(typecode, keys1, "any")
	assert.Empty(t, err, "error should be nil")
	assert.Equal(t, getListMapValSum(lm1), 3, "invalid sum of lm values")

	lm2, err := NewListMapFromArrays(typecode, keys1, "pos")
	assert.Empty(t, err, "error should be nil")
	assert.Equal(t, getListMapValSum(lm2), 3, "invalid sum of lm values")

	lm3, err := NewListMapFromArrays(typecode, keys1, "rnd")
	assert.Empty(t, err, "error should be nil")
	assert.Equal(t, getListMapValSum(lm3), 3, "invalid sum of lm values")

	lm4, err := NewListMapFromArrays(typecode, keys1, "foo")
	assert.Empty(t, lm4, "lm should be nil")
	assert.Error(t, err, "error should be returned")

	keys2 := []ArrayElementTypeVal{NewFTIntegerArray(1, 1)}
	lm5, err := NewListMapFromArrays(typecode, keys2, "pos")
	assert.Empty(t, lm5, "lm should be nil")
	assert.Error(t, err, "error should be returned")
}

func TestKeysToArrayTypeVals(t *testing.T) {
	// Test empty key slice
	keys1 := make([]Key, 0)
	_, err := KeysToArrayTypeVals(keys1, []TypeCode{"i"})
	assert.NoError(t, err, "error should be nil")

	// Test several key elements
	keys2 := NewKey([]interface{}{int64(1), 1.4, int64(4), 3.2})
	keys3 := NewKey([]interface{}{int64(2), 5.4, int64(3), 4.2})
	arrVals, err := KeysToArrayTypeVals([]Key{keys2, keys3}, []TypeCode{"i", "f", "i", "f"})
	assert.NoError(t, err, "error should be nil")
	assert.Len(t, arrVals, 4, "invalid slice width")
}

func TestListmapGetItems(t *testing.T) {
	lm, _ := BuildListMap()
	lookupKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2, 3),
		NewFTFloatArray(5.5, 6.6),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{4, 5, 6},
			[]byte{7, 8, 9},
		),
		NewFTEd25519IntArrayFromInt64s(5, 5),
	}

	values, err := lm.GetItems(lookupKeys, nil)
	assert.NoError(t, err, "error should be nil")
	assert.True(t, slices.Equal(values.Values(), []int64{1, 2}), "did not get expected values back")
}

func TestListmapGetItemsWithDefault(t *testing.T) {
	lm, _ := BuildListMap()
	lookupKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2, 10),
		NewFTFloatArray(5.5, 6.6),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{4, 5, 6},
			[]byte{7, 8, 9},
		),
		NewFTEd25519IntArrayFromInt64s(5, 5),
	}

	values, err := lm.GetItems(lookupKeys, int64(99))
	assert.NoError(t, err, "error should be nil")
	assert.True(t, slices.Equal(values.Values(), []int64{1, 99}), "did not get expected values back")
}

func TestListMapGetItems_FromEmptyListmap_WithDefaults(t *testing.T) {
	lm, _ := BuildListMapNoEncryptionEmpty()
	lookupKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2, 10),
		NewFTFloatArray(5.5, 6.6),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{4, 5, 6},
			[]byte{7, 8, 9},
		),
	}

	values, err := lm.GetItems(lookupKeys, int64(99))
	assert.NoError(t, err, "error calling listmap getitems")
	assert.Equal(t, NewFTIntegerArray(99, 99), values, "getitems returned unexpected values")
}

func TestListMapGetItems_FromEmptyListmap_WithNoDefault(t *testing.T) {
	lm, _ := BuildListMapNoEncryptionEmpty()
	lookupKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2, 10),
		NewFTFloatArray(5.5, 6.6),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{4, 5, 6},
			[]byte{7, 8, 9},
		),
	}

	_, err := lm.GetItems(lookupKeys, nil)
	expectedError := "key: {[2 5.5 [4 5 6]]} not found in listmap"
	assert.Error(t, err, "error calling listmap getitems")
	assert.Equal(t, expectedError, err.Error())
}

func TestListmapRemoveItems(t *testing.T) {
	lm, keys := BuildListMap()

	removeKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2),
		NewFTFloatArray(5.5),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{4, 5, 6},
		),
		NewFTEd25519IntArrayFromInt64s(5),
	}

	ints := keys[0].(*FTIntegerArray)
	floats := keys[1].(*FTFloatArray)
	bytearrays := keys[2].(*FTBytearrayArray)

	movedKeys, oldValues, newValues, err := lm.RemoveItems(removeKeys, false)
	assert.NoError(t, err, "error calling listmap removeitems")
	moveKey := movedKeys[0].ToSlice()

	assert.Equal(t, ints.Values()[2], moveKey[0], "did not get expected values back")
	assert.Equal(t, floats.Values()[2], moveKey[1], "did not get expected values back")
	assert.Equal(t, bytearrays.Values()[2], moveKey[2], "did not get expected values back")
	assert.True(t, slices.Equal(oldValues.Values(), []int64{2}), "did not get expected values back")
	assert.True(t, slices.Equal(newValues.Values(), []int64{1}), "did not get expected values back")
	assert.Equal(t, 2, len(lm.m), "did not get expected length of lm back")
}

func TestListmapRemoveItems_FromEnd(t *testing.T) {
	lm, _ := BuildListMap()

	removeKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2, 3),
		NewFTFloatArray(5.5, 6.6),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{4, 5, 6},
			[]byte{7, 8, 9},
		),
		NewFTEd25519IntArrayFromInt64s(5, 5),
	}

	movedKeys, oldValues, newValues, _ := lm.RemoveItems(removeKeys, false)

	assert.Equal(t, 0, len(movedKeys), "did not get expected value back for length")
	assert.Equal(t, int64(0), oldValues.Length(), "did not get expected values back")
	assert.Equal(t, int64(0), newValues.Length(), "did not get expected values back")

	assert.Equal(t, 1, len(lm.m), "did not get expected length of lm back")
}

func TestListmapRemoveItems_FromMiddle(t *testing.T) {
	ints := NewFTIntegerArray(1, 2, 3, 4)
	floats := NewFTFloatArray(1.1, 2.2, 3.3, 4.4)

	typecode := []TypeCode{"i", "f"}

	keys := []ArrayElementTypeVal{
		ints,
		floats,
	}
	lm, err := NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err)

	removeKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2, 3),
		NewFTFloatArray(2.2, 3.3),
	}

	movedKeys, oldValues, newValues, _ := lm.RemoveItems(removeKeys, false)
	moveKey := movedKeys[0].ToSlice()

	assert.Equal(t, ints.Values()[3], moveKey[0], "did not get expected values back")
	assert.Equal(t, floats.Values()[3], moveKey[1], "did not get expected values back")
	assert.Equal(t, []int64{3}, oldValues.Values(), "did not get expected values back")
	assert.Equal(t, []int64{1}, newValues.Values(), "did not get expected values back")
	assert.Equal(t, 2, len(lm.m), "did not get expected length of lm back")
}

func TestListmapRemoveItems_NoMovedKeys(t *testing.T) {
	ints := NewFTIntegerArray(1, 2, 3)
	floats := NewFTFloatArray(1.1, 2.2, 3.3)

	typecode := []TypeCode{"i", "f"}

	keys := []ArrayElementTypeVal{
		ints,
		floats,
	}
	lm, err := NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err)

	removeKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2, 3),
		NewFTFloatArray(2.2, 3.3),
	}

	keys2 := []ArrayElementTypeVal{
		NewFTIntegerArray(1),
		NewFTFloatArray(1.1),
	}

	expectedLm, err := NewListMapFromArrays(typecode, keys2, "any")
	assert.NoError(t, err, "error in NewListMapFromArrays")

	movedKeys, oldValues, newValues, _ := lm.RemoveItems(removeKeys, false)
	assert.Equal(t, 0, len(movedKeys), "did not get expected values back")
	assert.Equal(t, int64(0), oldValues.Length(), "did not get expected values back")
	assert.Equal(t, int64(0), newValues.Length(), "did not get expected values back")
	assert.Equal(t, expectedLm, lm, "did not get expected values back for listmap")
}

func TestListmapRemoveItems_FromEmptyListmap(t *testing.T) {
	lm, _ := BuildListMapNoEncryptionEmpty()

	removeKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2, 3),
		NewFTFloatArray(2.2, 3.3),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{1, 2, 3},
			[]byte{4, 5, 6}),
	}

	movedKeys, oldValues, newValues, _ := lm.RemoveItems(removeKeys, false)
	assert.Equal(t, 0, len(movedKeys))
	assert.True(t, oldValues == nil)
	assert.True(t, newValues == nil)
}

func TestListmapAddItems(t *testing.T) {
	lm, _ := BuildListMap()

	addKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(4),
		NewFTFloatArray(-3.6),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{1, 2, 3},
		),
		NewFTEd25519IntArrayFromInt64s(5, 5),
	}

	newKeys, newValues, _ := lm.AddItems(addKeys, false)

	assert.Equal(t, int64(4), newKeys[0].(*FTIntegerArray).Values()[0], "did not get expected values back")
	assert.Equal(t, -3.6, newKeys[1].(*FTFloatArray).Values()[0], "did not get expected values back")
	assert.True(t, reflect.DeepEqual(newKeys[2].(*FTBytearrayArray).Values()[0], []byte{1, 2, 3}), "did not get expected values back")
	assert.True(t, slices.Equal(newValues.Values(), []int64{3}), "did not get expected values back")
	assert.Equal(t, 4, len(lm.m), "did not get expected length of lm back")
}

func TestListmapAddItems_ListmapEmpty(t *testing.T) {
	lm, _ := BuildListMapNoEncryptionEmpty()

	addKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(4),
		NewFTFloatArray(-3.6),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{1, 2, 3},
		),
	}

	newKeys, newValues, err := lm.AddItems(addKeys, false)

	assert.NoError(t, err)
	assert.Equal(t, int64(4), newKeys[0].(*FTIntegerArray).Values()[0], "did not get expected values back")
	assert.Equal(t, -3.6, newKeys[1].(*FTFloatArray).Values()[0], "did not get expected values back")
	assert.True(t, reflect.DeepEqual(newKeys[2].(*FTBytearrayArray).Values()[0], []byte{1, 2, 3}), "did not get expected values back")
	assert.True(t, slices.Equal(newValues.Values(), []int64{0}), "did not get expected values back")
	assert.Equal(t, 1, len(lm.m), "did not get expected length of lm back")
}

func TestListmapIntersectItems(t *testing.T) {
	lm1, keys1 := BuildListMap()

	ints := NewFTIntegerArray(4, 2, 3)
	keys1[0] = ints
	typecode := []TypeCode{"i", "f", "b3", "I"}
	lm2, err := NewListMapFromArrays(typecode, keys1, "any")
	assert.Empty(t, err, "error should be nil")

	keys2 := lm2.GetKeys(false)
	arrVals, _ := KeysToArrayTypeVals(keys2, lm2.tc)
	res, _ := lm1.IntersectItems(arrVals)
	assert.Equal(t, int64(2), res[0].Length(), "did not get expected values back")
	assert.True(t, slices.Equal(res[0].(*FTIntegerArray).Values(), []int64{2, 3}))
	assert.True(t, slices.Equal(res[1].(*FTFloatArray).Values(), []float64{5.5, 6.6}))
	assert.True(t, reflect.DeepEqual(res[2].(*FTBytearrayArray).Values()[0], []byte{4, 5, 6}))
	assert.True(t, reflect.DeepEqual(res[2].(*FTBytearrayArray).Values()[1], []byte{7, 8, 9}))
}

func TestListmapIntersectItems_NoIntersection(t *testing.T) {
	ints := NewFTIntegerArray(1, 2, 3)
	floats := NewFTFloatArray(1.1, 2.2, 3.3)
	keys := []ArrayElementTypeVal{
		ints,
		floats,
	}

	typecode := []TypeCode{"i", "f"}
	lm, err := NewListMapFromArrays(typecode, keys, "any")
	assert.NoError(t, err)

	ints2 := NewFTIntegerArray(4, 5, 6)
	floats2 := NewFTFloatArray(4.4, 5.5, 6.6)
	keys2 := []ArrayElementTypeVal{
		ints2,
		floats2,
	}

	result, err := lm.IntersectItems(keys2)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result[0].Length(), "did not get expected values back for length")
	assert.True(t, slices.Equal(result[0].(*FTIntegerArray).Values(), []int64{}), "did not get expected values back for FTIntegerArray")
	assert.True(t, slices.Equal(result[1].(*FTFloatArray).Values(), []float64{}), "did not get expected values back for FTFloatArray")
}

func TestListmapIntersectItems_Empty_NoIntersection(t *testing.T) {

	lm, _ := BuildListMapNoEncryptionEmpty()

	keys2 := []ArrayElementTypeVal{
		NewFTIntegerArray(4, 5, 6),
		NewFTFloatArray(4.4, 5.5, 6.6),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{1, 2, 3},
		),
	}

	result, err := lm.IntersectItems(keys2)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result[0].Length(), "did not get expected values back")
	assert.True(t, slices.Equal(result[0].(*FTIntegerArray).Values(), []int64{}), "did not get expected values back for FTIntegerArray")
	assert.True(t, slices.Equal(result[1].(*FTFloatArray).Values(), []float64{}), "did not get expected values back for FTFloatArray")
	assert.Equal(t, 0, len(result[2].(*FTBytearrayArray).Values()))
}

func TestListmapContainItems(t *testing.T) {
	lm, keys := BuildListMap()

	newKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(keys[0].(*FTIntegerArray).Values()[0:2]...),
		NewFTFloatArray(keys[1].(*FTFloatArray).Values()[0:2]...),
		NewFTBytearrayArrayOrPanic(3, keys[2].(*FTBytearrayArray).Values()[0:2]...),
		NewFTEd25519IntArray(keys[3].(*FTEd25519IntArray).Values()[0:2]...),
	}

	res := lm.Contains(newKeys)

	assert.True(t, slices.Equal(res.Values(), []int64{1, 1}), "did not get expected values back")
	assert.Equal(t, int64(2), res.Length(), "did not get expected values back")
}

func TestListmapContainsItems_WhenEmpty(t *testing.T) {
	lm, _ := BuildListMapNoEncryptionEmpty()

	newKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(1),
		NewFTFloatArray(1.1),
		NewFTBytearrayArrayOrPanic(
			3,
			[]byte{1, 2, 3},
		),
	}

	res := lm.Contains(newKeys)
	assert.True(t, slices.Equal(res.Values(), []int64{}), "result array should be empty")

}

func TestListmapContainsItems_OneValueNotContained(t *testing.T) {
	lm, _ := NewListMapFromArrays(
		[]TypeCode{"i", "f"},
		[]ArrayElementTypeVal{
			NewFTIntegerArray(1),
			NewFTFloatArray(1.1),
		},
		"any",
	)

	findKeys := []ArrayElementTypeVal{
		NewFTIntegerArray(2),
		NewFTFloatArray(2.2),
	}
	result := lm.Contains(findKeys)
	assert.Equal(t, 1, len(result.array))
	assert.Equal(t, NewFTIntegerArray(0), result)
}

func TestListmapGetKeys(t *testing.T) {
	lm, keys := BuildListMap()
	getKeys := lm.GetKeys(true)
	res, _ := KeysToArrayTypeVals(getKeys, lm.tc)

	assert.Equal(t, 4, len(res), "did not get expected values back")
	assert.True(t, reflect.DeepEqual(keys, res), "did not get expected values back")
}

func TestListmapGetKeys_EmptyListmap(t *testing.T) {
	t.Skip()
	lm, keys := BuildListMapNoEncryptionEmpty()
	getKeys := lm.GetKeys(true)
	res, _ := KeysToArrayTypeVals(getKeys, lm.tc)

	assert.Equal(t, 3, len(res), "did not get expected values back")
	assert.True(t, reflect.DeepEqual(keys, res), "did not get expected values back")
}

func TestListmapEstimatedSize(t *testing.T) {
	lm, _ := BuildListMap()
	expected := int64(177)
	actual := lm.EstimatedSize()
	assert.Equal(t, expected, actual)
}

func TestListmapGetBinaryArray(t *testing.T) {
	lm, _ := BuildListMap()
	// Test 1
	result0, err := lm.GetBinaryArray(0)
	// int(1,2,3)
	expected0 := []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	assert.NoError(t, err, "error getting binary array 0")
	assert.Equal(t, expected0, result0, "binary array 0 didn't return expected result")
	// Test 2
	result1, err := lm.GetBinaryArray(1)
	// float(4.4,5.5,6.6)
	expected1 := []byte{0x9a, 0x99, 0x99, 0x99, 0x99, 0x99, 0x11, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x16, 0x40, 0x66, 0x66, 0x66, 0x66, 0x66, 0x66, 0x1a, 0x40}
	assert.NoError(t, err, "error getting binary array 1")
	assert.Equal(t, expected1, result1, "binary array 1 didn't return expected result")
	// Test 3
	result2, err := lm.GetBinaryArray(2)
	// []byte{1,2,3}
	// []byte{4,5,6}
	// []byte{7,8,9}
	expected2 := []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9}
	assert.NoError(t, err, "error getting binary array 2")
	assert.Equal(t, expected2, result2)
}

func TestListmapTypeCode(t *testing.T) {
	lm, _ := BuildListMap()
	expected := TypeCode("ifb3I")
	actual := lm.TypeCode()
	assert.Equal(t, expected, actual)
}

func TestListmapName(t *testing.T) {
	lm, _ := BuildListMap()
	expected := "ListMap(ifb3I)"
	actual := lm.Name()
	assert.Equal(t, expected, actual)
}

func TestListmapDebugString(t *testing.T) {
	lm, _ := BuildListMap()
	expected := "ifb3I(Typecode=ListMap(ifb3I),Width=4,Length=3,Memory=177 B)"
	actual := lm.DebugString()
	assert.Equal(t, expected, actual)
}

func TestKeyAt(t *testing.T) {
	ints := NewFTIntegerArray(1, 2, 3, 4)
	floats := NewFTFloatArray(1.1, 2.2, 3.3, 4.4)
	keys := make([]ArrayElementTypeVal, 2)
	keys[0] = ints
	keys[1] = floats

	expected := make([]interface{}, 2)
	expected[0] = int64(1)
	expected[1] = 1.1
	expectedKey := NewKey(expected)
	actualKey := KeyAt(keys, int64(0))
	assert.Equal(t, expectedKey, actualKey)

	expected[0] = int64(2)
	expected[1] = 2.2
	expectedKey = NewKey(expected)
	actualKey = KeyAt(keys, int64(1))
	assert.Equal(t, expectedKey, actualKey)
}

func TestCreateKey(t *testing.T) {
	// k := [3]interface{}{1, 2, 3}
	// fmt.Printf("%v", k)

	// m := make(map[[3]interface{}]int64)
	// m[k] = 0

	// fmt.Printf("%v", m)
	// t.Fail()

	key1 := make([]interface{}, 4)
	key1[0] = 1
	key1[1] = 1.4
	key1[2] = 4
	key1[3] = 3.2

	key2 := make([]interface{}, 4)
	key2[0] = 2
	key2[1] = 5.4
	key2[2] = 3
	key2[3] = 4.2

	k1 := NewKey(key1)
	k2 := NewKey(key2)

	m := make(map[Key]int64)
	m[k1] = 0
	m[k2] = 1
}

func BenchmarkNewKey(b *testing.B) {

	keyCount := 125

	for n := 0; n < b.N; n++ {

		m := make(map[Key]int64)
		v := int64(0)

		for i := 0; i < keyCount; i++ {
			for j := 0; j < keyCount; j++ {
				for k := 0; k < keyCount; k++ {

					components := make([]interface{}, 3)
					components[0] = i
					components[1] = j
					bs := make([]byte, 32)
					bs[0] = byte(k)
					components[2] = bs

					key := NewKey(components)

					m[key] = v
					v++
				}
			}
		}
		b.ReportMetric(float64(v), "elements")
	}
}

func BenchmarkGetItems(b *testing.B) {
	buildBenchmark(b, "get", 20000000)
}
func BenchmarkAddItems(b *testing.B) {
	buildBenchmark(b, "add", 20000000)
}
func BenchmarkRemoveItems(b *testing.B) {
	buildBenchmark(b, "remove", 20000000)
}

func buildBenchmark(b *testing.B, testType string, lmSize int) {
	count := lmSize

	b.StopTimer()

	as := make([]int64, count)
	bs := make([]float64, count)
	cs := make([]int64, count)

	asLookup := make([]int64, 0)
	bsLookup := make([]float64, 0)
	csLookup := make([]int64, 0)

	for i := 0; i < count; i++ {
		as[i] = int64(i)
		bs[i] = float64(i)
		cs[i] = int64(i)
		if i%5 == 0 {
			asLookup = append(asLookup, int64(i))
			bsLookup = append(bsLookup, float64(i))
			csLookup = append(csLookup, int64(i))
		}
	}

	b.ReportMetric(float64(count), "elements")
	b.ReportMetric(float64(len(bsLookup)), "removes")

	typecode := []TypeCode{"i", "f", "i"}

	keys := []ArrayElementTypeVal{
		NewFTIntegerArray(as...),
		NewFTFloatArray(bs...),
		NewFTIntegerArray(cs...),
	}

	lm, err := NewListMapFromArrays(typecode, keys, "any")
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()

	for n := 0; n < b.N; n++ {
		items := []ArrayElementTypeVal{
			NewFTIntegerArray(asLookup...),
			NewFTFloatArray(bsLookup...),
			NewFTIntegerArray(csLookup...),
		}

		switch testType {
		case "get":
			_, _ = lm.GetItems(items, nil)
		case "add":
			_, _, _ = lm.AddItems(items, false)
		case "remove":
			_, _, _, _ = lm.RemoveItems(items, false)
		}
	}
}

func getListMapValSum(m *ListMap) int {
	v := 0
	for _, value := range m.m {
		v += int(value)
	}
	return v
}
