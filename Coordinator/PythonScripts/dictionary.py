# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

import ftillite as fl

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
      self.k.set_length(self.listmap.len())
      self.v.set_length(self.listmap.len())
      self.v[self.listmap[k.flatten()]] = v # If items in "k" are not unique,
                                            # this fails.
      self.k[new_values] = new_keys
      # A more complete implementation would have had the code wrapped inside a
      # "try" block in order to roll back the change if anything fails.

  def _raw_update_key(self, key):
    (new_keys, new_values) = self.listmap.merge_items(key.flatten())
    self.k.set_length(self.listmap.len())
    self.v.set_length(self.listmap.len())
    # A solution not based on "unflatten" would have been more generic.
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    # b/c unflatten is not implemented to return a value:
    # self.k[new_values] = self.k.stub().unflatten(new_keys)
    temp = self.k.stub()
    temp.unflatten(new_keys)
    self.k[new_values] = temp

  def _raw_update(self, other):
    if type(other) is not Dict:
      raise TypeError("Expected a Dict as parameter.")
    self._raw_update_key(other.keys())

  def update(self, other):
    self._raw_update(other)
    self.v[self.listmap[other.keys().flatten()]] = other.values()

  def __iadd__(self, other):
    self._raw_update(other)
    self.v[self.listmap[other.keys().flatten()]] += other.values()
    return self

  def __isub__(self, other):
    self._raw_update(other)
    self.v[self.listmap[other.keys().flatten()]] -= other.values()

  def __imul__(self, other):
    # This can be optimised. As is, it creates new keys with zero values,
    # then multiplies the zeroes by the values of "other".
    self._raw_update(other)
    self.v[self.listmap[other.keys().flatten()]] *= other.values()

  def __itruediv__(self, other):
    # As in __imul__, this can be optimised.
    self._raw_update(other)
    self.v[self.listmap[other.keys().flatten()]] /= other.values()

  def __ifloordiv__(self, other):
    # As in __imul__, this can be optimised.
    self._raw_update(other)
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
    self._raw_update_key(key)
    return self.v.reduce_sum(self.listmap[key.flatten()], value)

  def reduce_isum(self, key, value):
    self._raw_update_key(key)
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
