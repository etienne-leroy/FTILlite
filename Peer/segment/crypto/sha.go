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
	"golang.org/x/crypto/sha3"
)

func Sha3256Sum(xs [][]byte) [][]byte {
	target := make([][]byte, len(xs))

	for i, v := range xs {
		h := sha3.Sum256(v)
		target[i] = h[:]
	}

	return target
}
