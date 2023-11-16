# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from .util import fc, check_array
import ftillite as fl
import pytest

def test_new_array_with_value(fc):
    # Test Case 1
    v1 = fc.array("i", 1, 5)
    v2 = v1.astype("I")
    v3 = fc.array("I", 10, v2)
    assert check_array(v3, 'I', [5, 5, 5, 5, 5, 5, 5, 5, 5, 5])

def test_new_array_no_value(fc):
    # Test Case 1
    v1 = fc.array("I", 10)
    assert check_array(v1, 'I', [0, 0, 0, 0, 0, 0, 0, 0, 0, 0])
    
def test_randomarray(fc):
    # Test Case 1
    v1 = fc.randomarray("I", 10)
    v2 = fc.array("I", 10)
    assert not check_array(v1, 'I', [0, 0, 0, 0, 0, 0, 0, 0, 0, 0])
    
def test_contains(fc):
    # Test Case 1
    v1 = fc.array("i", 1, 5)
    v2 = v1.astype("I")
    v3 = fc.array("I", 10, v2)
    v4 = v3.contains(v2)
    assert fl.verify(v4, False)
 
def test_cumsum(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4, 5, 6, 7])
    v2 = v1.astype("I")
    v3 = v2.cumsum()
    assert check_array(v3, 'I', [1, 3, 6, 10, 15, 21, 28])
    
def test_mux(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 0, 1, 2, 0, 0, 3])
    v2 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v2b = v2.astype("I")
    v3 = fc.array("i", [11, 22, 33, 44, 55, 66, 77])
    v3b = v3.astype("I")
    v4 = fl.mux(v1,v2b,v3b)
    assert check_array(v4, 'I', [10, 22, 30, 40, 55, 66, 70])

    # Test Case 3: Relates to Issue 57
    cond = fc.array('i', [1, 0, 1, 0, 1])
    ifttrue = fc.array('i', [8]).astype("I")
    iffalse = fc.array('i', [4]).astype("I")
    v5 = fl.mux(cond, ifttrue, iffalse)
    assert check_array(v5, 'I', [8, 4, 8, 4, 8])
    
def test_eq(fc):
    # Test Case 1
    v1 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v1b = v1.astype("I")
    v2 = fc.array("i", [10, 25, 30, 45, 50, 60, 70])
    v2b = v2.astype("I")
    v3 = v1b == v2b
    assert not fl.verify(v3, False)

def test_neq(fc):
    # Test Case 1
    v1 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v1b = v1.astype("I")
    v2 = fc.array("i", [10, 25, 30, 45, 50, 60, 70])
    v2b = v2.astype("I")
    v3 = v1b != v2b
    assert not fl.verify(v3, False)
    v2c = -v2b
    v3 = v1b != v2c
    assert fl.verify(v3, False)
    
def test_add(fc):
    # Test Case 1
    v1 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v1b = v1.astype("I")
    v2 = fc.array("i", [10, 25, 30, 45, 50, 60, 70])
    v2b = v2.astype("I")
    v3 = v2b + v1b
    assert check_array(v3, 'I', [20, 45, 60, 85, 100, 120, 140])
    
def test_sub(fc):
    # Test Case 1
    v1 = fc.array("i", [20, 40, 50, 60, 70, 80, 100])
    v1b = v1.astype("I")
    v2 = fc.array("i", [10, 25, 30, 45, 50, 60, 70])
    v2b = v2.astype("I")
    v3 = v1b - v2b
    assert check_array(v3, 'I', [10, 15, 20, 15, 20, 20, 30])
    
def test_mul(fc):
    # Test Case 1
    v1 = fc.array("i", [20, 40, 50, 60, 70, 80, 100])
    v1b = v1.astype("I")
    v2 = fc.array("i", [1, 2, 3, 4, 5, 6, 7])
    v2b = v2.astype("I")
    v3 = v1b * v2b
    assert check_array(v3, 'I', [20, 80, 150, 240, 350, 480, 700])

