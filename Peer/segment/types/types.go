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
)

// TypeVal is implemented by all ftillite types
type TypeVal interface {
	Name() string
	TypeCode() TypeCode

	Equals(other TypeVal) bool
	GetBinaryArray(index int) ([]byte, error)

	EstimatedSize() int64
	DebugString() string
}

// ArrayTypeVal is implemented by all ftillite array types
type ArrayTypeVal interface {
	TypeVal

	Length() int64

	Clone() (TypeVal, error)
	AsType(tc TypeCode) (TypeVal, error)

	Get(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error)
	Lookup(indexes *FTIntegerArray, defaultValue ArrayTypeVal) (ArrayTypeVal, error)
	Set(indexes *FTIntegerArray, values ArrayTypeVal) error
	SetLength(x int64) error
	Broadcast(length int64) error
	Remove(indexes *FTIntegerArray) error
	Index() *FTIntegerArray
	Contains(values ArrayTypeVal) (*FTIntegerArray, error)

	ReduceSum(indexes *FTIntegerArray, values ArrayTypeVal) error
	ReduceISum(indexes *FTIntegerArray, values ArrayTypeVal) error

	CumSum() (ArrayTypeVal, error)
	Mux(condition *FTIntegerArray, ifFalse ArrayTypeVal) (ArrayTypeVal, error)

	Eq(other ArrayTypeVal) (*FTIntegerArray, error)
	Ne(other ArrayTypeVal) (*FTIntegerArray, error)
}

// ArrayComparableTypeVal is implemented by ftillite arrays which support the comparison operators
type ArrayComparableTypeVal interface {
	ArrayTypeVal

	Gt(other ArrayTypeVal) (*FTIntegerArray, error)
	Lt(other ArrayTypeVal) (*FTIntegerArray, error)
	Ge(other ArrayTypeVal) (*FTIntegerArray, error)
	Le(other ArrayTypeVal) (*FTIntegerArray, error)

	ReduceMax(indexes *FTIntegerArray, values ArrayTypeVal) error
	ReduceIMax(indexes *FTIntegerArray, values ArrayTypeVal) error
	ReduceMin(indexes *FTIntegerArray, values ArrayTypeVal) error
	ReduceIMin(indexes *FTIntegerArray, values ArrayTypeVal) error
}
type ArrayNegTypeVal interface {
	ArrayTypeVal

	Neg() (ArrayNegTypeVal, error)
}
type ArrayAbsTypeVal interface {
	ArrayTypeVal

	Abs() (ArrayAbsTypeVal, error)
}
type ArrayAddSubMulTypeVal interface {
	ArrayTypeVal

	Add(other ArrayTypeVal) (ArrayTypeVal, error)
	Sub(other ArrayTypeVal) (ArrayTypeVal, error)
	Mul(other ArrayTypeVal) (ArrayTypeVal, error)
}
type ArrayFloorDivTypeVal interface {
	ArrayTypeVal

	FloorDiv(other ArrayTypeVal) (ArrayTypeVal, error)
}
type ArrayTrueDivTypeVal interface {
	ArrayTypeVal

	TrueDiv(other ArrayTypeVal) (ArrayTypeVal, error)
}
type ArrayPowTypeVal interface {
	ArrayTypeVal

	Pow(other ArrayTypeVal) (ArrayTypeVal, error)
}

type ArrayBitwiseTypeVal interface {
	ArrayTypeVal

	LShift(other ArrayTypeVal) (ArrayTypeVal, error)
	RShift(other ArrayTypeVal) (ArrayTypeVal, error)
	And(other ArrayTypeVal) (ArrayTypeVal, error)
	Or(other ArrayTypeVal) (ArrayTypeVal, error)
	Xor(other ArrayTypeVal) (ArrayTypeVal, error)
	Invert() (ArrayTypeVal, error)
}

// ArrayElementTypeVal is implemented by arrays which are backed by Go arrays
type ArrayElementTypeVal interface {
	ArrayTypeVal

	Element(i int64) interface{}
}

// Freer is used by the variable store and implemented by TypeVals which need to
// clean up resources when they are no longer referenced by any variables in the
// variable store. The reference count is incremeneted when a variable is set to
// the TypeVal instance, and decremented when the variable is deleted or changed
// to a different TypeVal. When the reference count reaches 0, the variable store
// will call Free to clean up the resources.
//
// Note: implementers of the Freer interface do not need to ensure thread-safety, this
// is done by the variable store itself. The variable store guarentees that there is ever
// only one Go routine setting/deleting variables at any one time.
type Freer interface {
	ReferenceCount() int
	IncrementReferenceCount()
	Free()
}

func FromBytes(t TypeCode, bs []byte) (ArrayTypeVal, error) {
	switch t.GetBase() {
	case IntegerB:
		return NewFTIntegerArrayFromBytes(bs)
	case FloatB:
		return NewFTFloatArrayFromBytes(bs)
	case BytearrayB:
		return NewFTBytearrayArrayFromBytes(bs, int64(t.Length()))
	case Ed25519IntB:
		return NewFTEd25519IntArrayFromBytes(bs)
	case Ed25519B:
		return NewEd25519ArrayFromBytes(bs)
	default:
		return nil, fmt.Errorf("unrecognised type code: %v", t)
	}
}

func PrintSize(b uint64) string {
	if b < 1024 {
		return fmt.Sprintf("%v B", b)
	} else if b < (1024 * 1024) {
		return fmt.Sprintf("%v KB", b/1024)
	} else {
		return fmt.Sprintf("%v MB", b/(1024*1024))
	}
}

func BToI(b bool) int64 {
	if b {
		return 1
	} else {
		return 0
	}
}
