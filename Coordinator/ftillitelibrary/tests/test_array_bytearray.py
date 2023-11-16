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

check_type = lambda x, l: x.typecode() == f"b{l}"

def test_init(fc):
    # Test Case 1
    v1 = fc.array("b32", 30)
    assert check_array(v1, "b32", [0] * 30)

def test_contains(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4, 5])
    v2 = v1.astype("b8")
    v3 = fc.array("i", [5, 6, 7, 8])
    v4 = v3.astype("b8")
    v5 = v2.contains(v4)
    assert check_array(v5, "i", [1, 0, 0, 0])

def test_cumsum(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4, 5])
    v2 = v1.astype("b8")
   
    # cumsum using logical OR

    # 0|1  1|2  3|3  3|4  7|5
    # = 1  = 3  = 3  = 7  = 7

    # 0000|0001 0001|0010 0011|0011 0011|0100 0111|0101
    # 0001      0011      0011      0111      0111

    v3 = v2.cumsum()

    assert check_array(v3, "b8", [1, 3, 3, 7, 7])
    
def test_mux(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 0, 1, 2, 0, 0, 3])
    v2 = fc.array("i", [10, 20, 30, 40, 50, 60, 70])
    v2b = v2.astype("b8")
    v3 = fc.array("i", [11, 22, 33, 44, 55, 66, 77])
    v3b = v3.astype("b8")
    v4 = fl.mux(v1,v2b,v3b)
    assert check_array(v4, "b8", [10, 22, 30, 40, 55, 66, 70])

    # Test Case 2: Relates to Issue 57
    cond = fc.array('i', [1, 0, 1, 0, 1])
    ifttrue = fc.array('i', [8]).astype("b8")
    iffalse = fc.array('i', [4]).astype("b8")
    v5 = fl.mux(cond, ifttrue, iffalse)
    assert check_array(v5, "b8", [8, 4, 8, 4, 8])

def test_delitem(fc):
    # Test Case 1
    v1 = fc.array("i", [1, 2, 3, 4])
    v1b = v1.astype("b8")
    del v1b[1]
    assert check_array(v1b, "b8", [1, 3, 4])

def test_index(fc):
    # Test Case 1
    v1 = fc.array('i', [0, 0, 200, 0, 300, 400, -10, -10000, 0, 0, 855])
    v2 = v1.astype('b8').index()
    assert check_array(v2, 'i', [2, 4, 5, 6, 7, 10])

def test_len(fc):
    # Test Case 1
    v1 = fc.array('i', 100, 5)
    v2 = v1.astype('b8').len()
    assert check_array(v2, 'i', [100])

    # Test Case 2
    v3 = fc.array('i', [1, 2, 3, 4, 5])
    v4 = v3.astype('b8').len()
    assert check_array(v4, 'i', [5])

def test_setItem_slice(fc):
    values = fc.array("i", [44, 55, 66]).astype("b8")

    arr = fc.array("i", [20, 40, 55, 60, 70, 80, 100]).astype("b8")

    arr[1:4] = values

    expected = fc.array("i", [20, 44, 55, 66, 70, 80, 100]).astype("b8")
    assert fl.verify(arr == expected, False)

def test_setItem_replace(fc):
    values = fc.array("i", [44, 55, 66]).astype("b8")

    arr = fc.array("i", [20, 40, 55, 60, 70, 80, 100]).astype("b8")

    arr[:] = values

    expected = fc.array("i", [44, 55, 66]).astype("b8")
    assert fl.verify(arr == expected, False)
    
def test_concat(fc):
    a = fc.array("i", [1,2,3,4,5]).astype("b8")
    b = fc.array("i", [1,2,3,4,5]).astype("b8")
    c = a.concat(b)

    assert c.size() == 16

def test_concat_with_broadcast(fc):
    a = fc.array("i", [1,2,3,4,5]).astype("b8")
    b = fc.array("i", [1]).astype("b8")
    c = a.concat(b)

    assert c.size() == 16


def test_byteproject(fc):

    # [
    #    { 01000001 00000000 ... } 65
    #    { 00000000 01111111 ... } 32512
    #    { 00000001 00000000 ... } 1
    # ]

    # [
    #    { 00000000 01000001 ... } 16640
    #    { 01111111 00000000 ... } 127
    #    { 00000000 00000001 ... } 256
    # ]

    a = fc.array("i", [65, 32512, 1]).astype("b8")
    b = a.byteproject(2, {
        0: 1,
        1: 0
    })
    c = fc.array("i", [16640, 127, 256]).astype("b8").byteproject(2, {
        0: 0,
        1: 1
    })

    assert fl.verify(b == c, False)

def test_lshift(fc):
    # [
    #   {00001010 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {00000001 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {11111110 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    # ]
    in_1 = fc.array('i', [10, 1, 254]).astype('b16')
    in_2 = fc.array('i', [2, 2, 2])

    # [
    #   {00101000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {00001000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {11111000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    # ]
    x = fc.array('i', [40, 4, 248]).astype('b8')
    zeroes = fc.array('i', [0,0, 0]).astype('b8')

    out_expected = x.concat(zeroes)

    out_actual = in_1 << in_2

    assert fl.verify(out_expected == out_actual, False)
    # TODO: fail cases i.e. negative value for shift i.e. x << y, y cannot be negative

