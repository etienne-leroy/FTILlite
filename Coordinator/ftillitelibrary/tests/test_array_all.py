# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from .util import fc, check_array_for_types
import ftillite as fl
import pytest 

def test_add(fc):
    # Test Case 1
    in_1 = [10, 20, 30, 60, -50]
    in_2 = [10, -15, 2000, 40, -50]
    out_expected = [20, 5, 2030, 100, -100]
    test_tcs = ['i', 'f', 'I', 'E']
    def op(x, y): return x + y
    assert all(check_array_for_types(
        fc, [in_1, in_2], out_expected, op, types_supported=test_tcs))

def test_add_inplace(fc):
    # Test Case 1
    in_1 = [10, 20, 30, 60, -50]
    in_2 = [10, -15, 2000, 40, -50]
    out_expected = [20, 5, 2030, 100, -100]
    test_tcs = ['i', 'f', 'I', 'E']
    def op(x, y): 
        x += y
        return x
    assert all(check_array_for_types(
        fc, [in_1, in_2], out_expected, op, types_supported=test_tcs))

def test_subtract(fc):
    # Test Case 1
    in_1 = [10, 20, 30, 50, -50]
    in_2 = [10, -15, 2000, 40, -50]
    out_expected = [0, 35, -1970, 10, 0]
    test_tcs = ['i', 'f', 'I', 'E']
    def op(x, y): return x - y
    assert all(check_array_for_types(
        fc, [in_1, in_2], out_expected, op, types_supported=test_tcs))

def test_subtract_inplace(fc):
    # Test Case 1
    in_1 = [10, 20, 30, 50, -50]
    in_2 = [10, -15, 2000, 40, -50]
    out_expected = [0, 35, -1970, 10, 0]
    test_tcs = ['i', 'f', 'I', 'E']
    def op(x, y): 
        x -= y
        return x
    assert all(check_array_for_types(
        fc, [in_1, in_2], out_expected, op, types_supported=test_tcs))

def test_multiply(fc):
    # Test Case 1
    in_1 = [10, 20, 30, 50, -50]
    in_2 = [10, -15, 2000, 40, -50]
    out_expected = [100, -300, 60000, 2000, 2500]
    test_tcs = ['i', 'f', 'I']
    def op(x, y): return x * y
    assert all(check_array_for_types(
        fc, [in_1, in_2], out_expected, op, types_supported=test_tcs))

def test_multiply_inplace(fc):
    # Test Case 1
    in_1 = [10, 20, 30, 50, -50]
    in_2 = [10, -15, 2000, 40, -50]
    out_expected = [100, -300, 60000, 2000, 2500]
    test_tcs = ['i', 'f', 'I']
    def op(x, y): 
        x *= y
        return x
    assert all(check_array_for_types(
        fc, [in_1, in_2], out_expected, op, types_supported=test_tcs))


def test_truediv(fc):
    # Test Case 1
    in_1 = [10, -15, 2000, -15]
    in_2 = [10, 5, 40, -5]

    out_expected = [1, -3, 50, 3]

    test_tcs = ['i', 'f', 'I']
    def op(x, y): return x / y

    assert all(check_array_for_types(
        fc, [in_1, in_2], out_expected, op, types_supported=test_tcs))

    # Test case 2
    in_1 = fc.array("i", [4])
    in_2 = fc.array("i", [0])

    for t in test_tcs:
        try:
            in_1.astype(t) / in_2.astype(t)
            pytest.fail(f"division by zero did not cause an error for Array({t})")
        except RuntimeError as e:
            if 'divide by zero' not in str(e) and 'division by zero' not in str(e):
                pytest.fail("error was not division by zero")

def test_truediv_inplace(fc):
    # Test Case 1
    in_1 = [10, -15, 2000, -15]
    in_2 = [10, 5, 40, -5]

    out_expected = [1, -3, 50, 3]

    test_tcs = ['i', 'f', 'I']
    def op(x, y): 
        x /= y
        return x

    assert all(check_array_for_types(
        fc, [in_1, in_2], out_expected, op, types_supported=test_tcs))
    
    # Test case 2
    in_1 = fc.array("i", [4])
    in_2 = fc.array("i", [0])

    for t in test_tcs:
        try:
            in_3 = in_1.astype(t)
            in_3 /= in_2.astype(t)
            pytest.fail(f"division by zero did not cause an error for Array({t})")
        except RuntimeError as e:
            if 'divide by zero' not in str(e) and 'division by zero' not in str(e):
                pytest.fail("error was not division by zero")




def test_unflatten(fc):
    in_1 = [10, 20, 30, 40]
    in_2 = [20, 30, 40, 50]
    arr_types = ['i', 'f', 'I']
    
    if fc._backend.gpuenabled:
        arr_types += 'E'
    for tc in arr_types:
        # Test Case 1
        in_arr_1 = fc.array('i', in_1).astype(tc)
        in_arr_2 = fc.array('i', in_2).astype(tc)
        in_arr_3 = in_arr_1.unflatten([in_arr_2])
        assert isinstance(in_arr_3, type(in_arr_1))
        assert fl.verify(in_arr_2 == in_arr_3)

        # Test Case 2: check non-list input for param 'data' raises an exception
        with pytest.raises(TypeError):
            _ = in_arr_1.unflatten(in_arr_2)

        # Test Case 3: check list input for param 'data' with a length != 1 raises an exception
        with pytest.raises(TypeError):
            _ = in_arr_1.unflatten([in_arr_1, in_arr_2])

        # Test Case 4: check mismatched type for param 'data' raises an exception
        with pytest.raises(TypeError):
            if tc != 'i':
                _ = in_arr_1.unflatten([fc.array('i', in_1)])
            else:
                _ = in_arr_1.unflatten([fc.array('f', in_1)])