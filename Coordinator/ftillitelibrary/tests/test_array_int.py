# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from .util import fc, check_array, check_array_for_types
import ftillite as fl
import pytest

def test_init(fc):
    # Test Case 1
    v1 = fc.array('i', 10) # Only length provided
    assert check_array(v1, 'i', [0] * 10)
    
    # Test Case 2
    v2 = fc.array('i', 5, 7) # Length and single value provided
    assert check_array(v2, 'i', [7] * 5)
    
    # Test Case 3
    v3 = fc.array('i', [2, 3, 4]) # Python list of values provided
    assert check_array(v3, 'i', [2, 3, 4])
    
    # Test Case 4: Floats cannot be cast to ints
    with pytest.raises(TypeError):
        _ = fc.array('i', 10, 5.5)
    
    with pytest.raises(TypeError):
        _ = fc.array('i', [1, 2, 3.5, 4, -9.22])

    # del v1, v2, v3
    
def test_arange(fc):
    # Test Case 1
    v1 = fc.arange(10)
    assert check_array(v1, 'i', [i for i in range(10)])
    
    # Test Case 2: Floats shouldn't be accepted as input
    try:
        _ = fc.arange(10.5)
        assert False
    except TypeError:
        assert True

def test_setItem_slice(fc):
    values = fc.array("i", [44, 55, 66])

    arr = fc.array("i", [20, 40, 55, 60, 70, 80, 100])

    arr[1:4] = values

    expected = fc.array("i", [20, 44, 55, 66, 70, 80, 100])
    assert fl.verify(arr == expected, False)

def test_setItem_replace(fc):
    values = fc.array("i", [44, 55, 66])

    arr = fc.array("i", [20, 40, 55, 60, 70, 80, 100])

    arr[:] = values

    expected = fc.array("i", [44, 55, 66])
    assert fl.verify(arr == expected, False)

def test_contains(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4, 5])
    v2 = fc.array("i", [5, 6, 7, 8])
    v3 = v1.contains(v2)
    assert check_array(v3, 'i', [1, 0, 0, 0])
    
def test_cumsum(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4, 5, 6, 7])
    v2 = v1.cumsum()
    assert check_array(v2, 'i', [1, 3, 6, 10, 15, 21, 28])
    
def test_mux(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 0, 1, 2, 0, 0, 3])
    v2 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v3 = fc.array("i", [11, 22, 33, 44, 55, 66, 77])
    v4 = fl.mux(v1, v2, v3)
    assert check_array(v4, 'i', [10, 22, 30, 40, 55, 66, 70])

    cond_1 = fc.array('i', [1, 0, 1, 0, 1])
    cond_2 = [1, 0, 1, 0, 1]
    ifttrue_1 = fc.array('i', [4])
    ifttrue_2 = 4
    iffalse_1 = fc.array('i', [2])
    iffalse_2 = 2
    expected_out = [4, 2, 4, 2, 4]

    # Test Case 2: Relates to Issue 57
    v5 = fl.mux(cond_1, ifttrue_1, iffalse_1)
    assert check_array(v5, 'i', expected_out)

    # Test Case 3: Relates to Issue 117
    v6 = fl.mux(cond_1, ifttrue_1, iffalse_2)
    assert check_array(v6, 'i', expected_out)

    v7 = fl.mux(cond_1, ifttrue_2, iffalse_1)
    assert check_array(v7, 'i', expected_out)

    v8 = fl.mux(cond_2, ifttrue_1, iffalse_1)
    assert check_array(v8, 'i', expected_out)

    v9 = fl.mux(cond_2, ifttrue_2, iffalse_1)
    assert check_array(v9, 'i', expected_out)

def test_neg(fc):
    # Test Case 1
    v1 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v2 = -v1
    assert check_array(v2, 'i', [-10, -20, -30, -40, -50, -60, -70])

def test_delitem(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4])
    del v1[1]
    assert check_array(v1, 'i', [1, 3, 4])

def test_index(fc):
    # Test Case 1
    v1 = fc.array('i', [0, 0, 200, 0, 300, 400, -10, -10000, 0, 0, 855])
    v2 = v1.index()
    assert check_array(v2, 'i', [2, 4, 5, 6, 7, 10])

def test_len(fc):
    # Test Case 1
    v1 = fc.array('i', 100, 5)
    v2 = v1.len()
    assert check_array(v2, 'i', [100])

    # Test Case 2
    v3 = fc.array('i', [1, 2, 3, 4, 5])
    v4 = v3.len()
    assert check_array(v4, 'i', [5])

