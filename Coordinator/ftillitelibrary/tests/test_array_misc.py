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

def test_missing_broadcast_mux(fc):
    # Test Case 1: Relates to Issue 57
    needs_broadcast = fc.array('i', [8])
    cond = fc.array('i', [1,0,1,0,1])

    v1 = fl.mux(cond, fc.arange(5), needs_broadcast)
    v2 = fc.array('i', [0,8,2,8,4])
    v3 = v1 == v2
    assert fl.verify(v3, False)

def test_calc_broadcast_length_spec_change(fc):
    a = fc.array('i')
    b = fc.array('i', 1, 3)
    c = a % b

    empty = fc.array('i')
    assert fl.verify(c == empty, False)