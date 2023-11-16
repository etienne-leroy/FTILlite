// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package commands

import (
	"errors"
	"fmt"

	"github.com/AUSTRAC/ftillite/Peer/segment/crypto"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const (
	CommandAES256Encrypt = "command_aes256_encrypt"
	CommandAES256Decrypt = "command_aes256_decrypt"
)

func aes256(s SegmentHost, args []string, f func(key []byte, data [][]byte) ([][]byte, error)) (string, error) {
	hTarget := variables.Handle(args[0])
	hData := variables.Handle(args[1])
	hKey := variables.Handle(args[2])

	keyArray, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hKey)
	if err != nil {
		return "", err
	}
	key, err := keyArray.Single()
	if err != nil {
		return "", err
	}

	data, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hData)
	if err != nil {
		return "", err
	}
	if data.TypeCode().Length() != 16 {
		return "", errors.New("data must be b16")
	}

	targetBytes, err := f(key, data.Values())
	if err != nil {
		return "", err
	}

	target, err := types.NewFTBytearrayArray(16, targetBytes...)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b16 %v", hTarget), nil
}

func Aes256Encrypt(s SegmentHost, args []string) (string, error) {
	return aes256(s, args, func(key []byte, data [][]byte) ([][]byte, error) {
		return crypto.Aes256Encrypt(key, data)
	})
}
func Aes256Decrypt(s SegmentHost, args []string) (string, error) {
	return aes256(s, args, func(key []byte, data [][]byte) ([][]byte, error) {
		return crypto.Aes256Decrypt(key, data)
	})
}
