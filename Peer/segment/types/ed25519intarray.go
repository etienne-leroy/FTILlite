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
	"math/big"

	"filippo.io/edwards25519"
)

type FTEd25519IntArray struct {
	array []*edwards25519.Scalar
}

func (v *FTEd25519IntArray) TypeCode() TypeCode { return Ed25519Int }
func (v *FTEd25519IntArray) Equals(other TypeVal) bool {
	if os, ok := other.(*FTEd25519IntArray); ok {
		if v.Length() == os.Length() {
			for i, t := range v.array {
				o := os.array[i]

				if t.Equal(o) != 1 {
					return false
				}

			}
			return true
		}
	}
	return false
}
func (v *FTEd25519IntArray) Name() string { return "Ed25519IntArray" }

func (v *FTEd25519IntArray) GetBinaryArray(index int) ([]byte, error) {
	data := make([]byte, 0, len(v.array)*32)

	for _, v := range v.array {
		data = append(data, v.Bytes()...)
	}

	return data, nil
}
func (v *FTEd25519IntArray) EstimatedSize() int64 { return v.Length() * 32 }
func (v *FTEd25519IntArray) String() string {
	s := "["
	for _, v := range v.array {
		b := ScalarToBigInt(v)
		s += " " + b.String()
	}
	s += " ]"
	return s
}
func (v *FTEd25519IntArray) DebugString() string {
	return fmt.Sprintf("%v(Length=%v,Memory=%v)", v.Name(), v.Length(), PrintSize(uint64(v.EstimatedSize())))
}
func (v *FTEd25519IntArray) Clone() (TypeVal, error) {
	return NewFTEd25519IntArray(v.array...), nil
}
func (v *FTEd25519IntArray) AsType(tc TypeCode) (TypeVal, error) {
	switch tc.GetBase() {
	case Ed25519IntB:
		return v.Clone()
	case IntegerB:
		xs := make([]int64, len(v.array))
		for i, v := range v.array {
			var err error
			xs[i], err = ScalarToInt64(v)
			if err != nil {
				return nil, err
			}
		}
		return NewFTIntegerArray(xs...), nil
	case Ed25519B:
		return NewEd25519ArrayFromInt(v.array)
	case BytearrayB:
		if tc.Length() != 32 {
			return nil, errors.New("arrays of Ed25519 can only be converted to bytearrays of length 32")
		}
		xs := make([][]byte, len(v.array))
		for i, v := range v.array {
			xs[i] = v.Bytes()
		}
		return NewFTBytearrayArray(tc.Length(), xs...)
	default:
		return nil, fmt.Errorf("conversion not supported: %v -> %v", v.TypeCode(), tc)
	}

}

func (v *FTEd25519IntArray) Length() int64 { return int64(len(v.array)) }
func (v *FTEd25519IntArray) SetLength(x int64) error {
	if v.Length() == x {
		return nil
	}

	xs := make([]*edwards25519.Scalar, x)

	for i := range xs {
		if i < len(v.array) {
			xs[i] = v.array[i]
		} else {
			xs[i] = edwards25519.NewScalar()
		}
	}

	v.array = xs
	return nil
}
func (v *FTEd25519IntArray) Remove(indexes *FTIntegerArray) error {
	xs, err := SliceRemove(v.array, indexes.array)
	if err != nil {
		return err
	}
	v.array = xs
	return nil
}
func (v *FTEd25519IntArray) Broadcast(length int64) error {
	xs, err := SliceBroadcast(v.array, length)
	if err != nil {
		return err
	}
	v.array = xs
	return nil
}
func (v *FTEd25519IntArray) Lookup(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if defaultValue == nil {
		defaultValue = NewFTEd25519IntArrayFromInt64s(0)
	}
	return v.Get(indexes, defaultValue)
}
func (v *FTEd25519IntArray) Get(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error) {
	if indexes == nil {
		xs, err := v.Clone()
		if err != nil {
			return nil, err
		}
		return xs.(ArrayTypeVal), nil
	}

	var d *edwards25519.Scalar

	if defaultValue != nil {
		vs, ok := defaultValue.(*FTEd25519IntArray)
		if !ok {
			return nil, fmt.Errorf("defaultValue is not an %v", v.Name())
		}
		var err error
		d, err = vs.Single()
		if err != nil {
			return nil, err
		}
	}

	result := make([]*edwards25519.Scalar, len(indexes.array))

	for i, k := range indexes.array {
		key := int(k)

		var x *edwards25519.Scalar
		if key >= len(v.array) || key < 0 {
			if d == nil {
				return nil, fmt.Errorf("out of range: %d", key)
			}

			x = d
		} else {
			x = v.array[key]
		}

		y := *x
		result[i] = &y
	}

	return &FTEd25519IntArray{result}, nil
}
func (v *FTEd25519IntArray) Set(indexes *FTIntegerArray, values ArrayTypeVal) error {
	vs, ok := values.(*FTEd25519IntArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}
	return SliceSet(v.array, indexes.array, vs.array)
}
func (v *FTEd25519IntArray) Element(i int64) interface{} {
	return *v.array[i]
}

