# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

# Demo of ftillite usage
# ======================
#
# Revision history:
#
#######################################################


import ftillite as fl
import time

import cProfile
import pstats
import io

import logging
from datetime import datetime
from logging.handlers import RotatingFileHandler
import sys 
import time

# EXAMPLE 1: Creating a pair type.

class Pair(fl.ArrayIdentifier):
  def __init__(self, first, second):
    fc = fl.get_context([first, second])
    super().__init__(fc)
    (first, is_copy1) = fc.promote(first)
    (second, is_copy2) = fc.promote(second)
    length = fc.calc_broadcast_length([first, second])
    (first, is_copy) = first.broadcast_value(length)
    is_copy1 = is_copy1 or is_copy
    (second, is_copy) = second.broadcast_value(length)
    is_copy2 = is_copy2 or is_copy
    if is_copy1:
      self.first = first
    else:
      self.first = first.copy()
    if is_copy2:
      self.second = second
    else:
      self.second = second.copy()

  def transmitter(self):
    return PairTransmitter(self)

  def stub(self):
    return Pair(self.first.stub(), self.second.stub())

  def sametype(self, other):
    return type(other) is Pair and self.first.sametype(other.first) and \
             self.second.sametype(other.second)

  def __eq__(self, other):
    if type(other) is not Pair:
        return 0
    return (self.first == other.first) & (self.second == other.second)

  def __ne__(self, other):
    return 1 - (self == other)

  def flatten(self):
    return self.first.flatten() + self.second.flatten()

  def unflatten(self, data):
    data1 = data[:self.first.width()]
    print(f"Type data: {type(data1)}, len data: {len(data1)}, type data[0]: {type(data1[0])}")
    self.first.unflatten(data[:self.first.width()])
    self.second.unflatten(data[self.first.width():])

  def width(self):
    return self.first.width() + self.second.width()

  def typecode(self):
    return self.first.typecode() + self.second.typecode()

  def copy(self):
    return Pair(self.first, self.second)

  def promote(self, typecode):
    if self.typecode() != typecode:
      raise TypeError("Cannot promote a Pair")
    return (self, False)

  def astype(self, typecode):
    if self.typecode() != typecode:
      raise TypeError("Cannot convert a Pair")
    return self.copy()

  def len(self):
    return self.first.len()

  def set_length(self, new_length):
    self.first.set_length(new_length)
    self.second.set_length(new_length)

  def empty(self):
    self.first.empty()
    self.second.empty()

  def __getitem__(self, key):
    return Pair(self.first[key], self.second[key])

  def lookup(self, key, default = None):
    if default is None:
      return Pair(self.first.lookup(key), self.second.lookup(key))
    elif isinstance(default, Pair):
      return Pair(self.first.lookup(key, default.first), \
                    self.second.lookup(key, default.second))
    else:
      raise TypeError("Unexpected type for default.")

  def __setitem__(self, key, value):
    if not isinstance(value, Pair):
      raise TypeError("Cannot set a Pair to a non-Pair value.")
    self.first[key] = value.first
    self.second[key] = value.second

  def __delitem__(self, key):
    del self.first[key]
    del self.second[key]

  def contains(self, items):
    fc = self.context()
    if not self.sametype(items):
      raise TypeError("Type of items does not match type of container.")
    mylistmap = fc.listmap(self.flatten())
    return mylistmap.lookup(items.flatten()) != -1

  def index(self):
    fc = self.context()
    index1 = fc.listmap([self.first.index()])
    index2 = fc.listmap([self.second.index()])
    pair_index = index1.intersect_items(index2)
    return pair_index.keys()

  # We skip here implementation of "serialise" and "deserialise", for brevity.
  def serialise(self):
    raise NotImplementedError("Serialise not yet implemented.")

  def deserialise(self, data):
    raise NotImplementedError("Deserialise not yet implemented.")

  def broadcast_value(self, length):
    (first, is_copy1) = self.first.broadcast_value(length)
    (second, is_copy2) = self.second.broadcast_value(length)
    if not (is_copy1 or is_copy2):
      return (self, False)
    return (Pair(first, second), True)

  def __mux__(self, conditional, iffalse):
    if type(iffalse) is not Pair:
      raise TypeError("Both sides of mux must have same type.")
    return Pair(fl.mux(conditional, self.first, iffalse.first), \
                fl.mux(conditional, self.second, iffalse.second))

  # There is no need to implement __rmux__, because nothing can be promoted to
  # a Pair.


