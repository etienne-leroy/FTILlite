# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

import ftillite as fl

from pair import *

print("WARNING: Importing file mock_elgamal.py. These classes DO NOT PERFORM ANY ENCRYPTION. *NEVER* USE THESE CLASSES IN PRODUCTION CODE.")

class ElGamalCipher(Pair):
  def __init__(self, mask = None, masked_message = None):
    if type(mask) is fl.FTILContext and masked_message is None:
        mask = mask.array('i')
        masked_message = mask[:]
    if type(mask) is not fl.IntArrayIdentifier or \
        type(masked_message) is not fl.IntArrayIdentifier:
      raise TypeError("Unexpected argument types constructing ElGamalCipher.")
    super().__init__(mask, masked_message)

  def transmitter(self):
    return ElGamalCipherTransmitter(self)

  def stub(self):
    return ElGamalCipher(self.first.stub(), self.second.stub())

  def sametype(self, other):
    return type(other) is ElGamalCipher and \
             self.first.sametype(other.first) and \
             self.second.sametype(other.second)

  def __eq__(self, other):
    if type(other) is not ElGamalCipher:
        return 0
    return (self.first == other.first) & (self.second == other.second)

  def __ne__(self, other):
    return 1 - (self == other)

  def copy(self):
    return ElGamalCipher(self.first, self.second)

  def promote(self, typecode):
    if self.typecode() != typecode:
      raise TypeError("Cannot promote a (Mock)ElGamalCipher")
    return (self, False)

  def astype(self, typecode):
    if self.typecode() != typecode:
      raise TypeError("Cannot convert a (Mock)ElGamalCipher")
    return self.copy()

  def __getitem__(self, key):
    return ElGamalCipher(self.first[key], self.second[key])

  def lookup(self, key, default = None):
    if default is None:
      return ElGamalCipher(self.first.lookup(key), self.second.lookup(key))
    elif self.sametype(default):
      return ElGamalCipher(self.first.lookup(key, default.first), \
                             self.second.lookup(key, default.second))
    else:
      raise TypeError("Unexpected type for default.")

  def reduce_sum(self, key, value):
    if not self.sametype(value):
      raise TypeError("Value is not the same type as LHS.")
    self.first.reduce_sum(key, value.first)
    self.second.reduce_sum(key, value.second)

  def reduce_isum(self, key, value):
    if not self.sametype(value):
      raise TypeError("Value is not the same type as LHS.")
    self.first.reduce_isum(key, value.first)
    self.second.reduce_isum(key, value.second)

  def cumsum(self):
    self.first.cumsum()
    self.second.cumsum()

  def __setitem__(self, key, value):
    if type(value) is not ElGamalCipher:
      raise TypeError("Cannot set a (Mock)ElGamalCipher to a different type value.")
    self.first[key] = value.first
    self.second[key] = value.second
    return self

  def broadcast_value(self, length):
    (first, is_copy1) = self.first.broadcast_value(length)
    (second, is_copy2) = self.second.broadcast_value(length)
    if not (is_copy1 or is_copy2):
      return (self, False)
    return (ElGamalCipher(first, second), True)

  def __mux__(self, conditional, iffalse):
    if type(iffalse) is not ElGamalCipher:
      raise TypeError("Both sides of mux must have same type.")
    return ElGamalCipher(fl.mux(conditional, self.first, iffalse.first), \
                           fl.mux(conditional, self.second, iffalse.second))

  # There is no need to implement __rmux__, because nothing can be promoted to
  # an ElGamalCipher.

  def decrypt(self, priv_key):
    fc = fl.get_context([self, priv_key])
    # The following checks are not needed. Without them, the same errors will
    # cause the same exceptions to be raised, only a few lines further down.
    if not fc.scope().issubset(self.scope()) \
         or not fc.scope().issubset(priv_key.scope()):
      raise KeyError("Execution scope exceeds data scope.")
    fl.verify(priv_key.len() == 1)
    assert(type(priv_key) is fl.IntArrayIdentifier)
    with fl.massop():
      rc = self.second - self.first * priv_key
    return rc

  def __pos__(self):
    return self  # No real need to return a copy here.

  def __neg__(self):
    return ElGamalCipher(-self.first, -self.second)

  def __add__(self, other):
    if type(other) is not ElGamalCipher:
      return NotImplemented
    fc = self.context()
    length = fc.calc_broadcast_length([self, other])
    (other, is_copy) = other.broadcast_value(length)
    (self2, is_copy) = self.broadcast_value(length)
    return ElGamalCipher(self2.first + other.first, \
                           self2.second + other.second)

  def __iadd__(self, other):
    if type(other) is not ElGamalCipher:
      raise TypeError("Incompatible types for addition.")
    (other, is_copy) = other.broadcast_value(self.len())
    self.first += other.first
    self.second += other.second
    return self

  def __sub__(self, other):
    if type(other) is not ElGamalCipher:
      return NotImplemented
    fc = self.context()
    length = fc.calc_broadcast_length([self, other])
    (other, is_copy) = other.broadcast_value(length)
    (self2, is_copy) = self.broadcast_value(length)
    return ElGamalCipher(self2.first - other.first, \
                           self2.second - other.second)

  def __isub__(self, other):
    if type(other) is not ElGamalCipher:
      raise TypeError("Incompatible types for subtraction.")
    (other, is_copy) = other.broadcast_value(self.len())
    self.first -= other.first
    self.second -= other.second
    return self

  def __mul__(self, other):
    fc = self.context()
    fc.promote(other)
    if not isinstance(other, fl.IntArrayIdentifier):
      return NotImplemented
    length = fc.calc_broadcast_length([self, other])
    (other, is_copy) = other.broadcast_value(length)
    (self2, is_copy) = self.broadcast_value(length)
    return ElGamalCipher(self2.first * other, \
                           self2.second * other)

  def __imul__(self, other):
    fc = self.context()
    fc.promote(other)
    if not isinstance(other, fl.IntArrayIdentifier):
      raise TypeError("Incompatible types for multiplication.")
    (other, is_copy) = other.broadcast_value(self.len())
    self.first *= other
    self.second *= other
    return self

  def __rmul__(self, other):
    return self * other


