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
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"filippo.io/edwards25519"
	"golang.org/x/exp/slices"
)

type FTBytearrayArray struct {
	array [][]byte
	width int64
}

func (v *FTBytearrayArray) Width() int64       { return v.width }
func (v *FTBytearrayArray) TypeCode() TypeCode { return Bytearray(int(v.width)) }
func (v *FTBytearrayArray) Equals(other TypeVal) bool {
	if os, ok := other.(*FTBytearrayArray); ok {
		if v.Length() == os.Length() {
			for i, t := range v.array {
				o := os.array[i]
				if !slices.Equal(t, o) {
					return false
				}
			}
			return true
		}
	}
	return false
}
func (v *FTBytearrayArray) Name() string { return "BytearrayArray" }

func (v *FTBytearrayArray) GetBinaryArray(index int) ([]byte, error) {
	length := len(v.array)
	arraylength := v.TypeCode().Length()

	data := make([]byte, 0, length*arraylength)

	for i := 0; i < length; i++ {
		temp := any(v.array[i]).([]byte)
		data = append(data, temp...)
	}

	return data, nil
}
func (v *FTBytearrayArray) EstimatedSize() int64 { return v.Length() * v.width }
func (v *FTBytearrayArray) String() string       { return fmt.Sprintf("%v", v.array) }
func (v *FTBytearrayArray) DebugString() string {
	return fmt.Sprintf("%v(Length=%v,Memory=%v)", v.Name(), v.Length(), PrintSize(uint64(v.EstimatedSize())))
}
func (v *FTBytearrayArray) PythonString() string {
	var out []string
	for _, x := range v.array {
		out = append(out, strings.Trim(fmt.Sprint(x), "[]"))
	}
	return fmt.Sprintf("bytearraylist %v", strings.Join(out[:], ","))

}
func (v *FTBytearrayArray) Clone() (TypeVal, error) {
	return v.Copy()
}
func (v *FTBytearrayArray) Copy() (*FTBytearrayArray, error) {
	return NewFTBytearrayArray(int(v.width), v.array...)
}
func (v *FTBytearrayArray) AsType(tc TypeCode) (TypeVal, error) {
	switch tc.GetBase() {
	case BytearrayB:
		if v.Width() > int64(tc.Length()) {
			return nil, fmt.Errorf("can not convert to a smaller bytearray: %v -> %v", v.TypeCode(), tc)
		}
		xs := make([][]byte, len(v.array))
		for i, v := range v.array {
			x := make([]byte, tc.Length())
			copy(x, v)
			xs[i] = x
		}
		return &FTBytearrayArray{xs, int64(tc.Length())}, nil
	case IntegerB:
		if v.TypeCode().Length() != 8 {
			return nil, errors.New("can only convert from b8 to integer")
		}
		xs := make([]int64, len(v.array))
		for i, v := range v.array {
			xs[i] = int64(binary.LittleEndian.Uint64(v))
		}
		return NewFTIntegerArray(xs...), nil
	case Ed25519IntB:
		if v.Width() != 32 {
			return nil, errors.New("can only convert from b32 to Ed25519Int")
		}
		xs := make([]*edwards25519.Scalar, len(v.array))
		var err error
		for i, v := range v.array {
			xs[i], err = edwards25519.NewScalar().SetCanonicalBytes(v)
			if err != nil {
				return nil, err
			}
		}
		return NewFTEd25519IntArray(xs...), nil

	default:
		return nil, fmt.Errorf("conversion not supported: %v -> %v", v.TypeCode(), tc)
	}

}

func (v *FTBytearrayArray) Length() int64 { return int64(len(v.array)) }
func (v *FTBytearrayArray) SetLength(x int64) error {
	if v.Length() == x {
		return nil
	}

	xs := make([][]byte, x)

	for i := range xs {
		if i < len(v.array) {
			xs[i] = v.array[i]
		} else {
			xs[i] = make([]byte, v.width)
		}
	}

	v.array = xs
	return nil
}
func (v *FTBytearrayArray) Remove(indexes *FTIntegerArray) error {
	xs, err := SliceRemove(v.array, indexes.array)
	if err != nil {
		return err
	}
	v.array = xs
	return nil
}

