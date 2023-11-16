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
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"math/big"
)

const RSAPrivateKeyBytearraySize = 2720
const RSAPublicKeyBytearraySize = 400

func RSA3072Keygen() ([]byte, error) {
	k, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return nil, err
	}
	return RSAPrivateKeyToBytes(k), nil
}

func RSA3072Encrypt(pk *rsa.PublicKey, data [][]byte) ([][]byte, error) {
	xs := make([][]byte, len(data))
	for i, v := range data {
		ve, err := rsa.EncryptPKCS1v15(rand.Reader, pk, v)
		if err != nil {
			return nil, err
		}
		xs[i] = ve
	}
	return xs, nil
}
func RSA3072Decrypt(pk *rsa.PrivateKey, data [][]byte) ([][]byte, error) {
	xs := make([][]byte, len(data))
	for i, v := range data {
		ve, err := rsa.DecryptPKCS1v15(rand.Reader, pk, v)
		if err != nil {
			return nil, err
		}
		xs[i] = ve
	}
	return xs, nil
}

func RSAPrivateKeyToBytes(k *rsa.PrivateKey) []byte {
	// Why not use crypto/x509.MarshalPKCS*() ?
	// - it returns a variable sized byte array (even for the same sized rsa key). This causes
	//   problems with variable validation on the Coordinator since all nodes should have the
	//   same sized bytearray.
	// - Marshalling throws away the precomputed values, so we will incur a cost everytime the
	//   private key is used.
	//
	// Instead we write the various components of the key to a byte array directly, and then
	// populate those fields when reading back in.

	k.Precompute()
	bigIntByteSize := k.Size()

	primesCount := len(k.Primes)
	crtValuesCount := len(k.Precomputed.CRTValues)

	rws := make([]ReaderWriter, 0, 8+primesCount+crtValuesCount)

	rws = append(
		rws,
		AsInteger{&bigIntByteSize},
		AsBigInteger{bigIntByteSize, k.PublicKey.N},
		AsInteger{&k.PublicKey.E},
		AsBigInteger{bigIntByteSize, k.D},
		AsInteger{&primesCount},
	)

	for _, v := range k.Primes {
		rws = append(rws, AsBigInteger{bigIntByteSize, v})
	}

	rws = append(
		rws,
		AsBigInteger{bigIntByteSize, k.Precomputed.Dp},
		AsBigInteger{bigIntByteSize, k.Precomputed.Dq},
		AsBigInteger{bigIntByteSize, k.Precomputed.Qinv},
		AsInteger{&crtValuesCount},
	)

	for _, v := range k.Precomputed.CRTValues {
		rws = append(
			rws,
			AsBigInteger{bigIntByteSize, v.Coeff},
			AsBigInteger{bigIntByteSize, v.Exp},
			AsBigInteger{bigIntByteSize, v.R},
		)
	}

	return WriteToBytes(rws...)
}

func RSAPrivateKeyFromBytes(bs []byte) (*rsa.PrivateKey, error) {
	if len(bs) != RSAPrivateKeyBytearraySize {
		return nil, fmt.Errorf("byte array length must be %v, but got %v", RSAPrivateKeyBytearraySize, len(bs))
	}

	var bigIntByteSize int
	k := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: big.NewInt(0),
			E: 0,
		},
		D: big.NewInt(0),
		Precomputed: rsa.PrecomputedValues{
			Dq:   big.NewInt(0),
			Dp:   big.NewInt(0),
			Qinv: big.NewInt(0),
		},
	}

	primesCount := 0
	crtValuesCount := 0

	w := &Stream{bs, 0}

	w.Read(AsInteger{&bigIntByteSize})
	w.Read(AsBigInteger{bigIntByteSize, k.PublicKey.N})
	w.Read(AsInteger{&k.PublicKey.E})
	w.Read(AsBigInteger{bigIntByteSize, k.D})
	w.Read(AsInteger{&primesCount})

	k.Primes = make([]*big.Int, primesCount)
	for i := 0; i < primesCount; i++ {
		k.Primes[i] = big.NewInt(0)
		w.Read(AsBigInteger{bigIntByteSize, k.Primes[i]})
	}

	w.Read(AsBigInteger{bigIntByteSize, k.Precomputed.Dp})
	w.Read(AsBigInteger{bigIntByteSize, k.Precomputed.Dq})
	w.Read(AsBigInteger{bigIntByteSize, k.Precomputed.Qinv})
	w.Read(AsInteger{&crtValuesCount})

	k.Precomputed.CRTValues = make([]rsa.CRTValue, crtValuesCount)
	for i := 0; i < crtValuesCount; i++ {
		k.Precomputed.CRTValues[i] = rsa.CRTValue{
			Coeff: big.NewInt(0),
			Exp:   big.NewInt(0),
			R:     big.NewInt(0),
		}
		w.Read(AsBigInteger{bigIntByteSize, k.Precomputed.CRTValues[i].Coeff})
		w.Read(AsBigInteger{bigIntByteSize, k.Precomputed.CRTValues[i].Exp})
		w.Read(AsBigInteger{bigIntByteSize, k.Precomputed.CRTValues[i].R})
	}

	err := k.Validate()
	if err != nil {
		return nil, err
	}

	return k, nil
}

func RSAPublicKeyToBytes(k *rsa.PublicKey) []byte {
	bigIntByteSize := k.Size()

	return WriteToBytes(
		AsInteger{&bigIntByteSize},
		AsBigInteger{bigIntByteSize, k.N},
		AsInteger{&k.E},
	)
}

func RSAPublicKeyFromBytes(bs []byte) (*rsa.PublicKey, error) {
	if len(bs) != RSAPublicKeyBytearraySize {
		return nil, fmt.Errorf("byte array length must be %v, but got %v", RSAPublicKeyBytearraySize, len(bs))
	}

	var bigIntByteSize int
	k := &rsa.PublicKey{
		N: big.NewInt(0),
		E: 0,
	}

	w := &Stream{bs, 0}

	w.Read(AsInteger{&bigIntByteSize})
	w.Read(AsBigInteger{bigIntByteSize, k.N})
	w.Read(AsInteger{&k.E})

	return k, nil
}
