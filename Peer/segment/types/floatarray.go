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
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"

	"golang.org/x/exp/slices"
)

type FTFloatArray struct {
	array []float64
}

func (v *FTFloatArray) TypeCode() TypeCode { return Float }
func (v *FTFloatArray) Equals(other TypeVal) bool {
	os, ok := other.(*FTFloatArray)
	return ok && slices.Equal(v.array, os.array)
}
func (v *FTFloatArray) Name() string { return "FloatArray" }

func (v *FTFloatArray) GetBinaryArray(index int) ([]byte, error) {
	wbuf := new(bytes.Buffer)
	err := binary.Write(wbuf, binary.LittleEndian, v.array)
	if err != nil {
		return nil, err
	}
	return wbuf.Bytes(), nil
}
func (v *FTFloatArray) EstimatedSize() int64 { return v.Length() * 8 }
func (v *FTFloatArray) String() string       { return fmt.Sprintf("%v", v.array) }
func (v *FTFloatArray) DebugString() string {
	return fmt.Sprintf("%v(Length=%v,Memory=%v)", v.Name(), v.Length(), PrintSize(uint64(v.EstimatedSize())))
}
func (v *FTFloatArray) PythonString() string {
	return fmt.Sprintf("floatlist %s", strings.Trim(fmt.Sprint(v.array), "[]"))
}
func (v *FTFloatArray) Clone() (TypeVal, error) {
	return NewFTFloatArray(v.array...), nil
}
func (v *FTFloatArray) AsType(tc TypeCode) (TypeVal, error) {
	return nil, fmt.Errorf("conversion not supported: %v -> %v", v.TypeCode(), tc)
}
func (v *FTFloatArray) Sort() ArrayTypeVal {
	result := make([]float64, len(v.array))
	copy(result, v.array)
	slices.Sort(result)

	return &FTFloatArray{result}
}
func (v *FTFloatArray) IndexSort(indexes *FTIntegerArray) (*FTIntegerArray, error) {
	p, err := NewPairSlice(v.array, indexes.array)
	if err != nil {
		return nil, err
	}
	sort.Sort(p)

	return &FTIntegerArray{p.Index()}, nil
}
func (v *FTFloatArray) Length() int64 { return int64(len(v.array)) }
func (v *FTFloatArray) SetLength(x int64) error {
	if v.Length() == x {
		return nil
	}

	xs := make([]float64, x)

	for i := range xs {
		if i < len(v.array) {
			xs[i] = v.array[i]
		}
	}

	v.array = xs
	return nil
}
func (v *FTFloatArray) Remove(indexes *FTIntegerArray) error {
	xs, err := SliceRemove(v.array, indexes.array)
	if err != nil {
		return err
	}
	v.array = xs
	return nil
}
func (v *FTFloatArray) Broadcast(length int64) error {
	xs, err := SliceBroadcast(v.array, length)
	if err != nil {
		return err
	}
	v.array = xs
	return nil
}
func (v *FTFloatArray) Lookup(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if defaultValue == nil {
		defaultValue = NewFTFloatArray(0)
	}
	return v.Get(indexes, defaultValue)
}
func (v *FTFloatArray) Get(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if indexes == nil {
		xs, err := v.Clone()
		if err != nil {
			return nil, err
		}
		return xs.(ArrayTypeVal), nil
	}

	var d *float64

	if defaultValue != nil {
		vs, ok := defaultValue.(*FTFloatArray)
		if !ok {
			return nil, fmt.Errorf("defaultValue is not an %v", v.Name())
		}
		x, err := vs.Single()
		if err != nil {
			return nil, err
		}
		d = &x
	}

	result := make([]float64, len(indexes.array))

	for i, k := range indexes.array {
		key := int(k)

		if key >= len(v.array) || key < 0 {
			if d == nil {
				return nil, fmt.Errorf("out of range: %d", key)
			}
			result[i] = *d
		} else {
			result[i] = v.array[key]
		}
	}

	return &FTFloatArray{result}, nil
}
func (v *FTFloatArray) Set(indexes *FTIntegerArray, values ArrayTypeVal) error {
	vs, ok := values.(*FTFloatArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}
	return SliceSet(v.array, indexes.array, vs.array)
}
func (v *FTFloatArray) Element(i int64) interface{} { return v.array[i] }