def test_lshift_inplace(fc):
    # [
    #   {00001010 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {00000001 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {11111110 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    # ]
    in_1 = fc.array('i', [10, 1, 254]).astype('b16')
    in_2 = fc.array('i', [2, 2, 2])

    # [
    #   {00101000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {00001000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {11111000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    # ]
    x = fc.array('i', [40, 4, 248]).astype('b8')
    zeroes = fc.array('i', [0,0, 0]).astype('b8')

    out_expected = x.concat(zeroes)

    in_1 <<= in_2

    assert fl.verify(out_expected == in_1, False)

def test_rshift(fc):
    # [
    #   {0001010 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {0000001 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    # ]
    in_1 = fc.array('i', [10, 1]).astype('b16')
    in_2 = fc.array('i', [64, 64])

    # [
    #   {00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00001010 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000001 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    # ]
    zeroes = fc.array('i', [0,0]).astype('b8')
    x = fc.array('i', [10, 1]).astype('b8')
    
    out_expected = zeroes.concat(x)
    out_actual = in_1 >> in_2

    assert fl.verify(out_expected == out_actual, False)
    # TODO: fail cases i.e. negative value for shift i.e. x >> y, y cannot be negative

def test_rshift_inplace(fc):
    # [
    #   {0001010 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {0000001 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    # ]
    in_1 = fc.array('i', [10, 1]).astype('b16')
    in_2 = fc.array('i', [64, 64])

    # [
    #   {00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00001010 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    #   {00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000001 00000000 00000000 00000000 00000000 00000000 00000000 00000000}
    # ]
    zeroes = fc.array('i', [0,0]).astype('b8')
    x = fc.array('i', [10, 1]).astype('b8')
    
    out_expected = zeroes.concat(x)
    in_1 >>= in_2

    assert fl.verify(out_expected == in_1, False)

def test_tolist(fc):
    # Test Case 1
    v1 = fc.array('i', 2, 5)
    v2 = v1.astype('b8')
    v3 = v2.tolist()
    assert v3 == [['5', '0', '0', '0','0', '0', '0', '0'] for _ in range(2)]

def test_and(fc):
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 0, 1, 1, -1]).astype('b32')
    in_2 = fc.array('i', [2, 5, 1, 0, -1, -5]).astype('b32')
    out_expected = fc.array('i', [2, 1, 0, 0, 1, -5]).astype('b32')
    out_actual = in_1 & in_2
    assert check_array(out_actual, 'b32', out_expected)


def test_and_inplace(fc):
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 0, 1, 1, -1]).astype('b32')
    in_2 = fc.array('i', [2, 5, 1, 0, -1, -5]).astype('b32')
    out_expected = fc.array('i', [2, 1, 0, 0, 1, -5]).astype('b32')
    in_1 &= in_2
    assert check_array(in_1, 'b32', out_expected)
   
def test_or(fc):
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 0, 1, 1, -1]).astype('b32')
    in_2 = fc.array('i', [2, 5, 1, 0, -1, -5]).astype('b32')
    out_expected = fc.array('i', [10, 5, 1, 1, -1, -1]).astype('b32')
    out_actual = in_1 | in_2
    assert check_array(out_actual, 'b32', out_expected)

def test_or_inplace(fc):
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 0, 1, 1, -1]).astype('b32')
    in_2 = fc.array('i', [2, 5, 1, 0, -1, -5]).astype('b32')
    out_expected = fc.array('i', [10, 5, 1, 1, -1, -1]).astype('b32')
    in_1 |= in_2
    assert check_array(in_1, 'b32', out_expected)
    
def test_xor(fc):   
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 0, 1, 1, -1]).astype('b32')
    in_2 = fc.array('i', [2, 5, 1, 0, -1, -5]).astype('b32')
    out_expected = fc.array('i', [8, 4, 1, 1, -2, 4]).astype('b32')
    out_actual = in_1 ^ in_2
    assert check_array(out_actual, 'b32', out_expected)

def test_xor_inplace(fc):   
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 0, 1, 1, -1]).astype('b32')
    in_2 = fc.array('i', [2, 5, 1, 0, -1, -5]).astype('b32')
    out_expected = fc.array('i', [8, 4, 1, 1, -2, 4]).astype('b32')
    in_1 ^= in_2
    assert check_array(in_1, 'b32', out_expected)
    
def test_invert(fc):
    # Test Case 1
    in_1 = fc.array('i', [10, 1, 0, -10]).astype('b32')
    out_expected = fc.array('i', [-11, -2, -1, 9]).astype('b32')
    out_actual = ~in_1
    assert check_array(out_actual, 'b32', out_expected)

    # No in-place test since this is a unary operator.
    
def test_copy(fc):
    # Test Case 1
    in_1 = fc.array('b32', 0)
    in_1.set_length(10)
    in_2 = fc.array('b32', 10)
    assert fl.verify(in_1 == in_2)
    in_3 = in_1.copy()
    assert in_1.typecode() == in_3.typecode()
     
def test_serialise(fc):
    in_1 = fc.array('i', [10, 1, 30, 2]).astype('b32')
    in_2 = in_1.serialise()
    assert in_2.typecode() == 'b32'
    
def test_deserialise(fc):
    in_1 = fc.array('i', [10, 1, 30, 2]).astype('b32')
    in_2 = in_1.serialise()
    in_3 = fc.array('b32')
    in_3.deserialise(in_2)
    assert fl.verify(in_3 == in_1)