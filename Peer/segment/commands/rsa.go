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
	"fmt"

	"github.com/AUSTRAC/ftillite/Peer/segment/crypto"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

const (
	CommandRSA3072Keygen    = "command_rsa3072_keygen"
	CommandRSA3072PublicKey = "command_rsa3072_public_key"
	CommandRSA3072Encrypt   = "command_rsa3072_encrypt"
	CommandRSA3072Decrypt   = "command_rsa3072_decrypt"
)

func RSA3072Keygen(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])

	bs, err := crypto.RSA3072Keygen()
	if err != nil {
		return "", err
	}

	target, err := types.NewFTBytearrayArray(len(bs), bs)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b%v %v", target.Width(), hTarget), nil
}

func RSA3072PublicKey(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hPrivateKey := variables.Handle(args[1])

	xs, err := variables.GetAsBytes(s.Variables(), hPrivateKey)
	if err != nil {
		return "", err
	}

	pk, err := crypto.RSAPrivateKeyFromBytes(xs)
	if err != nil {
		return "", err
	}

	bs := crypto.RSAPublicKeyToBytes(&pk.PublicKey)

	target, err := types.NewFTBytearrayArray(len(bs), bs)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b%v %v", target.Width(), hTarget), nil
}

func RSA3072Encrypt(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hData := variables.Handle(args[1])
	hPublicKey := variables.Handle(args[2])

	data, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hData)
	if err != nil {
		return "", err
	}

	pkBytes, err := variables.GetAsBytes(s.Variables(), hPublicKey)
	if err != nil {
		return "", err
	}

	pk, err := crypto.RSAPublicKeyFromBytes(pkBytes)
	if err != nil {
		return "", err
	}

	targetBytes, err := crypto.RSA3072Encrypt(pk, data.Values())
	if err != nil {
		return "", err
	}

	target, err := types.NewFTBytearrayArray(types.CalcBytearrayWidth(targetBytes), targetBytes...)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b%v %v", target.Width(), hTarget), nil
}

func RSA3072Decrypt(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hData := variables.Handle(args[1])
	hPrivateKey := variables.Handle(args[2])

	data, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hData)
	if err != nil {
		return "", err
	}

	pkBytes, err := variables.GetAsBytes(s.Variables(), hPrivateKey)
	if err != nil {
		return "", err
	}

	pk, err := crypto.RSAPrivateKeyFromBytes(pkBytes)
	if err != nil {
		return "", err
	}

	targetBytes, err := crypto.RSA3072Decrypt(pk, data.Values())
	if err != nil {
		return "", err
	}

	target, err := types.NewFTBytearrayArray(types.CalcBytearrayWidth(targetBytes), targetBytes...)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b%v %v", target.Width(), hTarget), nil
}