func (v *FTFloatArray) Index() *FTIntegerArray {
	result := make([]int64, 0)

	for i := 0; i < len(v.array); i++ {
		if v.array[i] != 0 {
			result = append(result, int64(i))
		}
	}

	return &FTIntegerArray{result}
}
func (v *FTFloatArray) Contains(values ArrayTypeVal) (*FTIntegerArray, error) {
	vs, ok := values.(*FTFloatArray)
	if !ok {
		return nil, fmt.Errorf("values is not an %v", v.Name())
	}

	set := make(map[float64]struct{})
	for _, v := range v.array {
		set[v] = struct{}{}
	}

	results := make([]int64, len(vs.array))

	for i, x := range vs.array {
		if _, ok := set[x]; ok {
			results[i] = 1
		} else {
			results[i] = 0
		}
	}

	return &FTIntegerArray{results}, nil
}
func (v *FTFloatArray) reduce(indexes *FTIntegerArray, values *FTFloatArray, useCurrentValue bool, f func(x float64, y float64) float64) {
	updated := make(map[int64]struct{})

	for i, k := range indexes.array {
		if useCurrentValue {
			v.array[k] = f(v.array[k], values.array[i])
		} else {
			if _, ok := updated[k]; ok {
				v.array[k] = f(v.array[k], values.array[i])
			} else {
				v.array[k] = values.array[i]
				updated[k] = struct{}{}
			}
		}
	}
}

func (v *FTFloatArray) ReduceSum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTFloatArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, false, func(x, y float64) float64 { return x + y })

	return nil
}
func (v *FTFloatArray) ReduceISum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTFloatArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, true, func(x, y float64) float64 { return x + y })

	return nil
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (v *FTFloatArray) ReduceMin(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTFloatArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, false, minFloat64)

	return nil
}
func (v *FTFloatArray) ReduceIMin(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTFloatArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, true, minFloat64)

	return nil
}
func (v *FTFloatArray) ReduceMax(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTFloatArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, false, maxFloat64)

	return nil
}
func (v *FTFloatArray) ReduceIMax(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTFloatArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, true, maxFloat64)

	return nil
}

func (v *FTFloatArray) CumSum() (ArrayTypeVal, error) {
	results := make([]float64, len(v.array))

	sum := float64(0)

	for i, v := range v.array {
		sum += v
		results[i] = sum
	}

	return &FTFloatArray{results}, nil
}

func (v *FTFloatArray) Mux(condition *FTIntegerArray, ifFalse ArrayTypeVal) (ArrayTypeVal, error) {
	fs, ok := ifFalse.(*FTFloatArray)
	if !ok {
		return nil, fmt.Errorf("ifFalse is not an %v", v.Name())
	}
	result, err := SliceMux(v.array, condition.array, fs.array)
	if err != nil {
		return nil, err
	}
	return &FTFloatArray{result}, nil
}

func NewFTFloatArray(xs ...float64) *FTFloatArray { return &FTFloatArray{xs} }
func NewRandomFTFloatArray(min, max float64, length int64) (*FTFloatArray, error) {
	bg := new(big.Int).SetUint64(uint64(1 << 63))

	xs := make([]float64, length)

	for i := int64(0); i < length; i++ {
		n, err := rand.Int(rand.Reader, bg)

		if err != nil {
			return nil, fmt.Errorf("unable to create randomarray of floats: %v", err)
		}
		xs[i] = min + ((max-min)*float64(n.Int64()))/float64(1<<63)
	}
	return &FTFloatArray{xs}, nil
}
func NewFTFloatArrayFromBytes(bs []byte) (*FTFloatArray, error) {
	if len(bs)%8 != 0 {
		return nil, errors.New("can't convert bytes to float64 array")
	}

	count := len(bs) / 8
	xs := make([]float64, count)
	for i := 0; i < count; i++ {
		b := bs[i*8 : (i+1)*8]
		buf := bytes.NewReader(b)

		var x float64
		err := binary.Read(buf, binary.LittleEndian, &x)
		if err != nil {
			return nil, err
		}
		xs[i] = x
	}

	return &FTFloatArray{xs}, nil
}

