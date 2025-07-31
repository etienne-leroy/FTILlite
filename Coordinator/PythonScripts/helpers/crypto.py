"""
Cryptographic Helper Functions for FinTracer
"""
import os
import sys
import ftillite as fl
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../TypeDefinitions'))
from elgamal import elgamal_encrypt, elgamal_keygen, ElGamalCipher #type: ignore


def setup_crypto(fc):
    """Setup cryptographic keys and zero values"""
    with fl.on(fc.CoordinatorID):
        (private_key, local_pub_key) = elgamal_keygen(fc)

    public_key = fl.transmit({i : local_pub_key for i in fc.scope()})[fc.CoordinatorID]
    encrypted_zero = ElGamalCipher(fc)
    encrypted_zero.set_length(1)

    with fl.on(fc.CoordinatorID):
        local_plain_zero = encrypted_zero.decrypt(private_key)

    plain_zero = fl.transmit({n : local_plain_zero for n in fc.scope()})[fc.CoordinatorID]
    default_zero = elgamal_encrypt( fc.array( "i", 1, 0 ), public_key )
    return private_key, public_key, encrypted_zero, default_zero, plain_zero