def test_len_magic_method(fc):
    # Test Case 1
    v1 = fc.array('i', 100, 5)
    assert len(v1.tolist()) == 100

def test_getitem(fc):
    # Test Case 1: Check python native index works
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = v1[2]
    v3 = fc.array('i', [3])
    v4 = v2 == v3
    assert fl.verify(v4, False)
    
    # Test Case 2: Check singleton index works
    v2 = fc.array('i', [2])
    v3 = v1[v2]
    v4 = fc.array('i', [3])
    v5 = v4 == v3
    assert fl.verify(v5, False)
    
    # Test Case 3: Check array index works
    v2 = fc.array('i', [1, 3, 4])
    v3 = v1[v2]
    v4 = fc.array('i', [2, 4, 5])
    v5 = v3 == v4
    assert fl.verify(v5, False)
    
    # Test Case 4: Check slices with python native values
    v2 = v1[1:3]
    v3 = v1[-3:-1]
    v4 = v1[1:]
    v5 = v1[:3]
    v6 = fc.array('i', [2, 3])
    v7 = fc.array('i', [3, 4])
    v8 = fc.array('i', [2, 3, 4, 5])
    v9 = fc.array('i', [1, 2, 3])
    v10 = v2 == v6
    v11 = v3 == v7
    v12 = v4 == v8
    v13 = v5 == v9
    assert fl.verify(v10, False)
    assert fl.verify(v11, False)
    assert fl.verify(v12, False)
    assert fl.verify(v13, False)
    
    # Test Case 5: Check slices with singleton values
    v2 = (fc.array('i', [2]), fc.array('i', [4])) 
    v3 = (fc.array('i', [-3]), fc.array('i', [-1])) 
    v4 = v1[v2[0]:v2[1]]
    v5 = v1[v3[0]:v3[1]]
    v6 = v1[v2[0]:] # 2:
    v7 = v1[:v2[1]] # :4
    v8 = fc.array('i', [3, 4])
    v9 = fc.array('i', [3, 4])
    v10 = fc.array('i', [3, 4, 5])
    v11 = fc.array('i', [1, 2, 3, 4])
    v12 = v4 == v8
    v13 = v5 == v9
    v14 = v6 == v10
    v15 = v7 == v11
    assert fl.verify(v12, False)
    assert fl.verify(v13, False)
    assert fl.verify(v14, False)
    assert fl.verify(v15, False)
    
    # Test Case 6: Should not be able to get index greater than the length of the array
    try:
        v2 = v1[100]
        assert False
    except RuntimeError as ex:
        assert True    
     
    try:
        v2 = v1[1:100]
        assert False
    except RuntimeError as ex:
        assert True
        
    # Test Case 7: Check select all
    v2 = v1[:]
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 8: Empty array checks
    v1 = fc.array('i')
    v2 = v1[:]
    v3 = v1 == v2
    assert fl.verify(v3, False)