func (v *FTFloatArray) Single() (float64, error) {
	if len(v.array) != 1 {
		return 0, errors.New("not a singleton array")
	}
	return v.array[0], nil
}
func (v *FTFloatArray) Values() []float64 {
	return v.array
}

func asFTFloatArray(xs ArrayTypeVal) (*FTFloatArray, error) {
	ys, ok := xs.(*FTFloatArray)
	if !ok {
		return nil, fmt.Errorf("value is not an %v", xs.Name())
	}
	return ys, nil
}
func (v *FTFloatArray) Eq(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) int64 { return BToI(a == b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Ne(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) int64 { return BToI(a != b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Gt(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) int64 { return BToI(a > b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Lt(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) int64 { return BToI(a < b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Ge(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) int64 { return BToI(a >= b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Le(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) int64 { return BToI(a <= b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Neg() (ArrayNegTypeVal, error) {
	result, err := SliceMapUnary(v.array, func(a float64) float64 { return -a })
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}
func (v *FTFloatArray) Abs() (ArrayAbsTypeVal, error) {
	result, err := SliceMapUnary(v.array, math.Abs)
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}

func (v *FTFloatArray) Floor() (*FTIntegerArray, error) {
	result, err := SliceMapUnary(v.array, func(a float64) int64 { return int64(math.Floor(a)) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Ceil() (*FTIntegerArray, error) {
	result, err := SliceMapUnary(v.array, func(a float64) int64 { return int64(math.Ceil(a)) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Round() (*FTIntegerArray, error) {
	result, err := SliceMapUnary(v.array, func(a float64) int64 { return int64(math.Floor(a + 0.5)) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTFloatArray) Trunc() (*FTFloatArray, error) {
	result, err := SliceMapUnary(v.array, math.Trunc)
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}

func (v *FTFloatArray) Add(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) float64 { return a + b })
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}
func (v *FTFloatArray) Sub(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) float64 { return a - b })
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}

func (v *FTFloatArray) Mul(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b float64) float64 { return a * b })
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}

func (v *FTFloatArray) TrueDiv(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTFloatArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinaryE(v.array, bs.array, func(a, b float64) (float64, error) {
		if b == 0 {
			return 0.0, errors.New("cannot divide by zero")
		}
		return a / b, nil
	})
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}
func (v *FTFloatArray) Pow(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinaryE(v.array, bs.array, func(a float64, b int64) (float64, error) {
		if a == 0 && b == 0 {
			return 0, fmt.Errorf("cannot perform 0 to the power of 0")
		}
		return math.Pow(a, float64(b)), nil
	})
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}

func (v *FTFloatArray) Exp() (*FTFloatArray, error) {
	result, err := SliceMapUnary(v.array, math.Exp)
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}
func (v *FTFloatArray) Log() (*FTFloatArray, error) {
	for _, vx := range v.array {
		if vx < 0 {
			err := errors.New("cannot take log of a negative number")
			return nil, err
		}
	}
	result, err := SliceMapUnary(v.array, math.Log)
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}
func (v *FTFloatArray) Sin() (*FTFloatArray, error) {
	result, err := SliceMapUnary(v.array, math.Sin)
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}
func (v *FTFloatArray) Cos() (*FTFloatArray, error) {
	result, err := SliceMapUnary(v.array, math.Cos)
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}
