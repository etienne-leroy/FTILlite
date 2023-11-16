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
	"errors"
	"fmt"

	"github.com/ericlagergren/lwcrypto/grain"
)

func Grain128Aeadv2(key []byte, iv []byte, width int64, length int64) ([][]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("key must have a length of 16")
	}
	if len(iv) != 12 {
		return nil, errors.New("iv must have a length of 12")
	}

	cipherStream, err := grain.NewUnauthenticated(key, iv)
	if err != nil {
		return nil, fmt.Errorf("unable to create Grain128-AEAD cypher: %v", err)
	}
	target := make([][]byte, length)

	for i := range target {
		t := make([]byte, width)

		cipherStream.XORKeyStream(t, t)

		target[i] = t
	}

	return target, nil
}