def test_setitem(fc):
    # Test Case 1: Check basic python index works
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v1[0] = 10
    v2 = fc.array('i', [10, 2, 3, 4, 5])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 2: Check singleton array indexes and values work
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [1])
    v3 = fc.array('i', [2])
    v1[v2] = 100
    v1[v3] = v2
    v4 = fc.array('i', [1, 100, 1, 4, 5])
    v5 = v1 == v4
    assert fl.verify(v5, False)
    
    # Test Case 3: Check Python slices work
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v1[1:3] = [100, 200]
    v2 = fc.array('i', [1, 100, 200, 4, 5])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 4: Check Python slices work with broadcast value
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [75])
    v1[1:3] = 50
    v1[3:5] = v2
    v2 = fc.array('i', [1, 50, 50, 75, 75])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 5: Check Singleton slices work
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [1])
    v3 = fc.array('i', [3])
    v1[v2:v3] = [100, 200]
    v2 = fc.array('i', [1, 100, 200, 4, 5])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 6: Check Singleton slices work with broadcast value
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [1])
    v3 = fc.array('i', [3])
    v4 = fc.array('i', [3])
    v5 = fc.array('i', [5])
    v6 = fc.array('i', [60])
    v1[v2:v3] = 90
    v1[v4:v5] = v6
    v7 = fc.array('i', [1, 90, 90, 60, 60])
    v8 = v1 == v7
    assert fl.verify(v8, False)
    
    # Test Case 7: Check mixture of singleton and python values work
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [1])
    v3 = fc.array('i', [5])
    v1[v2:3] = -20
    v1[3:v3] = 120
    v4 = fc.array('i', [1, -20, -20, 120, 120])
    v5 = v1 == v4
    assert fl.verify(v5, False)
    
    # Test Case 8: Check int array index work
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [1, 2, 3])
    v1[v2] = [500, 600, -700]
    v3 = fc.array('i', [1, 500, 600, -700, 5])
    v4 = v1 == v3
    assert fl.verify(v4, False)
    
    # Test Case 9: Check setting all items with python native value
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v1[:] = -5
    v2 = fc.array('i', [-5, -5 , -5, -5, -5])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 10: Check setting all items with singleton array
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v1[:] = fc.array('i', 1, 25)
    v2 = fc.array('i', [25, 25 , 25, 25, 25])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 11: Check asetting all items with python list of same length
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v1[:] = [6, 7, 8, 9, 10, 11, 12, 13]
    v2 = fc.array('i', [6, 7, 8, 9, 10, 11, 12, 13])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 12: Check asetting all items with array of same length
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v1[:] = fc.array('i', [6, 7, 8, 9, 10])
    v2 = fc.array('i', [6, 7, 8, 9, 10])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 13: Should not be able to get index greater than the length of the array
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    try:
        v1[500] = 5
        assert False
    except RuntimeError:
        assert True
        
    try:
        v1[20:30] = 5
        assert False
    except RuntimeError:
        assert True
        
    try:
        v1[-20] = 5
        assert False
    except RuntimeError:
        assert True
        
    v2 = fc.array('i', [20])
    try:
        v1[v2] = 5
        assert False
    except RuntimeError:
        assert True
        
    v2 = fc.array('i', [20])
    try:
        v1[v2] = 5
        assert False
    except RuntimeError:
        assert True
        
    v3 = fc.array('i', [21])
    try:
        v1[v2:v3] = 5
        assert False
    except RuntimeError:
        assert True
        
    v4 = fc.array('i', [1, 20, 5])
    try:
        v1[v4] = 5
        assert False
    except RuntimeError:
        assert True
        
    v5 = fc.array('i', [1, 2, 3, 4, 5])
    v6 = v1 == v5
    assert fl.verify(v5, False)
        
    # Test Case 14: Start value of a slice should be less than or equal to stop value
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    try:
        v1[2:1] = 5
        assert False
    except RuntimeError:
        assert True
        
    # Test Case 15: Check syntax [:x]
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v1[:3] = -200
    v2 = fc.array('i', [-200, -200, -200, 4, 5])
    v3 = v1 == v2
    assert fl.verify(v3, False) 
    
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [3])
    v1[:v2] = -200
    v3 = fc.array('i', [-200, -200, -200, 4, 5])
    v4 = v1 == v3
    assert fl.verify(v4, False) 
    
    # Test Case 16: Check syntax [x:]
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v1[3:] = -200
    v2 = fc.array('i', [1, 2, 3, -200, -200])
    v3 = v1 == v2
    assert fl.verify(v3, False)       
    
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [3])
    v1[v2:] = -200
    v3 = fc.array('i', [1, 2, 3, -200, -200])
    v4 = v1 == v3
    assert fl.verify(v4, False) 
    
    # Test Case 17: See Issue 76
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    v2 = v1.len()
    v1.set_length(2*v2)
    v1[v2:] = v1[:v2]
    v3 = fc.array('i', [1, 2, 3, 4, 5, 1, 2, 3, 4, 5])
    v4 = v1 == v3
    assert fl.verify(v4, False)
    
    # Test Case 18: Empty array checks
    v1 = fc.array('i')
    v1[:] = fc.array('i', [1, 2, 3, 4, 5])
    v2 = fc.array('i', [1, 2, 3, 4, 5])
    v3 = v1 == v2
    assert fl.verify(v3, False)
    
    # Test Case 19: Shouldn't be able to extend array length
    v1 = fc.array('i', [1, 2, 3, 4, 5])
    try:
        v1[1:10] = 5
        assert False
    except RuntimeError:
        assert True
        
def test_tolist(fc):
    # Test Case 1
    v1 = fc.array('i', 5, 10)
    assert v1.tolist() == [10, 10, 10, 10, 10]
    
def test_floordiv(fc):
    # Test Case 1
    in_1 = fc.array('i',[10, -15, 2000, -15])
    in_2 = fc.array('i',[10, 5, 40, -5])

    out_expected = fc.array('i',[1, -3, 50, 3])

    out_arr_actual = in_1 // in_2
    assert check_array(out_arr_actual, 'i', out_expected)

    # Test case 2
    in_1 = fc.array("i", [4])
    in_2 = fc.array("i", [0])

    try:
        in_1 // in_2
        pytest.fail("division by zero did not cause an error")
    except RuntimeError as e:
        if 'integer divide by zero' not in str(e):
            pytest.fail("error was not division by zero")