class PairTransmitter(fl.Transmitter):
  def __init__(self, obj):
    super().__init__(obj.context())
    self.first = obj.first.transmitter()
    self.second = obj.second.transmitter()
  def transmit(self, map):
    if type(map) is not dict:
      raise TypeError("Expected dict argument in 'transmit'.")
    self.context().verify_context(map.values())
    if not set(map.keys()).issubset(self.context().scope()):
      raise KeyError("Map keys are out of scope in 'transmit'.")
    for k in map:
      if type(map[k]) is not Pair:
        raise TypeError("Map values incompatible with transmitter.")
      if not map[k].scope().issubset(self.context().scope()):
        raise KeyError("Transmitted values are outside execution scope.")
    t1 = self.first.transmit({n : item.first for n, item in map.items()})
    t2 = self.second.transmit({n : item.second for n, item in map.items()})
    out_map = {}
    for n in t1.keys():
      with fl.on(t1[n].scope()):
        out_map[n] = Pair(t1[n], t2[n])
    return out_map


# EXAMPLE 2: Creating a dictionary type

class Dict(fl.Identifier):
  def __init__(self, k, v):
    # The keys must be unique, for valid Dict construction.
    fc = fl.get_context([k, v])
    super().__init__(fc)
    (self.k, is_copy) = fc.promote(k)
    if not is_copy:
      self.k = self.k.copy()
    (self.v, is_copy) = fc.promote(v)
    (self.v, is_copy2) = self.v.broadcast_value(self.k.len())
    if not is_copy and not is_copy2:
      self.v = self.v.copy()
    # The next line verifies that the keys are, indeed, unique.
    self.listmap = fc.listmap(self.k.flatten(), order = "pos")

  def transmitter(self):
    return DictTransmitter(self)

  def stub(self):
    return Dict(self.k.stub(), self.v.stub())

  def sametype(self, other):
    return type(other) is Dict and \
             self.k.sametype(other.k) and self.v.sametype(other.v)

  def flatten(self):
    return self.k.flatten() + self.v.flatten()

  def unflatten(self, data):
    self.k.unflatten(data[:self.k.width()])
    self.v.unflatten(data[self.k.width():])
    self.listmap = self.context().listmap(self.k.flatten(), order = "pos")

  def width(self):
    return self.k.width() + self.v.width()

  def typecode(self):
    return self.k.typecode() + self.v.typecode()

  def copy(self):
    return Dict(self.k, self.v)

  def __setitem__(self, k, v):
    if type(k) is slice:
      if k != slice(None, None, None):
        raise KeyError("Only the slice ':' is supported.")
      self.context().verify_context(v)
      if not self.sametype(v):
        raise TypeError("Incompatible type in assignment.")
      self.k[:] = v.k
      self.v[:] = v.v
      self.listmap = fc.listmap(self.k.flatten(), order = "pos")
    else:
      # Items in "k" must be unique.
      (new_keys, new_values) = self.listmap.merge_items(k.flatten())
      # b/c this is not implemented properly, we work around:
      # x = self.listmap.merge_items(k.flatten())
      # new_keys = x[:-1]
      # new_values = x[-1]
      self.k.set_length(self.listmap.len())
      self.v.set_length(self.listmap.len())
      self.v[self.listmap[k.flatten()]] = v # If items in "k" are not unique, this fails.
      self.k[new_values] = new_keys
      # A more complete implementation would have had the code wrapped inside a
      # "try" block in order to roll back the change if anything fails.

  def update(self, other):
    if type(other) is not Dict:
      raise TypeError("Expected a Dict as parameter.")
    (new_keys, new_values) = self.listmap.merge_items(other.keys().flatten())
    # b/c this is not implemented properly, we work around:
    # x = self.listmap.merge_items(other.keys().flatten())
    # new_keys = x[:-1]
    # new_values = x[-1]
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp
    # The following line populates the new keys with a zero value.
    # self.v[new_values] = self.v.stub().set_length(new_values.len())  # Not sure why this line was needed.
    self.v[self.listmap[other.keys().flatten()]] = other.values()

  def __iadd__(self, other):
    if type(other) is not Dict:
      raise TypeError("Expected a Dict as parameter.")
    (new_keys, new_values) = self.listmap.merge_items(other.keys().flatten())
    # b/c this is not implemented properly, we work around:
    # x = self.listmap.merge_items(other.keys().flatten())
    # new_keys = x[:-1]
    # new_values = x[-1]
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp
    # The following line populates the new keys with a zero value.
    # self.v[new_values] = self.v.stub().set_length(new_values.len())  # Not sure why this line was needed.
    self.v[self.listmap[other.keys().flatten()]] += other.values()
    return self

  def __isub__(self, other):
    if type(other) is not Dict:
      raise TypeError("Expected a Dict as parameter.")
    (new_keys, new_values) = self.listmap.merge_items(other.keys().flatten())
    # b/c this is not implemented properly, we work around:
    # x = self.listmap.merge_items(other.keys().flatten())
    # new_keys = x[:-1]
    # new_values = x[-1]
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp
    # The following line populates the new keys with a zero value.
    # self.v[new_values] = self.v.stub().set_length(new_values.len())  # Not sure why this line was needed.
    self.v[self.listmap[other.keys().flatten()]] -= other.values()

  def __imul__(self, other):
    if type(other) is not Dict:
      raise TypeError("Expected a Dict as parameter.")
    (new_keys, new_values) = self.listmap.merge_items(other.keys().flatten())
    # b/c this is not implemented properly, we work around:
    # x = self.listmap.merge_items(other.keys().flatten())
    # new_keys = x[:-1]
    # new_values = x[-1]
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp
    # The following line populates the new keys with a zero value.
    # self.v[new_values] = self.v.stub().set_length(new_values.len())  # Not sure why this line was needed.
    self.v[self.listmap[other.keys().flatten()]] *= other.values()

  def __itruediv__(self, other):
    if type(other) is not Dict:
      raise TypeError("Expected a Dict as parameter.")
    (new_keys, new_values) = self.listmap.merge_items(other.keys().flatten())
    # b/c this is not implemented properly, we work around:
    # x = self.listmap.merge_items(other.keys().flatten())
    # new_keys = x[:-1]
    # new_values = x[-1]
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp
    # The following line populates the new keys with a zero value.
    # self.v[new_values] = self.v.stub().set_length(new_values.len())  # Not sure why this line was needed.
    self.v[self.listmap[other.keys().flatten()]] /= other.values()

  def __ifloordiv__(self, other):
    if type(other) is not Dict:
      raise TypeError("Expected a Dict as parameter.")
    (new_keys, new_values) = self.listmap.merge_items(other.keys().flatten())
    # b/c this is not implemented properly, we work around:
    # x = self.listmap.merge_items(other.keys().flatten())
    # new_keys = x[:-1]
    # new_values = x[-1]
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp
    # The following line populates the new keys with a zero value.
    # self.v[new_values] = self.v.stub().set_length(new_values.len())  # Not sure why this line was needed.
    self.v[self.listmap[other.keys().flatten()]] //= other.values()

  def len(self):
    return self.keys().len()

  def keys(self):
    return self.k

  def values(self):
    return self.v

  def __getitem__(self, k):
    return self.v[self.listmap[k.flatten()]]

  def lookup(self, k, default = None):
    # default must be of the same type as self.v or None
    if default is None:
      default = self.v.stub().set_length(1)
    indices = self.listmap.lookup(k.flatten())
    return self.v.lookup(indices, default)
      # This previous line fails if default is not of the same type as self.v

  def reduce_sum(self, key, value):
    (new_keys, new_values) = self.listmap.merge_items(key.flatten())
    # b/c this is not implemented properly, we work around:
    # x = self.listmap.merge_items(key.flatten())
    # new_keys = x[:-1]
    # new_values = x[-1]
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp
    return self.v.reduce_sum(self.listmap[key.flatten()], value)

  def reduce_isum(self, key, value):
    (new_keys, new_values) = self.listmap.merge_items(key.flatten())
    # b/c this is not implemented properly, we work around:
    # x = self.listmap.merge_items(key.flatten())
    # new_keys = x[:-1]
    # new_values = x[-1]
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp
    self.v.reduce_isum(self.listmap[key.flatten()], value)

  def contains(self, k):
    return self.listmap.contains(k.flatten())

  def __delitems__(self, k):
    (moved_keys, old_values, new_values) = self.listmap.remove_items(k.flatten())
    self.v[new_values] = self.v[old_values]
    self.k[new_values] = self.k[old_values]
    self.v.set_length(self.listmap.len())
    self.k.set_length(self.listmap.len())

  def discard_items(self, k):
    (moved_keys, old_values, new_values) = self.listmap.discard_items(k.flatten())
    self.v[new_values] = self.v[old_values]
    self.k[new_values] = self.k[old_values]
    self.v.set_length(self.listmap.len())
    self.k.set_length(self.listmap.len())