func (v *FTBytearrayArray) Broadcast(length int64) error {
	xs, err := SliceBroadcast(v.array, length)
	if err != nil {
		return err
	}
	v.array = xs
	return nil
}
func (v *FTBytearrayArray) Lookup(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if defaultValue == nil {
		defaultValue = NewFTBytearrayArrayOrPanic(
			int(v.width),
			make([]byte, v.width),
		)
	}
	return v.Get(indexes, defaultValue)
}
func (v *FTBytearrayArray) Get(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if indexes == nil {
		xs, err := v.Clone()
		if err != nil {
			return nil, err
		}
		return xs.(ArrayTypeVal), nil
	}

	var d *[]byte

	if defaultValue != nil {
		vs, ok := defaultValue.(*FTBytearrayArray)
		if !ok {
			return nil, fmt.Errorf("defaultValue is not an %v", v.Name())
		}
		x, err := vs.Single()
		if err != nil {
			return nil, err
		}
		d = &x
	}

	result := make([][]byte, len(indexes.array))

	for i, k := range indexes.array {
		key := int(k)

		var x []byte
		if key >= len(v.array) || key < 0 {
			if d == nil {
				return nil, fmt.Errorf("out of range: %d", key)
			}

			x = *d
		} else {
			x = v.array[key]
		}

		result[i] = make([]byte, len(x))
		copy(result[i], x)
	}

	return &FTBytearrayArray{result, v.width}, nil
}
func (v *FTBytearrayArray) Set(indexes *FTIntegerArray, values ArrayTypeVal) error {
	vs, ok := values.(*FTBytearrayArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}
	return SliceSet(v.array, indexes.array, vs.array)
}
func (v *FTBytearrayArray) Element(i int64) interface{} { return v.array[i] }

func (v *FTBytearrayArray) Index() *FTIntegerArray {
	result := make([]int64, 0)

	for i := 0; i < len(v.array); i++ {
		for j := 0; j < len(v.array[i]); j++ {
			if v.array[i][j] != 0 {
				result = append(result, int64(i))
				break
			}
		}
	}

	return &FTIntegerArray{result}
}
func (v *FTBytearrayArray) Contains(values ArrayTypeVal) (*FTIntegerArray, error) {
	items, ok := values.(*FTBytearrayArray)
	if !ok {
		return nil, fmt.Errorf("values is not an %v", v.Name())
	}

	// TODO: use a better algorithm
	results := make([]int64, len(items.array))
	for i, x := range items.array {
		if SliceContains(v.array, x, func(a, b []byte) bool { return slices.Equal(a, b) }) {
			results[i] = 1
		} else {
			results[i] = 0
		}
	}

	return &FTIntegerArray{results}, nil
}

func (v *FTBytearrayArray) reduce(indexes *FTIntegerArray, values *FTBytearrayArray, useCurrentValue bool, f func(x, y []byte) []byte) {
	updated := make(map[int64]struct{})

	for i, k := range indexes.array {
		if useCurrentValue {
			v.array[k] = f(v.array[k], values.array[i])
		} else {
			if _, ok := updated[k]; ok {
				v.array[k] = f(v.array[k], values.array[i])
			} else {

				x := make([]byte, len(values.array[i]))
				copy(x, values.array[i])

				v.array[k] = x
				updated[k] = struct{}{}
			}
		}
	}
}

func orBytes(a, b []byte) []byte {
	result := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = a[i] | b[i]
	}
	return result
}

func (v *FTBytearrayArray) ReduceSum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTBytearrayArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, false, orBytes)

	return nil
}
func (v *FTBytearrayArray) ReduceISum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTBytearrayArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, true, orBytes)

	return nil
}

func (v *FTBytearrayArray) CumSum() (ArrayTypeVal, error) {
	results := make([][]byte, len(v.array))

	sum := make([]byte, v.width)

	for i, x := range v.array {
		for i, v := range x {
			sum[i] = sum[i] | v
		}

		y := make([]byte, len(sum))
		copy(y, sum)

		results[i] = y
	}

	return &FTBytearrayArray{results, v.width}, nil
}

