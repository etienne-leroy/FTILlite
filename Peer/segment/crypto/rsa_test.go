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
	"testing"
)

func TestRSA3073PrivateKey_ToFromBytes(t *testing.T) {
	k, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		t.Error(err)
	}

	bs := RSAPrivateKeyToBytes(k)
	k2, err := RSAPrivateKeyFromBytes(bs)
	if err != nil {
		t.Error(err)
	}

	if !k.Equal(k2) {
		t.Error("public keys are not equal after to/from bytes")
	}
}

func TestRSA3073PublicKey_ToFromBytes(t *testing.T) {
	k, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		t.Error(err)
	}

	bs := RSAPublicKeyToBytes(&k.PublicKey)
	k2, err := RSAPublicKeyFromBytes(bs)
	if err != nil {
		t.Error(err)
	}

	if !k.PublicKey.Equal(k2) {
		t.Error("public keys are not equal after to/from bytes")
	}
}
