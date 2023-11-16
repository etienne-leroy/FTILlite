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
import pytest

@pytest.mark.skip(reason="Exception raised in different thread isn't detected.")
def test_transmit_error(fc):
    v1 = fc.array("i", 1, 5)
    v2 = v1.astype("I")

    with pytest.raises(RuntimeError):
        x = fl.transmit({i : v2 for i in fc.scope()})[fc.CoordinatorID]
    assert "something"

def test_transmit_intarray(fc):
    int1 = fc.array("i", 1, 5)
    int2 = fl.transmit({i : int1 for i in fc.scope()})[fc.CoordinatorID]
    int3 = int1 == int2

    assert fl.verify(int3, False)

def test_transmit_emptyintarray(fc):
    int1 = fc.array("i", 1)
    int2 = fl.transmit({i : int1 for i in fc.scope()})[fc.CoordinatorID]
    int3 = int1 == int2

    assert fl.verify(int3, False)

def test_transmit_emptyintarraynolength(fc):
    int1 = fc.array("i", 0)
    int2 = fl.transmit({i : int1 for i in fc.scope()})[fc.CoordinatorID]
    int3 = int1 == int2

    assert fl.verify(int3, False)

def test_transmit_emptyfloatarraynolength(fc):
    f1 = fc.array("f", 0)
    f2 = fl.transmit({i : f1 for i in fc.scope()})[fc.CoordinatorID]
    f3 = f1 == f2

    assert fl.verify(f3, False)

def test_transmit_emptybytearrayarraynolength(fc):
    b1 = fc.array("i", 0)
    b2 = b1.astype("b8")
    b3 = fl.transmit({i : b2 for i in fc.scope()})[fc.CoordinatorID]
    b4 = b2 == b3

    assert fl.verify(b4, False)

def test_transmit_emptyed25519arraynolength(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")
    int1 = fc.array("E", 0)
    int2 = fl.transmit({i : int1 for i in fc.scope()})[fc.CoordinatorID]
    int3 = int1 == int2

    assert fl.verify(int3, False)

def test_transmit_floatarray(fc):
    f1 = fc.array("f", [1.1, 2.1, 3.3, 4.4, 5.5, 6.6, 7.7])
    f2 = fl.transmit({i : f1 for i in fc.scope()})[fc.CoordinatorID]
    i1 = f1 == f2

    assert fl.verify(i1, False)

def test_transmit_bytearrayarray(fc):
    i1 = fc.array("i", [1, 2, 3, 4, 5])
    b2 = i1.astype("b8")
    b3 = fl.transmit({i : b2 for i in fc.scope()})[fc.CoordinatorID]
    i2 = b2 == b3

    assert fl.verify(i2, False)

def test_transmit_ed25519Integer(fc):
    i1 = fc.array("i", 1, 5)
    e1 = i1.astype("I")
    e2 = fl.transmit({i : e1 for i in fc.scope()})[fc.CoordinatorID]
    i2 = e1 == e2

    assert fl.verify(i2, False)
    