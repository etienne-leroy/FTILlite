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
	CommandECDSA256Keygen    = "command_ecdsa256_keygen"
	CommandECDSA256PublicKey = "command_ecdsa256_public_key"
	CommandECDSA256Sign      = "command_ecdsa256_sign"
	CommandECDSA256Verify    = "command_ecdsa256_verify"
)

func ECDSA256Keygen(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])

	targetBytes, err := crypto.ECDSA256Keygen()
	if err != nil {
		return "", nil
	}

	target, err := types.NewFTBytearrayArray(len(targetBytes), targetBytes)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b%v %v", target.Width(), hTarget), nil
}

func ECDSA256PublicKey(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hPrivateKey := variables.Handle(args[1])

	pkBytes, err := variables.GetAsBytes(s.Variables(), hPrivateKey)
	if err != nil {
		return "", err
	}

	pk := crypto.ECDSA256PrivateKeyBytes(pkBytes).PublicKeyBytes()

	target, err := types.NewFTBytearrayArray(len(pk), pk)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b%v %v", target.Width(), hTarget), nil
}

func ECDSA256Sign(s SegmentHost, args []string) (string, error) {
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

	pk, err := crypto.ECDSA256PrivateKeyBytes(pkBytes).Unmarshal()
	if err != nil {
		return "", err
	}

	targetBytes, err := crypto.ECDSA256Sign(pk, data.Values())
	if err != nil {
		return "", err
	}

	target, err := types.NewFTBytearrayArray(types.CalcBytearrayWidth(targetBytes), targetBytes...)
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, target)

	return fmt.Sprintf("array b64 %v", hTarget), nil
}

func ECDSA256Verify(s SegmentHost, args []string) (string, error) {
	hTarget := variables.Handle(args[0])
	hData := variables.Handle(args[1])
	hSignatures := variables.Handle(args[2])
	hPublicKey := variables.Handle(args[3])

	data, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hData)
	if err != nil {
		return "", err
	}

	sigs, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), hSignatures)
	if err != nil {
		return "", err
	}

	pkBytes, err := variables.GetAsBytes(s.Variables(), hPublicKey)
	if err != nil {
		return "", err
	}

	pk, err := crypto.ECDSA256PublicKeyBytes(pkBytes).Unmarshal()
	if err != nil {
		return "", err
	}

	target, err := crypto.ECDSA256Verify(pk, data.Values(), sigs.Values())
	if err != nil {
		return "", err
	}

	s.Variables().Set(hTarget, types.NewFTIntegerArray(target...))

	return fmt.Sprintf("array i %v", hTarget), nil
}
