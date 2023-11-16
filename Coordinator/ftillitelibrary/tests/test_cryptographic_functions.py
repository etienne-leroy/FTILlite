# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from .util import fc
import ftillite as fl

def test_sha3_256(fc):
    v1 = fc.array("i", [1, 2, 3])
    v2 = v1.astype("b8")
    v3 = v1.astype("b32")

    v2_hash = fl.sha3_256(v2)
    assert v2_hash.typecode() == "b32"
    
    v3_hash = fl.sha3_256(v3)
    assert v3_hash.typecode() == "b32"
    assert fl.verify(v3_hash != v3, False)

def test_aes256_encrypt_decrypt(fc):
    key = fc.aes256_keygen()
    data = fc.randomarray("b16", 2)

    encryptedData = fl.aes256_encrypt(data, key)
    assert fl.verify(data != encryptedData, False)

    decryptedData = fl.aes256_decrypt(encryptedData, key)
    assert fl.verify(data == decryptedData, False)

def test_grain128aeadv2_keygen(fc):
    key = fc.grain128aeadv2_keygen()
    key2 = fc.grain128aeadv2_keygen()

    assert key.typecode() == "b16"
    assert key2.typecode() == "b16"
    assert fl.verify(key.len() == 1, False)
    assert fl.verify(key2.len() == 1, False)
    assert fl.verify(key != key2, False)

def test_grain128aeadv2(fc):
    key = fc.grain128aeadv2_keygen()
    key2 = fc.grain128aeadv2_keygen()

    iv = fc.randomarray("b12", 1)
    iv2 = fc.randomarray("b12", 1)

    size = 32
    length = fc.array("i", [2])

    result = fl.grain128aeadv2(key, iv, size, length)
    assert result.typecode() == "b32"

    result2 = fl.grain128aeadv2(key, iv, size, length)
    assert fl.verify(result == result2, False)

    result3 = fl.grain128aeadv2(key, iv2, size, length)
    assert fl.verify(result != result3, False)

    result4 = fl.grain128aeadv2(key2, iv, size, length)
    assert fl.verify(result != result4, False)

def test_ecdsa_keygen(fc):
    (privateKey, publicKey) = fc.ecdsa256_keygen()

    assert privateKey.typecode() == "b65"
    assert publicKey.typecode() == "b33"

def test_ecdsa_sign_verify(fc):
    (privateKey, publicKey) = fc.ecdsa256_keygen()
    data = fc.array("i", [1, 2, 3, 4, 5]).astype("b16")
    sigs = fl.ecdsa256_sign(data, privateKey)

    data2 = fc.array("i", [1, 2, 3, 4, 5]).astype("b16")
    result = fl.ecdsa256_verify(data2, sigs, publicKey)

    assert fl.verify(result, False)

def test_rsa_keygen(fc):
    (privateKey, publicKey) = fc.rsa3072_keygen()

    assert privateKey.typecode() == "b2720"
    assert publicKey.typecode() == "b400"

def test_rsa_encrypt_decrypt(fc):
    (privateKey, publicKey) = fc.rsa3072_keygen()

    data = fc.randomarray("b48", 10)

    encryptedData = fl.rsa3072_encrypt(data, publicKey)

    # Can't compare data and encryptedData as the width is different, but asserting 
    # the width is not 48 should be enough to know that the arrays are different. 
    assert encryptedData.width() != 48

    decryptedData = fl.rsa3072_decrypt(encryptedData, privateKey)
    assert decryptedData == data