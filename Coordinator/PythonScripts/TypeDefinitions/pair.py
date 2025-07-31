import ftillite as fl

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
  def __len__(self):
    return len(self.first)
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
  def __iter__(self):
    return zip(list(self.first), list(self.second))

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

