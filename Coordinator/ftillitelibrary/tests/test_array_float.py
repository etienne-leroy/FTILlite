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
import math
import pytest

def test_init(fc):
    # Test Case 1
    v1 = fc.array('f', 10) # Only length provided
    assert check_array(v1, 'f', [0.0] * 10)
    
    # Test Case 2
    v2 = fc.array('f', 5, 7.5) # Length and single value provided
    assert check_array(v2, 'f', [7.5] * 5)
    
    # Test Case 3
    v3 = fc.array('f', [-2.22223, 3.4, 1002.332351235]) # Python list of values provided
    assert check_array(v3, 'f', [-2.22223, 3.4, 1002.332351235])
    
    # Test Case 4: Check Python ints are cast to floats
    v4 = fc.array('f', 15, 2) # Length and single value provided
    assert check_array(v4, 'f', [2.0] * 15)
    
    v5 = fc.array('f', [1.0, 2, 5, -7.662, 10]) # Length and single value provided'
    assert check_array(v5, 'f', [1.0, 2, 5, -7.662, 10])

def test_contains(fc):
    # Test Case 1
    v1 = fc.array("f", [1.0, 2.0, 3.0, 4.0, 5.0])
    v2 = fc.array("f", [5.0, 6.0, 7.0, 8.0])
    v3 = v1.contains(v2)
    assert check_array(v3, 'i', [1, 0, 0, 0])
    
def test_cumsum(fc):
    # Test Case 1
    v1 = fc.array("f", [1.1, 2.1, 3.3, 4.4, 5.5, 6.6, 7.7])
    v2 = v1.cumsum()
    assert check_array(v2, 'f', [1.1, 3.2, 6.5, 10.9, 16.4, 23, 30.7])
    
def test_mux(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 0, 1, 2, 0, 0, 3])
    v2 = fc.array("f", [1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0])
    v3 = fc.array("f", [1.1, 2.2, 3.3, 4.4, 5.5, 6.6, 7.7])
    v4 = fl.mux(v1,v2,v3)
    assert check_array(v4, 'f', [1.0, 2.2, 3.0, 4.0, 5.5, 6.6, 7.0])
    
    cond_1 = fc.array('i', [1, 0, 1, 0, 1])
    cond_2 = [1, 0, 1, 0, 1]
    ifttrue_1 = fc.array('f', [4.2])
    ifttrue_2 = 4.2
    iffalse_1 = fc.array('f', [2.4])
    iffalse_2 = 2.4
    expected_out = [4.2, 2.4, 4.2, 2.4, 4.2]

    # Test Case 2: Relates to Issue 57
    v5 = fl.mux(cond_1, ifttrue_1, iffalse_1)
    assert check_array(v5, 'f', expected_out)

    # Test Case 3: Relates to Issue 117
    v6 = fl.mux(cond_1, ifttrue_1, iffalse_2)
    assert check_array(v6, 'f', expected_out)

    v7 = fl.mux(cond_1, ifttrue_2, iffalse_1)
    assert check_array(v7, 'f', expected_out)

    v8 = fl.mux(cond_2, ifttrue_1, iffalse_1)
    assert check_array(v8, 'f', expected_out)

    v9 = fl.mux(cond_2, ifttrue_2, iffalse_1)
    assert check_array(v9, 'f', expected_out)
    
def test_neg(fc):
    # Test Case 1
    v1 = fc.array("f", [10, 20, 30, 40, 50, 60, 70])
    v2 = -v1
    assert check_array(v2, 'f', [-10, -20, -30, -40, -50, -60, -70])

def test_delitem(fc):
    # Test Case 1
    v1 = fc.array("f", [1.0, 2.0, 3.0, 4.0])
    del v1[1]
    assert check_array(v1, 'f', [1.0, 3.0, 4.0])

def test_setItem_slice(fc):
    values = fc.array("f", [44.4, 55.5, 66.6])

    arr = fc.array("f", [20, 40, 55, 60, 70, 80, 100])

    arr[1:4] = values

    expected = fc.array("f", [20, 44.4, 55.5, 66.6, 70, 80, 100])
    assert fl.verify(arr == expected, False)

def test_setItem_replace(fc):
    values = fc.array("f", [44.4, 55.5, 66.6])

    arr = fc.array("f", [20, 40, 55, 60, 70, 80, 100])

    arr[:] = values

    expected = fc.array("f", [44.4, 55.5, 66.6])
    assert fl.verify(arr == expected, False)