def test_floordiv_inplace(fc):
    # Test Case 1
    in_1 = fc.array('i',[10, -15, 2000, -15])
    in_2 = fc.array('i',[10, 5, 40, -5])

    out_expected = fc.array('i',[1, -3, 50, 3])

    in_1 //= in_2
    assert check_array(in_1, 'i', out_expected)

    # Test case 2
    in_1 = fc.array("i", [4])
    in_2 = fc.array("i", [0])

    try:
        in_1 //= in_2
        pytest.fail("division by zero did not cause an error")
    except RuntimeError as e:
        if 'integer divide by zero' not in str(e):
            pytest.fail("error was not division by zero")
    
def test_truediv(fc):
    # Test Case 1: Basic int division arithmetic - where results round to negative infinity
    in_1 = [10, 20, 30, 2000, -50]
    in_2 = [10, -15, 2000, 40, -250]
    out_expected = [1, -2, 0, 50, 0]
    in_arr_1 = fc.array('i', in_1)
    in_arr_2 = fc.array('i', in_2)
    out_arr_expected = fc.array('i', out_expected)
    out_arr_actual = in_arr_1 / in_arr_2
    assert check_array(out_arr_actual, 'i', out_arr_expected)
    
    # Test case 2
    in_1 = fc.array("i", [4])
    in_2 = fc.array("i", [0])

    try:
        in_1 / in_2
        pytest.fail("division by zero did not cause an error")
    except RuntimeError as e:
        if 'integer divide by zero' not in str(e):
            pytest.fail("error was not division by zero")

def test_truediv_inplace(fc):
    # Test Case 1: Basic int division arithmetic - where results round to negative infinity
    in_1 = [10, 20, 30, 2000, -50]
    in_2 = [10, -15, 2000, 40, -250]
    out_expected = [1, -2, 0, 50, 0]
    in_arr_1 = fc.array('i', in_1)
    in_arr_2 = fc.array('i', in_2)
    out_arr_expected = fc.array('i', out_expected)
    in_arr_1 /= in_arr_2
    assert check_array(in_arr_1, 'i', out_arr_expected)
    
    # Test case 2
    in_1 = fc.array("i", [4])
    in_2 = fc.array("i", [0])

    try:
        in_1 /= in_2
        pytest.fail("division by zero did not cause an error")
    except RuntimeError as e:
        if 'integer divide by zero' not in str(e):
            pytest.fail("error was not division by zero")


def test_divmod(fc):
    # Test Case 1: Divmod for int that returns the division result and modulo
    in_1 = [10, 20, 30, 25, -50]
    in_2 = [10, -15, 2000, 3, -250]
    out_expected_1 = [1, -2, 0, 8, 0]
    out_expected_2 = [0, -10, 30, 1, -50]
    in_arr_1 = fc.array('i', in_1)
    in_arr_2 = fc.array('i', in_2)
    out_arr_expected_1 = fc.array('i', out_expected_1)
    out_arr_expected_2 = fc.array('i', out_expected_2)
    out_arr_actual_1, out_arr_actual_2 = divmod(in_arr_1, in_arr_2)
    assert check_array(out_arr_actual_1, 'i', out_arr_expected_1)
    assert check_array(out_arr_actual_2, 'i', out_arr_expected_2)
    
def test_lshift(fc):
    # Test Case 1
    in_1 = [10, 1, -10]
    in_2 = [2, 5, 5]
    out_expected = [40, 32, -320]
    op = lambda x, y: x << y
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
    
    # TODO: fail cases i.e. negative value for shift i.e. x << y, y cannot be negative

def test_lshift_inplace(fc):
    # Test Case 1
    in_1 = [10, 1, -10]
    in_2 = [2, 5, 5]
    out_expected = [40, 32, -320]
    def op(x,y):
        x <<= y
        return x
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))    
    
def test_rshift(fc):
    # Test Case 1
    in_1 = [10, 1, -10]
    in_2 = [2, 5, 5]
    out_expected = [2, 0, -1]
    op = lambda x, y: x >> y
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
    
    # TODO: fail cases i.e. negative value for shift i.e. x >> y, y cannot be negative

