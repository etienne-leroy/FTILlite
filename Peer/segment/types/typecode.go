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
	"regexp"
	"strconv"
)

type BaseTypeCode int32

const (
	IntegerB    BaseTypeCode = 'i'
	FloatB      BaseTypeCode = 'f'
	Ed25519IntB BaseTypeCode = 'I'
	Ed25519B    BaseTypeCode = 'E'
	BytearrayB  BaseTypeCode = 'b'
)

func (b BaseTypeCode) String() string { return string(b) }

type TypeCode string

var TypeCodeRegEx = regexp.MustCompile(`([ifIE]|b[1-9][0-9]*)\s*`)

func ParseTypeCode(v string) (TypeCode, error) {
	if m, _ := regexp.MatchString("^[ifIE]|b\\d+$", v); m {
		return TypeCode(v), nil
	}
	return Integer, fmt.Errorf("unknown typecode: %v", v)
}

func ParseTypeCodes(value string) []TypeCode {
	xs := TypeCodeRegEx.FindAllString(value, -1)
	result := make([]TypeCode, len(xs))
	for i, x := range xs {
		result[i] = TypeCode(x)
	}
	return result
}

const (
	Integer    TypeCode = TypeCode(IntegerB)
	Float      TypeCode = TypeCode(FloatB)
	Ed25519Int TypeCode = TypeCode(Ed25519IntB)
	Ed25519    TypeCode = TypeCode(Ed25519B)
)

func Bytearray(length int) TypeCode {
	return TypeCode(fmt.Sprintf("%v%v", BytearrayB, length))
}
func (tc TypeCode) GetBase() BaseTypeCode { return BaseTypeCode(tc[0]) }
func (tc TypeCode) IsBytearray() bool     { return tc.GetBase() == BytearrayB }
func (tc TypeCode) Length() int {
	if !tc.IsBytearray() {
		switch tc.GetBase() {
		case 'i':
			return 8
		case 'f':
			return 8
		case 'E':
			return 64
		case 'I':
			return 32
		default:
			panic("no valid typecode found for length method")
		}
	}
	l, err := strconv.Atoi(string(tc[1:]))
	if err != nil {
		panic("invalid typecode: " + err.Error())
	}
	return l
}
func (tc TypeCode) GetTypeCodeAsSlice() []TypeCode {
	return ParseTypeCodes(string(tc))
}
