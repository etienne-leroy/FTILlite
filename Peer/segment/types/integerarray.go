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

type FTIntegerArray struct {
	array []int64
}

func (v *FTIntegerArray) TypeCode() TypeCode { return Integer }
func (v *FTIntegerArray) Equals(other TypeVal) bool {
	os, ok := other.(*FTIntegerArray)
	return ok && slices.Equal(v.array, os.array)
}
func (v *FTIntegerArray) Name() string { return "IntegerArray" }

func (v *FTIntegerArray) GetBinaryArray(index int) ([]byte, error) {
	wbuf := new(bytes.Buffer)
	err := binary.Write(wbuf, binary.LittleEndian, v.array)
	if err != nil {
		return nil, err
	}
	return wbuf.Bytes(), nil
}
func (v *FTIntegerArray) EstimatedSize() int64 { return v.Length() * 8 }
func (v *FTIntegerArray) String() string       { return fmt.Sprintf("%v", v.array) }
func (v *FTIntegerArray) DebugString() string {
	return fmt.Sprintf("%v(Length=%v,Memory=%v)", v.Name(), v.Length(), PrintSize(uint64(v.EstimatedSize())))
}
func (v *FTIntegerArray) PythonString() string {
	return fmt.Sprintf("intlist %s", strings.Trim(fmt.Sprint(v.array), "[]"))
}
func (v *FTIntegerArray) Length() int64 { return int64(len(v.array)) }
func (v *FTIntegerArray) SetLength(x int64) error {
	if v.Length() == x {
		return nil
	}

	xs := make([]int64, x)

	for i := range xs {
		if i < len(v.array) {
			xs[i] = v.array[i]
		}
	}

	v.array = xs
	return nil
}

func (v *FTIntegerArray) Clone() (TypeVal, error) {
	return NewFTIntegerArray(v.array...), nil
}
func (v *FTIntegerArray) AsType(tc TypeCode) (TypeVal, error) {
	switch tc.GetBase() {
	case IntegerB:
		return v.Clone()
	case FloatB:
		xs := make([]float64, len(v.array))
		for i := range v.array {
			xs[i] = float64(v.array[i])
		}
		return NewFTFloatArray(xs...), nil
	case BytearrayB:
		if tc.Length() < 8 {
			return nil, errors.New("arrays of integers can only be converted to bytearrays of length 8 or more")
		}

		xs := make([][]byte, len(v.array))
		for i := range v.array {
			b := make([]byte, tc.Length())

			binary.LittleEndian.PutUint64(b, uint64(v.array[i]))

			// Fill the rest of the array with the sign bit
			if len(b) > 8 {
				if b[7]>>7 == 1 {
					for i := 8; i < len(b); i++ {
						b[i] = 1<<8 - 1
					}
				}
			}

			xs[i] = b
		}

		return NewFTBytearrayArray(tc.Length(), xs...)
	case Ed25519IntB:
		return NewFTEd25519IntArrayFromInt64s(v.array...), nil
	case Ed25519B:
		return NewEd25519ArrayFromInt64s(v.array...)
	default:
		return nil, fmt.Errorf("conversion not supported: %v -> %v", v.TypeCode(), tc)
	}
}

func (v *FTIntegerArray) Sort() ArrayTypeVal {
	result := make([]int64, len(v.array))
	copy(result, v.array)
	slices.Sort(result)

	return &FTIntegerArray{result}
}
func (v *FTIntegerArray) IndexSort(indexes *FTIntegerArray) (*FTIntegerArray, error) {
	p, err := NewPairSlice(v.array, indexes.array)
	if err != nil {
		return nil, err
	}
	sort.Sort(p)

	return &FTIntegerArray{p.Index()}, nil
}

