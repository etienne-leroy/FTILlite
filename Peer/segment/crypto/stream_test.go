// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package crypto

import (
	"math/big"
	"testing"
)

func TestByteWriter_IntRW(t *testing.T) {

	bs := make([]byte, 24)
	w := Stream{bs, 0}

	i1 := 15
	i2 := 234
	i3 := 955543

	w.Write(AsInteger{&i1})
	w.Write(AsInteger{&i2})
	w.Write(AsInteger{&i3})

	i1a := 0
	i2a := 0
	i3a := 0
	w.pos = 0

	w.Read(AsInteger{&i1a})
	w.Read(AsInteger{&i2a})
	w.Read(AsInteger{&i3a})

	if i1 != i1a || i2 != i2a || i3 != i3a {
		t.Fail()
	}
}

func TestByteWriter_BigIntRW(t *testing.T) {
	size := 32
	bs := make([]byte, 96)
	w := Stream{bs, 0}

	i1 := big.NewInt(15)
	i2 := big.NewInt(234)
	i3 := big.NewInt(955543)

	w.Write(AsBigInteger{size, i1})
	w.Write(AsBigInteger{size, i2})
	w.Write(AsBigInteger{size, i3})

	i1a := big.NewInt(0)
	i2a := big.NewInt(0)
	i3a := big.NewInt(0)
	w.pos = 0

	w.Read(AsBigInteger{size, i1a})
	w.Read(AsBigInteger{size, i2a})
	w.Read(AsBigInteger{size, i3a})

	if i1.Cmp(i1a) != 0 || i2.Cmp(i2a) != 0 || i3.Cmp(i3a) != 0 {
		t.Fail()
	}
}