def test_index(fc):
    # Test Case 1
    v1 = fc.array('f', [0, 1.555, 0, 0, 0, 0, 42.5, -34.2, 300, 400, -10, -10000, 0, 0, 855])
    v2 = v1.index()
    assert check_array(v2, 'i', [1, 6, 7, 8, 9, 10, 11, 14])

def test_len(fc):
    # Test Case 1
    v1 = fc.array('f', 100, 9.0)
    v2 = v1.len()
    assert check_array(v2, 'i', [100])

    # Test Case 2
    v3 = fc.array('f', [1.5, -2.8, 3333.0, -554, 2003])
    v4 = v3.len()
    assert check_array(v4, 'i', [5])

def test_len_magic_method(fc):
    # Test Case 1
    v1 = fc.array('f', 100, 5.7)
    assert len(v1.tolist()) == 100
    
def test_tolist(fc):
    # Test Case 1
    v1 = fc.array('f', 3, 5.7)
    assert v1.tolist() == [5.7, 5.7, 5.7]
    
def test_truediv(fc):
    # Test Case 1: Basic float division arithmetic
    in_1 = [10.0, 20.25, 30.0001, 2000.3333, -50.5]
    in_2 = [10.0, -15.75, 2000.02, 40.6666, -250.0]
    out_expected = [1.0, (20.25/-15.75), (30.0001/2000.02), (2000.3333/40.6666), (-50.5/-250.0)]
    in_arr_1 = fc.array('f', in_1)
    in_arr_2 = fc.array('f', in_2)
    out_arr_expected = fc.array('f', out_expected)
    out_arr_actual = in_arr_1 / in_arr_2
    assert check_array(out_arr_actual, 'f', out_arr_expected)

def test_truediv_inplace(fc):
    # Test Case 1: Basic float division arithmetic
    in_1 = [10.0, 20.25, 30.0001, 2000.3333, -50.5]
    in_2 = [10.0, -15.75, 2000.02, 40.6666, -250.0]
    out_expected = [1.0, (20.25/-15.75), (30.0001/2000.02), (2000.3333/40.6666), (-50.5/-250.0)]
    in_arr_1 = fc.array('f', in_1)
    in_arr_2 = fc.array('f', in_2)
    out_arr_expected = fc.array('f', out_expected)
    in_arr_1 /= in_arr_2
    assert check_array(in_arr_1, 'f', out_arr_expected)
    
def test_exp(fc):
    in_1 = fc.array('f', [1, -2.3, 20.22])
    out_expected = fc.array('f', [math.exp(1), math.exp(-2.3), math.exp(20.22)])
    out_actual = fl.exp(in_1)
    assert check_array(out_actual, 'f', out_expected)

def test_log(fc):
    in_1 = fc.array('f', [1, 20.22])
    out_expected = fc.array('f', [math.log(1), math.log(20.22)])
    out_actual = fl.log(in_1)
    assert check_array(out_actual, 'f', out_expected)

def test_sin(fc):
    in_1 = fc.array('f', [1, 20.22])
    out_expected = fc.array('f', [math.sin(1), math.sin(20.22)])
    out_actual = fl.sin(in_1)
    assert check_array(out_actual, 'f', out_expected)

def test_cos(fc):
    in_1 = fc.array('f', [1, 20.22])
    out_expected = fc.array('f', [math.cos(1), math.cos(20.22)])
    out_actual = fl.cos(in_1)
    assert check_array(out_actual, 'f', out_expected)
    
def test_pow(fc):
    # Test Case 1
    in_1 = fc.array('f', [10.0, 1.0, 2.5, 2.0, 1.5, -5])
    in_2 = fc.array('i', [2, -1, 5, -2, 0, 6])
    out_expected = fc.array('f', [100, 1, 97.65625, 0.25, 1, 15625])
    out_actual = in_1 ** in_2
    assert check_array(out_actual, 'f', out_expected)
    
    # Test Case 2: In-place tests
    out_actual = fc.array('f', [10.0, 1.0, 2.5, 2.0, 1.5, -5])
    out_actual **= in_2
    assert check_array(out_actual, 'f', out_expected)

    # Test Case 3: 0 ** 0 should throw an error
    with pytest.raises(RuntimeError):
        in_3 = fc.array('f', [0])
        in_4 = fc.array('i', [0])
        _ = in_3 ** in_4   

def test_serialise(fc):
    in_1 = fc.array('f', [10, 1, 30, 2])
    in_2 = in_1.serialise()
    assert in_2.typecode() == 'b8'
    
def test_deserialise(fc):
    in_1 = fc.array('f', [10, 1, 30, 2])
    in_2 = in_1.serialise()
    in_3 = fc.array('f')
    in_3.deserialise(in_2)
    assert fl.verify(in_1 == in_3)