func (v *FTIntegerArray) Remove(indexes *FTIntegerArray) error {
	xs, err := SliceRemove(v.array, indexes.array)
	if err != nil {
		return err
	}
	v.array = xs
	return nil
}
func (v *FTIntegerArray) Broadcast(length int64) error {
	xs, err := SliceBroadcast(v.array, length)
	if err != nil {
		return err
	}
	v.array = xs
	return nil
}
func (v *FTIntegerArray) Lookup(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if defaultValue == nil {
		defaultValue = NewFTIntegerArray(0)
	}
	return v.Get(indexes, defaultValue)
}
func (v *FTIntegerArray) Get(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if indexes == nil {
		xs, err := v.Clone()
		if err != nil {
			return nil, err
		}
		return xs.(ArrayTypeVal), nil
	}

	var d *int64

	if defaultValue != nil {
		vs, ok := defaultValue.(*FTIntegerArray)
		if !ok {
			return nil, fmt.Errorf("defaultValue is not an %v", v.Name())
		}
		x, err := vs.Single()
		if err != nil {
			return nil, err
		}
		d = &x
	}

	result := make([]int64, len(indexes.array))

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

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Set(indexes *FTIntegerArray, values ArrayTypeVal) error {
	vs, ok := values.(*FTIntegerArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}
	return SliceSet(v.array, indexes.array, vs.array)
}
func (v *FTIntegerArray) Element(i int64) interface{} {
	return v.array[i]
}

func (v *FTIntegerArray) Index() *FTIntegerArray {
	result := make([]int64, 0)

	for i := 0; i < len(v.array); i++ {
		if v.array[i] != 0 {
			result = append(result, int64(i))
		}
	}

	return &FTIntegerArray{result}
}

func (v *FTIntegerArray) NonZero() bool {
	for _, r := range v.array {
		if r == 0 {
			return false
		}
	}
	return true
}
func (v *FTIntegerArray) Contains(values ArrayTypeVal) (*FTIntegerArray, error) {
	vs, ok := values.(*FTIntegerArray)
	if !ok {
		return nil, fmt.Errorf("values is not an %v", v.Name())
	}

	set := make(map[int64]struct{})
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

func (v *FTIntegerArray) reduce(indexes *FTIntegerArray, values *FTIntegerArray, useCurrentValue bool, f func(x int64, y int64) int64) {
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

func (v *FTIntegerArray) ReduceSum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTIntegerArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, false, func(x, y int64) int64 { return x + y })

	return nil
}
func (v *FTIntegerArray) ReduceISum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTIntegerArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, true, func(x, y int64) int64 { return x + y })

	return nil
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func (v *FTIntegerArray) ReduceMin(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTIntegerArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, false, minInt64)

	return nil
}
func (v *FTIntegerArray) ReduceIMin(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTIntegerArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, true, minInt64)

	return nil
}
func (v *FTIntegerArray) ReduceMax(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTIntegerArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, false, maxInt64)

	return nil
}
func (v *FTIntegerArray) ReduceIMax(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTIntegerArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, true, maxInt64)

	return nil
}

func (v *FTIntegerArray) CumSum() (ArrayTypeVal, error) {
	results := make([]int64, len(v.array))

	sum := int64(0)

	for i, v := range v.array {
		sum += v
		results[i] = sum
	}

	return &FTIntegerArray{results}, nil
}

func (v *FTIntegerArray) Mux(condition *FTIntegerArray, ifFalse ArrayTypeVal) (ArrayTypeVal, error) {
	fs, ok := ifFalse.(*FTIntegerArray)
	if !ok {
		return nil, fmt.Errorf("ifFalse is not an %v", v.Name())
	}
	result, err := SliceMux(v.array, condition.array, fs.array)
	if err != nil {
		return nil, err
	}
	return &FTIntegerArray{result}, nil
}

func NewFTIntegerArray(xs ...int64) *FTIntegerArray { return &FTIntegerArray{xs} }

func NewRandomFTIntegerArray(min, max, length int64) (*FTIntegerArray, error) {
	// rand.Int is [min,max), adding one to be [min,max].
	var bgMaxInc big.Int
	bgMaxInc.Add(big.NewInt(max), big.NewInt(1))

	bMin := big.NewInt(min)

	var bg big.Int
	bg.Sub(&bgMaxInc, bMin)

	xs := make([]int64, length)
	for i := int64(0); i < length; i++ {
		n, err := rand.Int(rand.Reader, &bg)

		if err != nil {
			return nil, fmt.Errorf("unable to create randomarray of integers: %v", err)
		}
		xs[i] = n.Int64() + min
	}
	return &FTIntegerArray{xs}, nil
}

func NewRandomPermFTIntegerArray(length int64, n int64) (*FTIntegerArray, error) {
	if n < length {
		return nil, errors.New("n must be greater than or equal to length")
	}

	m := make([]int64, n)
	var i int64
	for i = 0; i < n; i++ {
		m[i] = i
	}

	randReader := rand.Reader

	for i := n - 1; n-i <= length; i-- {
		r, err := rand.Int(randReader, big.NewInt(i+1))
		if err != nil {
			return nil, fmt.Errorf("could not create integer for random perm, %v", err)
		}
		j := r.Int64()
		m[i], m[j] = m[j], m[i]
	}

	return &FTIntegerArray{m[n-length:]}, nil
}