class ElGamalCipherTransmitter(PairTransmitter):
  def __init__(self, obj):
    super().__init__(obj)
  def transmit(self, map):
    if type(map) is not dict:
      raise TypeError("Expected dict argument in 'transmit'.")
    self.context().verify_context(map.values())
    if not set(map.keys()).issubset(self.context().scope()):
      raise KeyError("Map keys are out of scope in 'transmit'.")
    for k in map:
      if type(map[k]) is not ElGamalCipher:
        raise TypeError("Map values incompatible with transmitter.")
      if not map[k].scope().issubset(self.context().scope()):
        raise KeyError("Transmitted values are outside execution scope.")
    t1 = self.first.transmit({n : item.first for n, item in map.items()})
    t2 = self.second.transmit({n : item.second for n, item in map.items()})
    out_map = {}
    for n in t1.keys():
      with fl.on(t1[n].scope()):
        out_map[n] = ElGamalCipher(t1[n], t2[n])
    return out_map



def elgamal_encrypt(plaintext, pub_key):
  # For simplicity, this example does not feature a zero stockpile.
  #
  assert(type(pub_key) is fl.IntArrayIdentifier)
  fl.verify(pub_key.len() == 1) # This also verifies the scope of pub_key.
  # note that len() of an fc.array is a singleton fc.array, which may have a
  # different value in each node.
  fc = fl.get_context([plaintext, pub_key])
  (plaintext, is_copy) = fc.promote(plaintext)
    # The above line also verifies the scope of plaintext.
  with fl.massop(): # This context defines a single mass operation.
    nonce = fc.randomarray('i', plaintext.len(), 0, 1023)
    rc = ElGamalCipher(nonce, nonce * pub_key + plaintext)
    del nonce
  return rc

def elgamal_refresh(ciphertext, pub_key):
  fc = fl.get_context([ciphertext, pub_key])
  if type(ciphertext) is not ElGamalCipher:
    raise TypeError("Expected (Mock)ElGamalCipher as ciphertext in refresh.")
  zero = fc.array('i', ciphertext.len())
  nonce = elgamal_encrypt(zero, pub_key)
  ciphertext += nonce

def elgamal_sanitise(ciphertext):
  if type(ciphertext) is not ElGamalCipher:
    raise TypeError("Expected (Mock)ElGamalCipher in elgamal_sanitise.")
  mask = ciphertext.context().randomarray('i', ciphertext.len(), 0, 1023)
  ciphertext *= mask

def elgamal_keygen(fc):
  priv_key = fc.randomarray('i', 1, 1, 1023)
  pub_key = priv_key.copy()
  return (priv_key, pub_key)

