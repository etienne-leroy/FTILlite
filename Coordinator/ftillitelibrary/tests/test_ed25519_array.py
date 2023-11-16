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

def test_new_array_with_value(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    zeroes = fc.array("E", 10)
    v1 = fc.array("i", 1, 5)
    v2 = v1.astype("E")
    v3 = fc.array("E", 10, v2)
    v4 = v3 != zeroes
    assert fl.verify(v4, False)

def test_new_array_no_value(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    zeroes = fc.array("E", 10)
    v1 = fc.array("E", 10)
    v2 = zeroes == v1
    assert fl.verify(v2, False)

def test_contains(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", 1, 5)
    v2 = v1.astype("E")
    v3 = fc.array("E", 10, v2)
    v4 = v3.contains(v2)
    assert fl.verify(v4, False)

def test_cumsum(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4, 5, 6, 7])
    v2 = v1.astype("E")
    v3 = v2.cumsum()
    v4 = fc.array("i", [1, 3, 6, 10, 15, 21, 28])
    v5 = v4.astype("E")
    v6 = v3 == v5
    assert fl.verify(v6, False)

def test_mux(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [1, 0, 1, 2, 0, 0, 3])
    v2 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v2b = v2.astype("E")
    v3 = fc.array("i", [11, 22, 33, 44, 55, 66, 77])
    v3b = v3.astype("E")
    v4 = fl.mux(v1,v2b,v3b)
    v5 = fc.array("i", [10, 22, 30, 40, 55, 66, 70])
    v5b = v5.astype("E")
    v6 = v4 == v5b
    assert fl.verify(v6, False)

    # Test Case 2: Relates to Issue 57
    cond = fc.array('i', [1,0,1,0,1])
    ifttrue = fc.array('i', [8]).astype("E")
    iffalse = fc.array('i', [4]).astype("E")
    v7 = fl.mux(cond, ifttrue, iffalse)
    v8 = fc.array('i', [8,4,8,4,8]).astype('E')
    v9 = v7 == v8
    assert fl.verify(v9, False)

def test_eq(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")
    
    # Test Case 1
    v1 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v1b = v1.astype("E")
    v2 = fc.array("i", [10, 25, 30, 45, 50, 60, 70])
    v2b = v2.astype("E")
    v3 = v1b == v2b
    v4 = fc.array("i", [1, 0, 1, 0, 1, 1, 1])
    v5 = v3 == v4
    assert fl.verify(v5) 

def test_neq(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v1b = v1.astype("E")
    v2 = fc.array("i", [10, 25, 30, 45, 50, 60, 70])
    v2b = v2.astype("E")
    v3 = v1b != v2b
    v4 = fc.array("i", [0, 1, 0, 1, 0, 0, 0])
    v5 = v3 == v4
    assert fl.verify(v5, False)

    # Test Case 2
    v1 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v1b = v1.astype("E")
    v1c = -v1b
    v2 = v1c != v1b
    assert fl.verify(v2, False)

def test_add(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v1b = v1.astype("E")
    v2 = fc.array("i", [10, 25, 30, 45, 50, 60, 70])
    v2b = v2.astype("E")
    v3 = fc.array("i", [20, 45, 60, 85, 100, 120, 140])
    v3b = v3.astype("E")
    v4 = v2b + v1b
    v3 = v3b == v4
    assert fl.verify(v3, False)

def test_sub(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [20, 40, 50, 60, 70, 80, 100])
    v1b = v1.astype("E")
    v2 = fc.array("i", [10, 25, 30, 45, 50, 60, 70])
    v2b = v2.astype("E")
    v3 = fc.array("i", [10, 15, 20, 15, 20, 20, 30])
    v3b = v3.astype("E")
    v4 = v1b - v2b
    v3 = v3b == v4
    assert fl.verify(v3, False)

def test_mul(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [20, 40, 50, 60, 70, 80, 100])
    v1b = v1.astype("E")
    v2 = fc.array("i", [1, 2, 3, 4, 5, 6, 7])
    v2b = v2.astype("I")
    v3 = fc.array("i", [20, 80, 150, 240, 350, 480, 700])
    v3b = v3.astype("E")
    v4 = v1b * v2b
    v5 = v3b == v4
    assert fl.verify(v5, False)

    # Test Case 2: see issue #140
    # check E * i
    v6 = fc.array("i", [20, 40, 50, 60, 70, 80, 100]).astype("E")
    v7 = v6 * 5
    v8 = fc.array("i", [100, 200, 250, 300, 350, 400, 500]).astype("E")
    assert fl.verify(v7 == v8)
    # check i * E
    v9 = 5 * v6
    assert fl.verify(v9 == v8)


def test_setLength(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [20, 40, 50, 60, 70, 80, 100])
    v1b = v1.astype("E")
    v1b.set_length(3)
    v2 = fc.array("i", [20, 40, 50])
    v2b = v2.astype("E")
    v3 = v1b == v2b
    assert fl.verify(v3, False)

    # Test Case 2: See Issue #131
    v1 = fc.array("E", 0)
    v1.set_length(0)
    assert True

def test_setItem_slice(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    values = fc.array("i", [44, 55, 66]).astype("E")

    arr = fc.array("i", [20, 40, 55, 60, 70, 80, 100]).astype("E")

    arr[1:4] = values

    expected = fc.array("i", [20, 44, 55, 66, 70, 80, 100]).astype("E")
    assert fl.verify(arr == expected, False)

def test_setItem_replace(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    values = fc.array("i", [44, 55, 66]).astype("E")

    arr = fc.array("i", [20, 40, 55, 60, 70, 80, 100]).astype("E")

    arr[:] = values

    expected = fc.array("i", [44, 55, 66]).astype("E")
    assert fl.verify(arr == expected, False)

def test_index(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [ 1, 2, 0, 4, 0, 6, 7])
    v1b = v1.astype("E")
    v2 = v1b.index()
    v3 = fc.array("i", [0, 1, 3, 5, 6])
    v4 = v2 == v3
    assert fl.verify(v4, False)

def test_reducesum(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [ 1, 2, 3, 4, 5, 6, 7 ])
    v1b = v1.astype("E")
    v2 = fc.array("i", [5, 11, 2])
    v2b = v2.astype("E")
    v3 = fc.array("i", [2, 2, 3])
    v1b.reduce_sum(v3, v2b)
    v4 = fc.array("i", [1, 2, 16, 2, 5, 6, 7])
    v4b = v4.astype("E")
    v5 = v1b == v4b
    assert fl.verify(v5, False)

def test_reduceisum(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [ 1, 2, 3, 4, 5, 6, 7 ])
    v1b = v1.astype("E")
    v2 = fc.array("i", [5, 11, 2])
    v2b = v2.astype("E")
    v3 = fc.array("i", [2, 2, 3])
    v1b.reduce_isum(v3, v2b)
    v4 = fc.array("i", [1, 2, 19, 6, 5, 6, 7])
    v4b = v4.astype("E")
    v5 = v1b == v4b
    assert fl.verify(v5, False)

# def test_ed_folded_bytearray_ed_project(fc):
#     v1 = fc.array("i", [ 1, 2, 3, 4, 5, 6, 7 ])
#     v1b = v1.astype("E")

#     v2 = v1b.ed_folded()
#     v3 = v2.ed_project()

#     v4 = v1b == v3
#     assert fl.verify(v4, False)

def test_ed_affine_bytearray_ed_affine_project(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [ 1, 2, 3, 4, 5, 6, 7 ])
    v1b = v1.astype("E")
    v2 = v1b.ed_affine()
    v3 = v2.ed_affine_project()
    v4 = v1b == v3
    assert fl.verify(v4, False)

def test_delitem(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4])
    v1b = v1.astype("E")
    del v1b[1]
    v2 = fc.array("i", [1, 3, 4])
    v2b = v2.astype("E")
    v3 = v1b == v2b
    assert fl.verify(v3, False)

def test_len(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1
    v1 = fc.array('i', 100, 5)
    v2 = v1.astype("E").len()
    v3 = fc.array('i', 1, 100)
    v4 = v2 == v3
    assert fl.verify(v4, False)

    # Test Case 2
    v5 = fc.array('i', [1, 2, 3, 4, 5])
    v6 = v5.astype("E").len()
    v7 = fc.array('i', 1, 5)
    v8 = v6 == v7
    assert fl.verify(v8, False)


def test_lookup(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")
        
    arr = fc.array('i', [1, 2, 3, 4, 5]).astype("E")
    actual = arr.lookup([1,3])
    expected = fc.array('i', [2,4]).astype("E")

    assert fl.verify(actual == expected, False)
    
def test_serialise(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")
    in_1 = fc.array('i', [10, 1, 30, 2]).astype('E')
    in_2 = in_1.serialise()
    assert in_2.typecode() == 'b64'
    
def test_deserialise(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")
    in_1 = fc.array('i', [10, 1, 30, 2]).astype('E')
    in_2 = in_1.serialise()
    in_3 = fc.array('E')
    in_3.deserialise(in_2)
    assert fl.verify(in_1 == in_3)