class DictTransmitter(fl.Transmitter):
  def __init__(self, obj):
    super().__init__(obj.context())
    self.k = obj.k.transmitter()
    self.v = obj.v.transmitter()
  def transmit(self, map):
    if type(map) is not dict:
      raise TypeError("Expected dict argument in 'transmit'.")
    self.context().verify_context(map.values())
    if not set(map.keys()).issubset(self.context().scope()):
      raise KeyError("Map keys are out of scope in 'transmit'.")
    for k in map:
      if type(map[k]) is not Dict:
        raise TypeError("Map values incompatible with transmitter.")
      if not map[k].scope().issubset(self.context().scope()):
        raise KeyError("Transmitted values are outside execution scope.")
    # The following two lines fail if items are not of the same type as obj.
    tk = self.k.transmit({n : item.k for n, item in map.items()})
    tv = self.v.transmit({n : item.v for n, item in map.items()})
    out_map = {}
    for n in tk.keys():
      with fl.on(tk[n].scope()):
        out_map[n] = Dict(tk[n], tv[n])
    return out_map


# EXAMPLE 3: Sparse matrix by vector multiplication.

def matvecmult(m,v):
  rc = v.stub()
  with fl.massop():
    rc.reduce_sum(m.keys().second, m.values() * v[m.keys().first])
  return rc