func (v *FTBytearrayArray) Mux(condition *FTIntegerArray, ifFalse ArrayTypeVal) (ArrayTypeVal, error) {
	fs, ok := ifFalse.(*FTBytearrayArray)
	if !ok {
		return nil, fmt.Errorf("ifFalse is not an %v", v.Name())
	}
	result, err := SliceMux(v.array, condition.array, fs.array)
	if err != nil {
		return nil, err
	}
	return &FTBytearrayArray{result, v.width}, nil
}

func (v *FTBytearrayArray) Concat(values *FTBytearrayArray) (*FTBytearrayArray, error) {
	if len(v.array) != len(values.array) {
		return nil, errors.New("arrays have different lengths")
	}

	if len(v.array) == 0 {
		return nil, errors.New("arrays are empty")
	}

	result := make([][]byte, len(v.array))
	for i, a := range v.array {
		b := values.array[i]
		result[i] = append(a, b...)
	}

	return &FTBytearrayArray{result, v.width + values.width}, nil
}

func (v *FTBytearrayArray) Project(width int64, indexes *FTIntegerArray, values *FTIntegerArray) (*FTBytearrayArray, error) {
	if v.Length() == 0 {
		return NewFTBytearrayArrayFromBytes(nil, width)
	}

	if indexes.Length() != values.Length() {
		return nil, errors.New("mapping key and value arrays must match in length")
	}

	mappings := make(map[int64]int64)
	for i := range indexes.array {
		mappings[indexes.array[i]] = values.array[i]
	}

	result := make([][]byte, v.Length())

	for k, v := range v.array {
		result[k] = make([]byte, width)
		for i, j := range mappings {
			result[k][i] = v[j]
		}
	}

	return &FTBytearrayArray{result, width}, nil
}

func NewFTBytearrayArray(width int, xs ...[]byte) (*FTBytearrayArray, error) {
	ys := make([][]byte, len(xs))

	if len(xs) > 0 {

		for i, x := range xs {
			if len(x) != width {
				return nil, fmt.Errorf("bytearrays must all have the same length, %v != %v", len(x), width)
			}
			// Copy value
			y := make([]byte, len(x))
			copy(y, x)
			ys[i] = y
		}
	}

	return &FTBytearrayArray{ys, int64(width)}, nil
}
func NewFTBytearrayArrayOrPanic(width int, xs ...[]byte) *FTBytearrayArray {
	result, err := NewFTBytearrayArray(width, xs...)
	if err != nil {
		panic(err)
	}
	return result
}
func NewRandomFTBytearrayArray(width, length int64) (*FTBytearrayArray, error) {
	xs := make([][]byte, length)

	for i := int64(0); i < length; i++ {
		x := make([]byte, width)
		_, err := rand.Read(x)
		if err != nil {
			return nil, fmt.Errorf("unable to create randomarray of byte arrays: %v", err)
		}
		xs[i] = x
	}

	return &FTBytearrayArray{xs, width}, nil
}
func NewFTBytearrayArrayFromBytes(bs []byte, width int64) (*FTBytearrayArray, error) {
	if len(bs)%int(width) != 0 {
		panic("can't convert bytes to bytearray array")
	}

	elements := int64(len(bs)) / width
	xs := make([][]byte, elements)

	for i := int64(0); i < elements; i++ {
		xs[i] = bs[i*width : (i+1)*width]
	}

	return &FTBytearrayArray{xs, width}, nil
}

func (v *FTBytearrayArray) Single() ([]byte, error) {
	if len(v.array) != 1 {
		return nil, errors.New("not a singleton array")
	}
	return v.array[0], nil
}
func (v *FTBytearrayArray) Values() [][]byte {
	return v.array
}

func asFTBytearrayArray(xs ArrayTypeVal) (*FTBytearrayArray, error) {
	ys, ok := xs.(*FTBytearrayArray)
	if !ok {
		return nil, fmt.Errorf("value is not an %v", xs.Name())
	}
	return ys, nil
}

