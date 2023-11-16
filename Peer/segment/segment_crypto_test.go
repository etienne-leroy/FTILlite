// =====================================
//
// Copyright (c) 2023, AUSTRAC Australian Government
// All rights reserved.
//
// Licensed under BSD 3 clause license
//
// #####################################

package segment

import (
	"testing"

	"github.com/AUSTRAC/ftillite/Peer/segment/commands"
	"github.com/AUSTRAC/ftillite/Peer/segment/types"
	"github.com/AUSTRAC/ftillite/Peer/segment/variables"
)

func TestCommandSHA3_256(t *testing.T) {
	s := NewTestSegment()

	s.SetVariable("1", types.NewFTBytearrayArrayOrPanic(
		4,
		[]byte{1, 1, 1, 1},
		[]byte{0, 0, 0, 0},
		[]byte{1, 0, 1, 0},
	))

	AssertCommand(t, s, commands.CommandSHA3256, "2", "1")

	AssertValue(t, s, "2", types.NewFTBytearrayArrayOrPanic(
		32,
		[]byte{64, 29, 249, 80, 74, 37, 79, 121, 48, 205, 253, 191, 190, 26, 83, 131, 100, 65, 14, 9, 210, 119, 186, 63, 156, 115, 151, 3, 29, 188, 76, 239},
		[]byte{139, 10, 35, 133, 216, 60, 139, 247, 190, 39, 229, 153, 150, 247, 216, 129, 211, 191, 31, 198, 96, 111, 129, 206, 96, 11, 117, 58, 217, 65, 146, 162},
		[]byte{68, 178, 228, 82, 170, 50, 150, 94, 215, 127, 7, 19, 123, 196, 43, 136, 250, 57, 114, 47, 3, 80, 95, 118, 115, 24, 174, 215, 100, 178, 35, 40},
	))
}

func TestCommand_AES256_Encrypt_Decrypt(t *testing.T) {
	s := NewTestSegment()

	AssertCommand(t, s, commands.CommandNewilist, "1", "1") // count
	AssertCommand(t, s, commands.CommandRandomArray, "2", "b32", "1")

	original := types.NewFTBytearrayArrayOrPanic(
		16,
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		[]byte{16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31},
	)
	s.SetVariable("3", original)

	AssertCommand(t, s, commands.CommandAES256Encrypt, "4", "3", "2")

	// Assert encrypted data not equal to original data
	AssertValueNot(t, s, "4", original)

	AssertCommand(t, s, commands.CommandAES256Decrypt, "6", "4", "2")

	// Assert decrypted data equal to original data
	AssertValue(t, s, "6", original)
}

func TestCommand_Grain128aeadv2(t *testing.T) {
	s := NewTestSegment()
	PrintVariableStore(t, s)

	AssertCommand(t, s, commands.CommandNewilist, "1", "1")
	AssertCommand(t, s, commands.CommandRandomArray, "2", "b16", "1") // key
	AssertCommand(t, s, commands.CommandRandomArray, "3", "b12", "1") // iv
	AssertCommand(t, s, commands.CommandNewilist, "4", "32")          // size
	AssertCommand(t, s, commands.CommandNewilist, "5", "4")           // length

	AssertCommand(t, s, commands.CommandGrain128aeadv2, "6", "2", "3", "4", "5")
}

func TestCommand_ECDSA256Keygen(t *testing.T) {
	s := NewTestSegment()
	PrintVariableStore(t, s)

	AssertCommand(t, s, commands.CommandECDSA256Keygen, "1")
}
func TestCommand_ECDSA256PublicKey(t *testing.T) {
	s := NewTestSegment()
	PrintVariableStore(t, s)

	AssertCommand(t, s, commands.CommandECDSA256Keygen, "1")
	AssertCommand(t, s, commands.CommandECDSA256PublicKey, "2", "1")
}

func TestCommand_ECDSA256Sign_Verify_Correct(t *testing.T) {
	s := NewTestSegment()
	PrintVariableStore(t, s)

	AssertCommand(t, s, commands.CommandECDSA256Keygen, "1")
	AssertCommand(t, s, commands.CommandECDSA256PublicKey, "2", "1")

	AssertCommand(t, s, commands.CommandNewilist, "3", "5")
	AssertCommand(t, s, commands.CommandRandomArray, "4", "b16", "3")

	AssertCommand(t, s, commands.CommandECDSA256Sign, "5", "4", "1")

	AssertCommand(t, s, commands.CommandECDSA256Verify, "6", "4", "5", "2")

	AssertValue(t, s, "6", types.NewFTIntegerArray(1, 1, 1, 1, 1))
}

func TestCommand_ECDSA256Sign_Verify_Incorrect(t *testing.T) {
	s := NewTestSegment()
	PrintVariableStore(t, s)

	AssertCommand(t, s, commands.CommandECDSA256Keygen, "1")
	AssertCommand(t, s, commands.CommandECDSA256PublicKey, "2", "1")

	AssertCommand(t, s, commands.CommandNewilist, "3", "5")
	AssertCommand(t, s, commands.CommandRandomArray, "4", "b16", "3")

	AssertCommand(t, s, commands.CommandECDSA256Sign, "5", "4", "1")

	bs, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), "4")
	if err != nil {
		t.Error(err)
	}
	bs.Values()[1][1] = bs.Values()[1][1] - 1

	AssertCommand(t, s, commands.CommandECDSA256Verify, "6", "4", "5", "2")
	AssertValue(t, s, "6", types.NewFTIntegerArray(1, 0, 1, 1, 1))
}
func TestCommand_RSA30726Keygen(t *testing.T) {
	s := NewTestSegment()
	PrintVariableStore(t, s)

	AssertCommand(t, s, commands.CommandRSA3072Keygen, "1")
}
func TestCommand_RSA3072PublicKey(t *testing.T) {
	s := NewTestSegment()
	PrintVariableStore(t, s)

	AssertCommand(t, s, commands.CommandRSA3072Keygen, "1")
	AssertCommand(t, s, commands.CommandRSA3072PublicKey, "2", "1")
}

func TestCommand_RSA3072Encrypt_Decrypt(t *testing.T) {
	s := NewTestSegment()
	PrintVariableStore(t, s)

	// Singleton array for length of random bytearray
	AssertCommand(t, s, commands.CommandNewilist, "0", "10")

	// Generate keys
	AssertCommand(t, s, commands.CommandRSA3072Keygen, "1")
	AssertCommand(t, s, commands.CommandRSA3072PublicKey, "2", "1")

	// Generate some random data to encrypt
	AssertCommand(t, s, commands.CommandRandomArray, "3", "b32", "0")
	original, err := variables.GetAs[*types.FTBytearrayArray](s.Variables(), "3")
	if err != nil {
		t.Fatal(err)
	}

	// Encrypt the data
	AssertCommand(t, s, commands.CommandRSA3072Encrypt, "4", "3", "2")

	// Assert encrypted data not equal to original data
	AssertValueNot(t, s, "4", original)

	// Decrypt the data
	AssertCommand(t, s, commands.CommandRSA3072Decrypt, "5", "4", "1")

	// Assert decrypted data equal to original data
	AssertValue(t, s, "5", original)
}