def test_mul_empty_array(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    v0 = fc.array("i")
    v0E = v0.astype("E")
    v0I = v0.astype("I")

    v1 = fc.array("i", [20])
    v1E = v1.astype("E")
    v1I = v1.astype("I")

    v2 = fc.array("i")
    v2E = v2.astype("E")
    v2I = v2.astype("I")

    # Both operands are length zero
    assert fl.verify(v2I * v0E == v2E)
    assert fl.verify(v2E * v0I == v2E)

    # Automatic broadcast of v1 to zero length
    assert fl.verify(v2I * v1E == v2E)
    assert fl.verify(v2E * v1I == v2E)
    
def test_floordiv(fc):
    # Test case 1
    in_1 = fc.array("i", [4, 5, 5, -15]).astype("I")
    in_2 = fc.array("i", [2, 2, 5, -15]).astype("I")

    assert check_array(in_1 // in_2, 'I', [2, 2, 1, 1])

    # Test case 2
    in_1 = fc.array("i", [4]).astype("I")
    in_2 = fc.array("i", [0]).astype("I")

    try:
        in_1 // in_2
        pytest.fail("division by zero did not cause an error")
    except RuntimeError as e:
        if 'division by zero' not in str(e):
            pytest.fail("error was not division by zero")

def test_floordiv_inplace(fc):
    # Test case 1
    in_1 = fc.array("i", [4, 5, 5, -15]).astype("I")
    in_2 = fc.array("i", [2, 2, 5, -15]).astype("I")

    in_1 //= in_2
    assert check_array(in_1, 'I', [2, 2, 1, 1])

    # Test case 2
    in_1 = fc.array("i", [4]).astype("I")
    in_2 = fc.array("i", [0]).astype("I")

    try:
        in_1 //= in_2
        pytest.fail("division by zero did not cause an error")
    except RuntimeError as e:
        if 'division by zero' not in str(e):
            pytest.fail("error was not division by zero")

def test_setLength(fc):
    # Test Case 1
    v1 = fc.array("i", [20, 40, 50, 60, 70, 80, 100])
    v1b = v1.astype("I")
    v1b.set_length(3)
    assert check_array(v1b, 'I', [20, 40, 50])

def test_setItem_slice(fc):
    values = fc.array("i", [44, 55, 66]).astype("I")

    arr = fc.array("i", [20, 40, 55, 60, 70, 80, 100]).astype("I")

    arr[1:4] = values

    expected = fc.array("i", [20, 44, 55, 66, 70, 80, 100]).astype("I")
    assert fl.verify(arr == expected, False)

def test_setItem_replace(fc):
    values = fc.array("i", [44, 55, 66]).astype("I")

    arr = fc.array("i", [20, 40, 55, 60, 70, 80, 100]).astype("I")

    arr[:] = values

    expected = fc.array("i", [44, 55, 66]).astype("I")
    assert fl.verify(arr == expected, False)
    
def test_index(fc):
    # Test Case 1
    v1 = fc.array("i", [ 1, 2, 0, 4, 0, 6, 7])
    v1b = v1.astype("I")
    v2 = v1b.index()
    v3 = fc.array("i", [0, 1, 3, 5, 6])
    v4 = v2 == v3
    assert fl.verify(v4, False)
    
def test_reducesum(fc):
    # Test Case 1
    v1 = fc.array("i", [ 1, 2, 3, 4, 5, 6, 7 ])
    v1b = v1.astype("I")
    v2 = fc.array("i", [5, 11, 2])
    v2b = v2.astype("I")
    v3 = fc.array("i", [2, 2, 3])
    v1b.reduce_sum(v3, v2b)
    v4 = fc.array("i", [1, 2, 16, 2, 5, 6, 7])
    v4b = v4.astype("I")
    v5 = v1b == v4b
    assert fl.verify(v5, False)
    # See issue #139
    assert fl.verify(v2b == fc.array("i", [5, 11, 2]).astype('I'))
    
    # Test Case 2 - See issue #138
    v6 = fc.array("i", 5, 1).astype("I")
    v7 = fc.array("i", 1, 1)
    v8 = fc.array("i", 5, 1)
    v6.reduce_sum(v7, v8)

def test_delitem(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4])
    v1b = v1.astype("I")
    del v1b[1]
    v2 = fc.array("i", [1, 3, 4])
    v2b = v2.astype("I")
    v3 = v1b == v2b
    assert fl.verify(v3, False)

def test_len(fc):
    # Test Case 1
    v1 = fc.array('i', 100, 5)
    v2 = v1.astype("I").len()
    v3 = fc.array('i', 1, 100)
    v4 = v2 == v3
    assert fl.verify(v4, False)

    # Test Case 2
    v5 = fc.array('i', [1, 2, 3, 4, 5])
    v6 = v5.astype("I").len()
    v7 = fc.array('i', 1, 5)
    v8 = v6 == v7
    assert fl.verify(v8, False)

def test_lookup(fc):
    arr = fc.array('i', [1, 2, 3, 4, 5]).astype("I")
    actual = arr.lookup([1,3])
    expected = fc.array('i', [2,4]).astype("I")

    assert fl.verify(actual == expected, False)
    
def test_pow(fc):
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 30, 2, 10]).astype('I')
    in_2 = fc.array('i', [2, -1, 5, 40, 0])
    out_expected = fc.array('i', [100, 1, 24300000, 1099511627776, 1]).astype('I')
    out_actual = in_1 ** in_2
    assert check_array(out_expected, 'I', out_actual)
    
    # Test Case 2: In-place tests
    out_actual = fc.array('i', [10, 1, 30, 2, 10]).astype('I')
    out_actual **= in_2
    assert check_array(out_actual, 'I', out_expected)

    # Test Case 3: 0 ** 0 should throw an error
    with pytest.raises(RuntimeError):
        in_3 = fc.array('f', [0])
        in_4 = fc.array('i', [0])
        _ = in_3.astype('I') ** in_4  
        
def test_serialise(fc):
    in_1 = fc.array('i', [10, 1, 30, 2]).astype('I')
    in_2 = in_1.serialise()
    assert in_2.typecode() == 'b32'
    
def test_deserialise(fc):
    in_1 = fc.array('i', [10, 1, 30, 2]).astype('I')
    in_2 = in_1.serialise()
    in_3 = fc.array('I')
    in_3.deserialise(in_2)
    assert fl.verify(in_1 == in_3)