func (v *FTEd25519IntArray) Index() *FTIntegerArray {
	result := make([]int64, 0)

	zero := edwards25519.NewScalar()

	for i := 0; i < len(v.array); i++ {
		if v.array[i].Equal(zero) != 1 {
			result = append(result, int64(i))
		}
	}

	return &FTIntegerArray{result}
}
func (v *FTEd25519IntArray) Contains(values ArrayTypeVal) (*FTIntegerArray, error) {
	items, ok := values.(*FTEd25519IntArray)
	if !ok {
		return nil, fmt.Errorf("values is not an %v", v.Name())
	}

	// TODO: use a better algorithm
	results := make([]int64, len(items.array))
	for i, x := range items.array {
		if SliceContains(v.array, x, func(a, b *edwards25519.Scalar) bool { return a.Equal(b) == 1 }) {
			results[i] = 1
		} else {
			results[i] = 0
		}
	}

	return &FTIntegerArray{results}, nil
}
func (v *FTEd25519IntArray) reduce(indexes *FTIntegerArray, values *FTEd25519IntArray, useCurrentValue bool, f func(x *edwards25519.Scalar, y *edwards25519.Scalar) *edwards25519.Scalar) {
	updated := make(map[int64]struct{})

	for i, k := range indexes.array {
		if useCurrentValue {
			v.array[k] = f(v.array[k], values.array[i])
		} else {
			if _, ok := updated[k]; ok {
				v.array[k] = f(v.array[k], values.array[i])
			} else {
				x := *values.array[i]
				v.array[k] = &x
				updated[k] = struct{}{}
			}
		}
	}
}

func (v *FTEd25519IntArray) ReduceSum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTEd25519IntArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, false, func(x, y *edwards25519.Scalar) *edwards25519.Scalar { return x.Add(x, y) })

	return nil
}
func (v *FTEd25519IntArray) ReduceISum(indexes *FTIntegerArray, values ArrayTypeVal) error {
	xs, ok := values.(*FTEd25519IntArray)
	if !ok {
		return fmt.Errorf("values is not an %v", v.Name())
	}

	v.reduce(indexes, xs, true, func(x, y *edwards25519.Scalar) *edwards25519.Scalar { return x.Add(x, y) })

	return nil
}

func (v *FTEd25519IntArray) CumSum() (ArrayTypeVal, error) {
	results := make([]*edwards25519.Scalar, len(v.array))

	sum := edwards25519.NewScalar()

	for i, v := range v.array {
		s := edwards25519.NewScalar().Add(sum, v)
		results[i] = s
		sum = s
	}

	return &FTEd25519IntArray{results}, nil
}

func (v *FTEd25519IntArray) Mux(condition *FTIntegerArray, ifFalse ArrayTypeVal) (ArrayTypeVal, error) {
	fs, ok := ifFalse.(*FTEd25519IntArray)
	if !ok {
		return nil, fmt.Errorf("ifFalse is not an %v", v.Name())
	}
	result, err := SliceMux(v.array, condition.array, fs.array)
	if err != nil {
		return nil, err
	}
	return &FTEd25519IntArray{result}, nil
}

