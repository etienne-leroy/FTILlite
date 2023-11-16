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
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
)

func aes256(key []byte, data [][]byte, f func(cipher cipher.Block, dst []byte, src []byte)) ([][]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must have a length of 32 bytes")
	}

	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("unable to create AES cypher with key: %e", err)
	}
	target := make([][]byte, len(data))

	for i, v := range data {
		t := make([]byte, len(data[i]))
		f(cipher, t, v)
		target[i] = t
	}

	return target, nil
}

func Aes256Encrypt(key []byte, data [][]byte) ([][]byte, error) {
	return aes256(key, data, func(cipher cipher.Block, dst []byte, src []byte) {
		cipher.Encrypt(dst, src)
	})
}
func Aes256Decrypt(key []byte, data [][]byte) ([][]byte, error) {
	return aes256(key, data, func(cipher cipher.Block, dst []byte, src []byte) {
		cipher.Decrypt(dst, src)
	})
}