def test_rshift_inplace(fc):
    # Test Case 1
    in_1 = [10, 1, -10]
    in_2 = [2, 5, 5]
    out_expected = [2, 0, -1]
    def op(x,y):
        x >>= y
        return x
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
    
def test_and(fc):
    # Test Case 1
    in_1 = [10, 1, 0, 1, 1, -1]
    in_2 = [2, 5, 1, 0, -1, -5]
    out_expected = [2, 1, 0, 0, 1, -5]
    op = lambda x, y: x & y
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
    
def test_and_inplace(fc):
    # Test Case 1
    in_1 = [10, 1, 0, 1, 1, -1]
    in_2 = [2, 5, 1, 0, -1, -5]
    out_expected = [2, 1, 0, 0, 1, -5]
    def op(x,y):
        x &= y
        return x
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
   
def test_or(fc):
    # Test Case 1
    in_1 = [10, 1, 0, 1, 1, -1]
    in_2 = [2, 5, 1, 0, -1, -5]
    out_expected = [10, 5, 1, 1, -1, -1]
    op = lambda x, y: x | y
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
    
def test_or_inplace(fc):
    # Test Case 1
    in_1 = [10, 1, 0, 1, 1, -1]
    in_2 = [2, 5, 1, 0, -1, -5]
    out_expected = [10, 5, 1, 1, -1, -1]
    def op(x, y):
        x |= y
        return x
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
    
def test_xor(fc):
    # Test Case 1
    in_1 = [10, 1, 0, 1, 1, -1]
    in_2 = [2, 5, 1, 0, -1, -5]
    out_expected = [8, 4, 1, 1, -2, 4]
    op = lambda x, y: x ^ y
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
    
def test_xor_inplace(fc):
    # Test Case 1
    in_1 = [10, 1, 0, 1, 1, -1]
    in_2 = [2, 5, 1, 0, -1, -5]
    out_expected = [8, 4, 1, 1, -2, 4]
    def op(x, y):
        x ^= y
        return x
    assert all(check_array_for_types(fc, [in_1, in_2], out_expected, op, types_supported=['i']))
    
def test_invert(fc):
    # Test Case 1
    in_1 = [10, 1, 0, -10]
    out_expected = [-11, -2, -1, 9]
    op = lambda x: ~x
    assert all(check_array_for_types(fc, [in_1], out_expected, op, types_supported=['i']))
    
    # Note: no inplace operator for unary operators.
    
def test_nearest(fc):
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 0, 1, 1, -1])
    out_expected = fc.array('f', [10.0, 1.0, 0.0, 1.0, 1.0, -1.0])
    out_actual = fl.nearest(in_1)
    assert check_array(out_actual, 'f', out_expected)
    
def test_verify(fc):
    # Test Case 1
    arr1 = fc.arange(10)
    assert not fl.verify(arr1, False)
    
    # Test Case 2
    arr2 = fc.array('i', 10, 1)
    assert fl.verify(arr2, False)
    
    # Test Case 3: verify can only accept IntArrayIdentifier as a param
    arr2 = fc.array('f', 10, 1)
    try:
        fl.verify(arr2, False)
        assert False
    except TypeError:
        assert True
        
def test_tolist(fc):
    # Test Case 1
    v1 = fc.array('i', 5, 10)
    assert v1.tolist() == [10, 10, 10, 10, 10]
    
def test_pow(fc):
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 30, 2, 10, -5, -100])
    in_2 = fc.array('i', [2, -1, 5, 40, 0, 2, 1])
    out_expected = fc.array('i', [100, 1, 24300000, 1099511627776, 1, 25, -100])
    out_actual = in_1 ** in_2
    assert check_array(out_actual, 'i', out_expected)

    # Test Case 2: In-place tests
    out_actual = fc.array('i', [10, 1, 30, 2, 10, -5, -100])
    out_actual **= in_2
    assert check_array(out_actual, 'i', out_expected)

    # Test Case 3: 0 ** 0 should throw an error
    with pytest.raises(RuntimeError):
        in_3 = fc.array('i', [0])
        in_4 = fc.array('i', [0])
        _ = in_3 ** in_4        
        
def test_serialise(fc):
    in_1 = fc.array('i', [10, 1, 30, 2])
    in_2 = in_1.serialise()
    assert in_2.typecode() == 'b8'
    
def test_deserialise(fc):
    in_1 = fc.array('i', [10, 1, 30, 2])
    in_2 = in_1.serialise()
    in_3 = fc.array('i')
    in_3.deserialise(in_2)
    assert fl.verify(in_1 == in_3)