func NewFTEd25519IntArray(xs ...*edwards25519.Scalar) *FTEd25519IntArray {
	ys := make([]*edwards25519.Scalar, len(xs))
	for i, x := range xs {
		// Copy value
		y := *x
		ys[i] = &y
	}

	return &FTEd25519IntArray{ys}
}
func NewFTEd25519IntArrayFromInt64s(xs ...int64) *FTEd25519IntArray {
	ys := make([]*edwards25519.Scalar, len(xs))
	for i, x := range xs {
		ys[i] = Int64ToScalar(x)
	}

	return &FTEd25519IntArray{ys}
}
func NewRandomFTEd25519IntArray(nonZero bool, length int64) (*FTEd25519IntArray, error) {
	genRandED25519Int := func() (*edwards25519.Scalar, error) {
		r := make([]byte, 64)
		n, err := rand.Read(r)
		if err != nil {
			return nil, fmt.Errorf("unable to create randomarray of ed25519Int: %v", err)
		}
		if n != len(r) {
			panic("did not write all bytes")
		}

		s := edwards25519.NewScalar()
		_, err = s.SetUniformBytes(r)
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	xs := make([]*edwards25519.Scalar, length)

	zero := edwards25519.NewScalar()

	for i := int64(0); i < length; i++ {
		var v *edwards25519.Scalar
		for {
			var err error

			v, err = genRandED25519Int()
			if err != nil {
				return nil, err
			}

			if !nonZero || v.Equal(zero) == 0 {
				break
			}
		}
		xs[i] = v
	}

	return &FTEd25519IntArray{xs}, nil
}
func NewFTEd25519IntArrayFromBytes(bs []byte) (*FTEd25519IntArray, error) {
	bLen := len(bs)

	if bLen%32 != 0 {
		return nil, errors.New("can't convert bytes to Ed25519Int array")
	}

	count := bLen / 32
	xs := make([]*edwards25519.Scalar, count)
	for i := 0; i < count; i++ {
		bs := bs[i*32 : (i+1)*32]
		e, err := edwards25519.NewScalar().SetCanonicalBytes(bs)
		if err != nil {
			return nil, err
		}
		xs[i] = e
	}
	return &FTEd25519IntArray{xs}, nil
}

func (v *FTEd25519IntArray) Single() (*edwards25519.Scalar, error) {
	if len(v.array) != 1 {
		return nil, errors.New("not a singleton array")
	}
	return v.array[0], nil
}
func (v *FTEd25519IntArray) Values() []*edwards25519.Scalar {
	return v.array
}

func Int64ToScalar(x int64) *edwards25519.Scalar {
	xs := make([]byte, 32)

	isNegative := x < 0
	if isNegative {
		x = -x
	}

	binary.LittleEndian.PutUint64(xs, uint64(x))
	s := edwards25519.NewScalar()
	_, err := s.SetCanonicalBytes(xs)
	if err != nil {
		panic(err)
	}

	if isNegative {
		s.Negate(s)
	}

	return s
}

func ScalarToInt64(s *edwards25519.Scalar) (int64, error) {
	bInt := ScalarToBigInt(s)
	if bInt.IsInt64() {
		return bInt.Int64(), nil
	}
	return 0, errors.New("value can not be represented as a 64-bit integer")
}

func BigIntToScalar(x *big.Int) *edwards25519.Scalar {
	bytes := x.Bytes()
	bb := make([]byte, 64)
	for i := range bytes {
		bb[i] = bytes[len(bytes)-i-1]
	}
	s, err := edwards25519.NewScalar().SetUniformBytes(bb)
	if err != nil {
		panic(err)
	}
	return s
}

func ScalarToBigInt(s *edwards25519.Scalar) *big.Int {
	bytes := s.Bytes()
	bb := make([]byte, len(bytes))
	for i := range bytes {
		bb[i] = bytes[len(bytes)-i-1]
	}
	return new(big.Int).SetBytes(bb)
}

func ScalarsToBytes(xs []*edwards25519.Scalar) []byte {
	bytes := make([]byte, len(xs)*32)
	for i, v := range xs {
		offset := i * 32
		bs := v.Bytes()

		for i, b := range bs {
			bytes[offset+i] = b
		}
	}
	return bytes
}

func BytesToScalars(b []byte) ([]*edwards25519.Scalar, error) {
	if len(b)%32 != 0 {
		return nil, errors.New("byte array does not have correct dimensions")
	}
	scalars := make([]*edwards25519.Scalar, len(b)/32)
	for i := 0; i < len(b)/32; i++ {
		x := b[i*32 : (i+1)*32]
		temp, err := edwards25519.NewScalar().SetCanonicalBytes(x)
		if err != nil {
			return nil, err
		}
		scalars[i] = temp
	}
	return scalars, nil
}

func asFTEd25519IntArray(xs ArrayTypeVal) (*FTEd25519IntArray, error) {
	ys, ok := xs.(*FTEd25519IntArray)
	if !ok {
		return nil, fmt.Errorf("value is not an %v", xs.Name())
	}
	return ys, nil
}

func (v *FTEd25519IntArray) Eq(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTEd25519IntArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b *edwards25519.Scalar) int64 { return BToI(a.Equal(b) == 1) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTEd25519IntArray) Ne(other ArrayTypeVal) (*FTIntegerArray, error) {
	bs, err := asFTEd25519IntArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b *edwards25519.Scalar) int64 { return BToI(a.Equal(b) == 0) })
	if err != nil {
		return nil, err
	}

	return &FTIntegerArray{result}, nil
}
func (v *FTEd25519IntArray) Neg() (ArrayNegTypeVal, error) {
	result, err := SliceMapUnary(v.array, func(a *edwards25519.Scalar) *edwards25519.Scalar {
		return edwards25519.NewScalar().Negate(a)
	})
	if err != nil {
		return nil, err
	}

	return &FTEd25519IntArray{result}, nil
}

