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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/AUSTRAC/ftillite/Peer/segment/types"
)

func ECDSA256Keygen() ([]byte, error) {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return NewECDSA256PrivateKeyBytes(k), nil
}

func ECDSA256Sign(pk *ecdsa.PrivateKey, data [][]byte) ([][]byte, error) {
	target := make([][]byte, len(data))
	for i, v := range data {
		r, s, err := ecdsa.Sign(rand.Reader, pk, v)
		if err != nil {
			return nil, err
		}
		sig := make([]byte, 64)
		r.FillBytes(sig[:32])
		s.FillBytes(sig[32:])
		target[i] = sig
	}
	return target, nil
}

func ECDSA256Verify(pk *ecdsa.PublicKey, data [][]byte, signatures [][]byte) ([]int64, error) {
	if len(data) != len(signatures) {
		return nil, errors.New("mismatched number of signatures and data")
	}

	target := make([]int64, len(data))
	for i, v := range data {
		r := big.NewInt(0).SetBytes(signatures[i][:32])
		s := big.NewInt(0).SetBytes(signatures[i][32:])

		target[i] = types.BToI(ecdsa.Verify(pk, v, r, s))
	}

	return target, nil
}

// ECDSA256PrivateKeyBytes wraps a []byte containing a serialised ECDSA private key
type ECDSA256PrivateKeyBytes []byte

func NewECDSA256PrivateKeyBytes(k *ecdsa.PrivateKey) ECDSA256PrivateKeyBytes {
	publicBs := elliptic.MarshalCompressed(k.Curve, k.X, k.Y)

	bs := make([]byte, 32+len(publicBs))
	k.D.FillBytes(bs[0:32])
	copy(bs[32:], publicBs)

	return ECDSA256PrivateKeyBytes(bs)
}

func (bs ECDSA256PrivateKeyBytes) PublicKeyBytes() ECDSA256PublicKeyBytes {
	return ECDSA256PublicKeyBytes(bs[32:])
}

func (bs ECDSA256PrivateKeyBytes) Unmarshal() (*ecdsa.PrivateKey, error) {
	d := big.NewInt(0)
	d.SetBytes(bs[:32])

	p, err := ECDSA256PublicKeyBytes(bs[32:]).Unmarshal()

	if err != nil {
		return nil, err
	}

	return &ecdsa.PrivateKey{
		PublicKey: *p,
		D:         d,
	}, nil
}

// ECDSA256PublicKeyBytes wraps a []byte containing a serialised ECDSA public key
type ECDSA256PublicKeyBytes []byte

func (bs ECDSA256PublicKeyBytes) Unmarshal() (*ecdsa.PublicKey, error) {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), bs)

	zero := big.NewInt(0)
	if x.Cmp(zero) == 0 {
		return nil, errors.New("unable to unmarshal public key from bytes")
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}, nil
}