func NewFTIntegerArrayFromBytes(bs []byte) (*FTIntegerArray, error) {

	if len(bs)%8 != 0 {
		return nil, errors.New("can't convert bytes to int64 array")
	}

	count := len(bs) / 8
	xs := make([]int64, count)
	for i := 0; i < count; i++ {
		b := bs[i*8 : (i+1)*8]
		buf := bytes.NewReader(b)

		var x int64
		err := binary.Read(buf, binary.LittleEndian, &x)
		if err != nil {
			return nil, err
		}
		xs[i] = x
	}

	return &FTIntegerArray{xs}, nil
}
func ArangeFTIntegerArray(length int64) *FTIntegerArray {
	xs := make([]int64, length)
	for i := int64(0); i < int64(length); i++ {
		xs[i] = i
	}
	return NewFTIntegerArray(xs...)
}

func (v *FTIntegerArray) Single() (int64, error) {
	if len(v.array) != 1 {
		return 0, errors.New("not a singleton array")
	}
	return v.array[0], nil
}
func (v *FTIntegerArray) Values() []int64 {
	return v.array
}

func asFTIntegerArray(xs ArrayTypeVal) (*FTIntegerArray, error) {
	ys, ok := xs.(*FTIntegerArray)
	if !ok {
		return nil, fmt.Errorf("value is not an %v", xs.Name())
	}
	return ys, nil
}

func (v *FTIntegerArray) Eq(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return BToI(a == b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Ne(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return BToI(a != b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Gt(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return BToI(a > b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Lt(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return BToI(a < b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Ge(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return BToI(a >= b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Le(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return BToI(a <= b) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Neg() (ArrayNegTypeVal, error) {
	result, err := SliceMapUnary(v.array, func(a int64) int64 { return -a })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}

func (v *FTIntegerArray) Abs() (ArrayAbsTypeVal, error) {
	result, err := SliceMapUnary(v.array, func(a int64) int64 {
		if a > 0 {
			return a
		}
		return a * -1
	})
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}

func (v *FTIntegerArray) Add(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return a + b })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Sub(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return a - b })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Mul(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 { return a * b })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}

func (v *FTIntegerArray) FloorDiv(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b int64) int64 {
		quotient, _ := divmod(a, b)
		return quotient
	})
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}

func (v *FTIntegerArray) TrueDiv(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinaryE(v.array, bs.array, func(a, b int64) (float64, error) {
		if b == 0 {
			return 0.0, errors.New("cannot divide by zero")
		}
		return float64(a) / float64(b), nil
	})
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}

func (v *FTIntegerArray) Mod(other *FTIntegerArray) (*FTIntegerArray, error) {
	result, err := SliceMapBinary(v.array, other.array, leastNegativeRemainder)
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) DivMod(other *FTIntegerArray) (*FTIntegerArray, *FTIntegerArray, error) {
	as, bs, err := SliceMapBinary2(v.array, other.array, divmod)
	if err != nil {
		return nil, nil, err
	}

	return &FTIntegerArray{as}, &FTIntegerArray{bs}, nil
}
func (v *FTIntegerArray) Pow(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}
	result, err := SliceMapBinaryE(v.array, bs.array, func(a, b int64) (int64, error) {
		if a == 0 && b == 0 {
			return 0, fmt.Errorf("cannot perform 0 to the power of 0")
		}

		return int64(math.Pow(float64(a), float64(b))), nil
	})
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}

func (v *FTIntegerArray) LShift(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a int64, b int64) int64 { return a << b })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) RShift(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a int64, b int64) int64 { return a >> b })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) And(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a int64, b int64) int64 { return a & b })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Or(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a int64, b int64) int64 { return a | b })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Xor(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a int64, b int64) int64 { return a ^ b })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTIntegerArray) Invert() (ArrayTypeVal, error) {
	result, err := SliceMapUnary(v.array, func(a int64) int64 { return ^a })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}

func (v *FTIntegerArray) Nearest() (*FTFloatArray, error) {
	result, err := SliceMapUnary(v.array, func(a int64) float64 { return float64(a) })
	if err != nil {
		return nil, err
	}

	return &FTFloatArray{result}, nil
}
func divmod(x int64, y int64) (quotient int64, remainder int64) {
	quotient = x / y
	remainder = leastNegativeRemainder(x, y)

	if remainder != 0 && quotient < 0 {
		quotient--
	}
	return quotient, remainder
}
func leastNegativeRemainder(x int64, y int64) int64 {
	// Python mod returns the least negative remainder, which is different to Go which returns the least positive.
	return (x%y + y) % y
}