# Example 4: Implementing the ElGamal cryptosystem.

class ElGamalCipher(Pair):
  def __init__(self, mask, masked_message):
    if type(mask) is not Ed25519ArrayIdentifier or \
        type(masked_message) is not Ed25519ArrayIdentifier:
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
      raise TypeError("Cannot promote an ElGamalCipher")
    return (self, False)

  def astype(self, typecode):
    if self.typecode() != typecode:
      raise TypeError("Cannot convert an ElGamalCipher")
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
      raise TypeError("Cannot set an ElGamalCipher to a different type value.")
    self.first[key] = value.first
    self.second[key] = value.second

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
    # This check is not needed, either.
    assert(type(priv_key) is fl.Ed25519IntArrayIdentifier)
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

  def __mul__(self, other):
    fc = self.context()
    fc.promote(other)
    if not isinstance(other, \
                        (fl.IntArrayIdentifier, fl.Ed25519IntArrayIdentifier)):
      return NotImplemented
    length = fc.calc_broadcast_length([self, other])
    (other, is_copy) = other.broadcast_value(length)
    (self2, is_copy) = self.broadcast_value(length)
    return ElGamalCipher(self2.first * other, \
                           self2.second * other)

  def __imul__(self, other):
    fc = self.context()
    fc.promote(other)
    if not isinstance(other, \
                        (fl.IntArrayIdentifier, fl.Ed25519IntArrayIdentifier)):
      raise TypeError("Incompatible types for multiplication.")
    (other, is_copy) = other.broadcast_value(self.len())
    self.first *= other
    self.second *= other

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
  assert(type(pub_key) is fl.Ed25519ArrayIdentifier)
  fl.verify(pub_key.len() == 1) # This also verifies the scope of pub_key.
  # note that len() of an fc.array is a singleton fc.array, which may have a
  # different value in each node.
  fc = fl.get_context([plaintext, pub_key])
  (plaintext, is_copy) = fc.promote(plaintext)
    # The above line also verifies the scope of plaintext.
  with fl.massop(): # This context defines a single mass operation.
    nonce = fc.randomarray('I', plaintext.len(), False)
    rc = ElGamalCipher(nonce.astype('E'), \
                         nonce * pub_key + plaintext.astype('E'))
    del nonce
  return rc