func (v *FTEd25519IntArray) Add(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTEd25519IntArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b *edwards25519.Scalar) *edwards25519.Scalar { return edwards25519.NewScalar().Add(a, b) })
	if err != nil {
		return nil, err
	}

	return &FTEd25519IntArray{result}, nil
}
func (v *FTEd25519IntArray) Sub(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTEd25519IntArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b *edwards25519.Scalar) *edwards25519.Scalar { return edwards25519.NewScalar().Subtract(a, b) })
	if err != nil {
		return nil, err
	}

	return &FTEd25519IntArray{result}, nil
}
func (v *FTEd25519IntArray) Mul(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTEd25519IntArray(other)
	if err == nil {
		result, err := SliceMapBinary(v.array, bs.array, func(a, b *edwards25519.Scalar) *edwards25519.Scalar { return edwards25519.NewScalar().Multiply(a, b) })

		if err != nil {
			return nil, err
		}

		return &FTEd25519IntArray{result}, nil
	}

	cs, err := asEd25519Array(other)
	if err != nil {
		return nil, errors.New("other was expected to be an Ed25519Int or Ed25519 array")
	}

	return cs.Mul(v)
}

func (v *FTEd25519IntArray) FloorDiv(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTEd25519IntArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinary(v.array, bs.array, func(a, b *edwards25519.Scalar) *edwards25519.Scalar {
		bigZero := big.NewInt(0)
		bigOne := big.NewInt(1)

		bigX := ScalarToBigInt(a)
		bigY := ScalarToBigInt(b)
		bigQ := big.NewInt(0)
		bigR := big.NewInt(0)

		bigQ.DivMod(bigX, bigY, bigR)

		if bigR.Cmp(bigZero) != 0 && bigQ.Cmp(bigZero) == -1 {
			bigQ.Sub(bigQ, bigOne)
		}

		return BigIntToScalar(bigQ)

	})
	if err != nil {
		return nil, err
	}

	return &FTEd25519IntArray{result}, nil
}

func (v *FTEd25519IntArray) TrueDiv(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTEd25519IntArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinaryE(v.array, bs.array, func(a, b *edwards25519.Scalar) (*edwards25519.Scalar, error) {
		z := edwards25519.NewScalar()
		if b.Equal(z) == 1 {
			return z, errors.New("cannot divide by zero")
		}

		inverseB := edwards25519.NewScalar().Invert(b)
		return edwards25519.NewScalar().Multiply(inverseB, a), nil
	})
	if err != nil {
		return nil, err
	}

	return &FTEd25519IntArray{result}, nil
}

func (v *FTEd25519IntArray) Pow(other ArrayTypeVal) (ArrayTypeVal, error) {
	bs, err := asFTIntegerArray(other)
	if err != nil {
		return nil, err
	}

	result, err := SliceMapBinaryE(v.array, bs.array, func(a *edwards25519.Scalar, b int64) (*edwards25519.Scalar, error) {
		if b <= 0 {
			xVal, err := ScalarToInt64(a)
			if err != nil {
				return nil, err
			}

			if b < 0 {
				if xVal == 1 {
					return Int64ToScalar(1), nil
				} else {
					return Int64ToScalar(0), nil
				}
			} else {
				if xVal == 0 && b == 0 {
					return nil, fmt.Errorf("cannot perform 0 to the power of 0")
				}
				return Int64ToScalar(1), nil
			}
		} else {
			res := edwards25519.NewScalar().Set(a)
			for i := 1; i < int(b); i++ {
				res = edwards25519.NewScalar().Multiply(res, a)
			}
			return res, nil
		}
	})

	if err != nil {
		return nil, err
	}

	return &FTEd25519IntArray{result}, nil
}