func (v *FTBytearrayArray) Eq(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTBytearrayArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b []byte) int64 { return BToI(slices.Equal(a, b)) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTBytearrayArray) Ne(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTBytearrayArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b []byte) int64 { return BToI(!slices.Equal(a, b)) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}

func (v *FTBytearrayArray) LShift(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a []byte, b int64) []byte {
		return ShiftLeft(a, int(b))
	})
	if err != nil {
		return nil, err
	}

	return &FTBytearrayArray{result, v.Width()}, nil
}
func (v *FTBytearrayArray) RShift(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a []byte, b int64) []byte {
		return ShiftLeft(a, int(b*-1))
	})
	if err != nil {
		return nil, err
	}

	return &FTBytearrayArray{result, v.Width()}, nil
}
func (v *FTBytearrayArray) And(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTBytearrayArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinaryE(v.array, bs.array, func(a, b []byte) ([]byte, error) {
		if len(a) != len(b) {
			return nil, errors.New("bytearray size mismatch")
		}
		result := make([]byte, len(a))
		for i, v := range a {
			result[i] = v & b[i]
		}
		return result, nil
	})
	if err != nil {
		return nil, err
	}

	return &FTBytearrayArray{result, v.Width()}, nil
}
func (v *FTBytearrayArray) Or(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTBytearrayArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinaryE(v.array, bs.array, func(a, b []byte) ([]byte, error) {
		if len(a) != len(b) {
			return nil, errors.New("bytearray size mismatch")
		}
		result := make([]byte, len(a))
		for i, v := range a {
			result[i] = v | b[i]
		}
		return result, nil
	})
	if err != nil {
		return nil, err
	}

	return &FTBytearrayArray{result, v.Width()}, nil
}
func (v *FTBytearrayArray) Xor(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTBytearrayArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinaryE(v.array, bs.array, func(a, b []byte) ([]byte, error) {
		if len(a) != len(b) {
			return nil, errors.New("bytearray size mismatch")
		}
		result := make([]byte, len(a))
		for i, v := range a {
			result[i] = v ^ b[i]
		}
		return result, nil
	})
	if err != nil {
		return nil, err
	}

	return &FTBytearrayArray{result, v.Width()}, nil
}
func (v *FTBytearrayArray) Invert() (ArrayTypeVal, error) {
	result, err := SliceMapUnary(v.array, func(a []byte) []byte {
		result := make([]byte, len(a))
		for i, v := range a {
			result[i] = ^v
		}
		return result
	})
	if err != nil {
		return nil, err
	}

	return &FTBytearrayArray{result, v.Width()}, nil
}

func ShiftLeft(xs []byte, bits int) []byte {
	data := make([]byte, len(xs))
	copy(data, xs)

	n := len(data)
	shiftBytes := bits / 8
	shiftBits := bits % 8

	// First we shift the bytes

	if shiftBytes < 0 {
		shiftBytes = -shiftBytes
		for i := n - 1; i >= 0; i-- {
			bFrom := i - shiftBytes
			if bFrom >= 0 {
				data[i] = data[bFrom]
				data[bFrom] = 0
			} else {
				data[i] = 0
			}
		}
	} else if shiftBytes > 0 {
		for i := 0; i < n; i++ {
			bFrom := i + shiftBytes
			if bFrom < n {
				data[i] = data[bFrom]
				data[bFrom] = 0
			} else {
				data[i] = 0
			}
		}
	}

	// Then shift the bits

	// https://stackoverflow.com/questions/29442710/how-to-shift-byte-array-with-golang

	if shiftBits < 0 {
		shiftBits = -shiftBits
		for i := n - 1; i > 0; i-- {
			data[i] = data[i]>>shiftBits | data[i-1]<<(8-shiftBits)
		}
		data[0] >>= shiftBits
	} else if shiftBits > 0 {
		for i := 0; i < n-1; i++ {
			data[i] = data[i]<<shiftBits | data[i+1]>>(8-shiftBits)
		}
		data[n-1] <<= shiftBits
	}

	return data
}

func CalcBytearrayWidth(xs [][]byte) int {
	if len(xs) > 0 {
		return len(xs[0])
	}
	return 0
}