def elgamal_refresh(ciphertext, pub_key):
  fc = fl.get_context([ciphertext, pub_key])
  if type(ciphertext) is not ElGamalCipher:
    raise TypeError("Expected ElGamalCipher as ciphertext in refresh.")
  zero = fc.array('i', ciphertext.len())
  nonce = elgamal_encrypt(zero, pub_key)
  ciphertext += nonce

def elgamal_sanitise(ciphertext):
  if type(ciphertext) is not ElGamalCipher:
    raise TypeError("Expected ElGamalCipher in elgamal_sanitise.")
  mask = ciphertext.context().randomarray('I', ciphertext.len(), True)
  ciphertext *= mask

def elgamal_keygen(fc):
  priv_key = fc.randomarray('I', 1)
  pub_key = priv_key.astype('E')
  return (priv_key, pub_key)




class MockElGamalCipher(Pair):
  def __init__(self, mask, masked_message):
    super().__init__(mask, masked_message)

  def transmitter(self):
    return MockElGamalCipherTransmitter(self)

  def stub(self):
    return MockElGamalCipher(self.first.stub(), self.second.stub())

  def sametype(self, other):
    return type(other) is MockElGamalCipher and \
             self.first.sametype(other.first) and \
             self.second.sametype(other.second)

  def __eq__(self, other):
    if type(other) is not MockElGamalCipher:
        return 0
    return (self.first == other.first) & (self.second == other.second)

  def __ne__(self, other):
    return 1 - (self == other)

  def copy(self):
    return MockElGamalCipher(self.first, self.second)

  def promote(self, typecode):
    if self.typecode() != typecode:
      raise TypeError("Cannot promote a MockElGamalCipher")
    return (self, False)

  def astype(self, typecode):
    if self.typecode() != typecode:
      raise TypeError("Cannot convert a MockElGamalCipher")
    return self.copy()

  def __getitem__(self, key):
    return MockElGamalCipher(self.first[key], self.second[key])

  def lookup(self, key, default = None):
    if default is None:
      return MockElGamalCipher(self.first.lookup(key), self.second.lookup(key))
    elif self.sametype(default):
      return MockElGamalCipher(self.first.lookup(key, default.first), \
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
    if type(value) is not MockElGamalCipher:
      raise TypeError("Cannot set a MockElGamalCipher to a different type value.")
    self.first[key] = value.first
    self.second[key] = value.second
    return self

  def broadcast_value(self, length):
    (first, is_copy1) = self.first.broadcast_value(length)
    (second, is_copy2) = self.second.broadcast_value(length)
    if not (is_copy1 or is_copy2):
      return (self, False)
    return (MockElGamalCipher(first, second), True)

  def __mux__(self, conditional, iffalse):
    if type(iffalse) is not MockElGamalCipher:
      raise TypeError("Both sides of mux must have same type.")
    return MockElGamalCipher(fl.mux(conditional, self.first, iffalse.first), \
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
    # fl.verify(priv_key.len() == 1)
    rc = self.second - self.first * priv_key
    return rc

  def __pos__(self):
    return self  # No real need to return a copy here.

  def __neg__(self):
    return MockElGamalCipher(-self.first, -self.second)

  def __add__(self, other):
    if type(other) is not MockElGamalCipher:
      return NotImplemented
    fc = self.context()
    length = fc.calc_broadcast_length([self, other])
    (other, is_copy) = other.broadcast_value(length)
    (self2, is_copy) = self.broadcast_value(length)
    return MockElGamalCipher(self2.first + other.first, \
                           self2.second + other.second)

  def __iadd__(self, other):
    if type(other) is not MockElGamalCipher:
      raise TypeError("Incompatible types for addition.")
    (other, is_copy) = other.broadcast_value(self.len())
    self.first += other.first
    self.second += other.second
    return self

  def __sub__(self, other):
    if type(other) is not MockElGamalCipher:
      return NotImplemented
    fc = self.context()
    length = fc.calc_broadcast_length([self, other])
    (other, is_copy) = other.broadcast_value(length)
    (self2, is_copy) = self.broadcast_value(length)
    return MockElGamalCipher(self2.first - other.first, \
                           self2.second - other.second)

  def __isub__(self, other):
    if type(other) is not MockElGamalCipher:
      raise TypeError("Incompatible types for subtraction.")
    (other, is_copy) = other.broadcast_value(self.len())
    self.first -= other.first
    self.second -= other.second
    return self

  def __mul__(self, other):
    fc = self.context()
    fc.promote(other)
    length = fc.calc_broadcast_length([self, other])
    (other, is_copy) = other.broadcast_value(length)
    (self2, is_copy) = self.broadcast_value(length)
    return MockElGamalCipher(self2.first * other, \
                           self2.second * other)

  def __imul__(self, other):
    fc = self.context()
    fc.promote(other)
    (other, is_copy) = other.broadcast_value(self.len())
    self.first *= other
    self.second *= other
    return self

  def __rmul__(self, other):
    return self * other


class MockElGamalCipherTransmitter(PairTransmitter):
  def __init__(self, obj):
    super().__init__(obj)
  def transmit(self, map):
    if type(map) is not dict:
      raise TypeError("Expected dict argument in 'transmit'.")
    self.context().verify_context(map.values())
    if not set(map.keys()).issubset(self.context().scope()):
      raise KeyError("Map keys are out of scope in 'transmit'.")
    for k in map:
      if type(map[k]) is not MockElGamalCipher:
        raise TypeError("Map values incompatible with transmitter.")
      if not map[k].scope().issubset(self.context().scope()):
        raise KeyError("Transmitted values are outside execution scope.")
    t1 = self.first.transmit({n : item.first for n, item in map.items()})
    t2 = self.second.transmit({n : item.second for n, item in map.items()})
    out_map = {}
    for n in t1.keys():
      with fl.on(t1[n].scope()):
        out_map[n] = MockElGamalCipher(t1[n], t2[n])
    return out_map



def mock_elgamal_encrypt(plaintext, pub_key):
  # For simplicity, this example does not feature a zero stockpile.
  #
  # fl.verify(pub_key.len() == 1) # This also verifies the scope of pub_key.
  # note that len() of an fc.array is a singleton fc.array, which may have a
  # different value in each node.
  fc = fl.get_context([plaintext, pub_key])
  (plaintext, is_copy) = fc.promote(plaintext)
    # The above line also verifies the scope of plaintext.
  with fl.massop(): # This context defines a single mass operation.
    nonce = fc.randomarray('i', plaintext.len(), 0, 1023)
    rc = MockElGamalCipher(nonce, nonce * pub_key + plaintext)
    del nonce
  return rc

def mock_elgamal_refresh(ciphertext, pub_key):
  fc = fl.get_context([ciphertext, pub_key])
  if type(ciphertext) is not MockElGamalCipher:
    raise TypeError("Expected MockElGamalCipher as ciphertext in refresh.")
  zero = fc.array('i', ciphertext.len())
  nonce = mock_elgamal_encrypt(zero, pub_key)
  ciphertext += nonce

def mock_elgamal_sanitise(ciphertext):
  if type(ciphertext) is not MockElGamalCipher:
    raise TypeError("Expected ElGamalCipher in elgamal_sanitise.")
  mask = ciphertext.context().randomarray('i', ciphertext.len(), 0, 1023)
  ciphertext *= mask

def mock_elgamal_keygen(fc):
  priv_key = fc.randomarray('i', 1, 1, 1023)
  pub_key = priv_key.copy()
  return (priv_key, pub_key)


# We want to set up the mapping from 6-digit BSB numbers to nodes.
# If this is known, it can be set up like this:
# branch2node = fc.array('i', [-1, 1, -1, 4, 4, -1, 2, -1, 3]).lookup( \
#                   fc.arange(1000000) // 10000, -1)

# A more robust approach is to set it up from the actual data on the nodes,
# as follows.

def get_branch2node(fc):
  branch2node_first = fc.array('i', 1000000, -1)
  branch2node_last = fc.array('i', 1000000, -1)
  all_accounts = Pair(fc.array('i'), fc.array('i'))
  peer_nodes = fc.scope().difference(fc.CoordinatorID)
  with fl.on(peer_nodes):
    bsbs = fc.array('i')
    bsbs.auxdb_read("SELECT bsb FROM accounts;")
  bsbs = fl.transmit({i : bsbs for i in fc.scope()})
  k = bsbs.keys()
  for i in k:
    branch2node_last[bsbs[i]] = i.num()
  for i in reversed(k):
    branch2node_first[bsbs[i]] = i.num()
  # fl.verify(branch2node_first == branch2node_last)
  # "verify" in above command is not implemented yet. So, instead, we do this:
  if list((branch2node_first != branch2node_last).index().len()) != [0]:
    raise ValueError("Unexpected error: The same BSB is used by multiple REs.")
  return branch2node_first

def distribute_tag(tag, pub_key):
  out_scope = tag.context().scope().difference([tag.context().CoordinatorID])
  map = {}
  with fl.on(tag.scope()):
    for n in out_scope:
      # This can be done more efficiently using "discard_items".
      ind = (branch2node[tag.keys().first] == n.num()).index()
      map[n] = Dict(tag.keys()[ind], tag.values()[ind])
  # The next loop performs refreshing, prior to transmitting.
  for n in out_scope:
    with fl.on(tag.scope().difference(set([n]))):
      mock_elgamal_refresh(map[n].values(), pub_key)
  return fl.transmit(map)


# dist_tag = distribute_tag(tag, pub_key)[fc.CoordinatorID]

def fintracer_step(tag, txs, pub_key):
  fc = fl.get_context([tag, txs, pub_key])
  with fl.on(tag.scope()):
    zero = fc.array('i', 1)
    nonce = mock_elgamal_encrypt(zero, pub_key)
    next_tag = tag.stub()
    with fl.massop():
      next_tag.reduce_sum(txs.second, tag.lookup(txs.first, nonce))
  map = distribute_tag(next_tag, pub_key)
  scope = set()
  for n in map:
    scope.update(map[n].scope())
  with fl.on(scope):
    # We need to create here a new empty tag, but can't use ".stub()" because
    # it may have a different scope than any existing tag.
    # In a real implementation, we would have had a "tag" type, and all this
    # would have happened automatically.
    accounts_stub = Pair(fc.array('i'), fc.array('i'))
    cipher_stub = MockElGamalCipher(fc.array('i'), fc.array('i'))
    rc = Dict(accounts_stub, cipher_stub)
    for n in map:
      with fl.on(map[n].scope()):
        rc += map[n]
  return rc

if __name__ == "__main__":
  
  ### Start cProfile
  pr = cProfile.Profile()
  pr.enable()
  ###
  
  # EXAMPLE 5: Initialising a connection.
  conf = fl.FTILConf().set_rabbitmq_conf({'user': 'ftillite', 'password': 'ftillite', 'host': 'localhost'})
  fc = fl.FTILContext(conf = conf)

  # Example 6: A FinTracer propagation step.
  start = time.time()

  with fl.on(fc.CoordinatorID): # On the coordinator node...

    # Creating an ElGamal key pair.
    (priv_key, local_pub_key) = mock_elgamal_keygen(fc)

  # Back on all nodes: distributing the public key across all nodes.
  pub_key = fl.transmit({i : local_pub_key for i in fc.scope()})[fc.CoordinatorID]

  # On the peer nodes...
  peer_nodes = fc.scope().difference(fc.CoordinatorID)

  accounts = Pair(fc.array('i'), fc.array('i'))

  transactions = Pair(accounts, accounts)

  with fl.on(peer_nodes):
    transactions.auxdb_read("SELECT DISTINCT origin_bsb, origin_id, dest_bsb, dest_id FROM transactions;")

  with fl.on(peer_nodes):
    accounts.auxdb_read("SELECT DISTINCT bsb, account_id FROM accounts LIMIT 1;")
    values = fc.array('i', accounts.len(), 1)
    ciphertext = mock_elgamal_encrypt(values, pub_key)
    dist_tag = Dict(accounts, ciphertext)
  
  branch2node = get_branch2node(fc)
  
  dist_tag = fintracer_step(dist_tag, transactions, pub_key)

  print("Done.")

  end = time.time()
  print(f"Time for the full computation: {end - start}")
  
  ### Stop and output profiler results to disk
  pr.disable()
  result = io.StringIO()
  pstats.Stats(pr,stream=result).print_stats()
  result=result.getvalue()
  result='ncalls'+result.split('ncalls')[-1]
  result='\n'.join([','.join(line.rstrip().split(None,5)) for line in result.split('\n')])
  with open('profiler_stats/check_listmap_profiler_stats.csv', 'w+') as f:
      f.write(result)
      f.close()
  ###
  
  