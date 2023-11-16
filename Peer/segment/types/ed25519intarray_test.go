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
	"math/big"
	"testing"
)

func Test_ScalarToInt64(t *testing.T) {
	var high int64 = (2 << 62) - 1

	bigOne := big.NewInt(1)

	bigHigh := big.NewInt(high)
	scalarHigh := Int64ToScalar(high)
	bigTooHigh := big.NewInt(0).Add(bigHigh, bigOne)
	scalarTooHigh := BigIntToScalar(bigTooHigh)

	x, err := ScalarToInt64(scalarHigh)
	if err != nil {
		t.Error(err)
	}
	if x != high {
		t.Errorf("scalarHigh should be the same as high")
	}

	_, err = ScalarToInt64(scalarTooHigh)
	if err == nil {
		t.Error("scalarTooHigh should not be able to be represented as int64")
	}
}
