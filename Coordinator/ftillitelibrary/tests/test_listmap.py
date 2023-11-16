# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from ftillite.client import IntArrayIdentifier, ListMapIdentifier
from .util import fc
import pytest
import ftillite as fl

def test_constructor(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    k3 = fc.arange(10) + 5
    k4 = fc.array('f', 10, 1)
    k5 = fc.array('b20', 10)
    k6 = fc.array('I', 10)
    
    keys1 = [k1, k2]
    keys2 = [k1, k3, k4, k5]
    keys3 = [k1, k3, k4, k5, k6]
    
    # listmap(arrays)
    listmap_1 = fc.listmap(keys1)
    listmap_2 = fc.listmap(keys1, order="rnd")
    listmap_3 = fc.listmap([k2, k2], order="any")
    
    # listmap(typecode)
    listmap_4 = fc.listmap('if', keys1)
    
    # case: listmap(typecode, arrays)
    listmap_5 = fc.listmap('iifb20', keys2)
    
    # ed255129ints
    listmap_6 = fc.listmap('iifb20I', keys3)
    
    # Check non-keyword order params 
    listmap_7 = fc.listmap(keys1, "any")
    listmap_8 = fc.listmap('if', keys1, "rnd")
    listmap_9 = fc.listmap('if', "pos")
    
    # Check invalid order throws error
    try:
        fc.listmap('if', "foo")
        assert False
    except ValueError as ex:
        assert True
     
def test_todict(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    
    k2 = fc.arange(10) + 5
    k3 = fc.array('f', 10, 1)
    k4 = fc.array('b20', 10)
    keys2 = [k1, k2, k3, k4]
    
    listmap_1 = fc.listmap(keys1)
    assert listmap_1.todict() == {(i, 1.0): i for i in range(10)}
    
    listmap_2 = fc.listmap('iifb20', keys2)
    try:
        listmap_2.todict()
        assert False
    except TypeError as ex:
        assert True

def test_listmap_contains_keys_check(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    
    k3 = fc.arange(10) + 10
    k4 = fc.array('f', 10, 2)
    keys2 = [k3, k4]
    
    listmap_1 = fc.listmap([fc.arange(10), fc.array('f', 10, 1)])
    
    assert not listmap_1.listmap_contains_all_keys(keys2)
    assert listmap_1.listmap_contains_all_keys(keys1)
    
def test_listmap_contains_no_intersecting_keys_check(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    
    k3 = fc.arange(10) + 10
    k4 = fc.array('f', 10, 2)
    keys2 = [k3, k4]
    
    k5 = fc.arange(10) + 5
    k6 = fc.array('f', 10, 1)
    keys3 = [k5, k6]
    
    listmap_1 = fc.listmap([fc.arange(10), fc.array('f', 10, 1)])
    
    assert listmap_1.listmap_contains_no_keys(keys2)
    assert not listmap_1.listmap_contains_no_keys(keys1)
    assert not listmap_1.listmap_contains_no_keys(keys3)

def test_keys_unique_check(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    
    k3 = fc.arange(10) + 5
    k4 = fc.array('f', 10, 1)
    
    listmap_1 = fc.listmap('if', keys1)
    
    assert fc.keys_unique([k3])
    assert not fc.keys_unique([k4])
    assert fc.keys_unique(listmap_1.keys())

def test_get_keys(fc):
    k1 = fc.arange(3)
    k2 = fc.array('f', 3, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    assert verify_keys(listmap_1.keys(), [[0, 1.0], [1, 1.0], [2, 1.0]])

def test_contains(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    k3 = fc.arange(2)
    k4 = fc.array('f', 2, 1)
    
    contains_res1 = listmap_1.contains([k3, k4])
    assert isinstance(contains_res1, IntArrayIdentifier)
    assert verify_array_contents(contains_res1, [1, 1])
 
def test_getitems(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    # Test Case 1
    get_items_res1 = listmap_1[keys1]
    assert isinstance(get_items_res1, IntArrayIdentifier)
    assert verify_array_contents(get_items_res1, [0,1,2,3,4,5,6,7,8,9])
    
    # Test Case 2
    try:
        listmap_1[[k2, k2]]
        assert False
    except ValueError as ex:
        assert True
    
    # Test Case 3
    get_items_res2 = listmap_1[listmap_1.keys()]
    assert verify_array_contents(get_items_res2, [0,1,2,3,4,5,6,7,8,9])
    
    # Test Case 4
    try:
        listmap_1[listmap_1]
        assert False
    except ValueError as ex:
        assert True
     
def test_lookup(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    # Test Case 1
    k3 = fc.arange(20)
    k4 = fc.array('f', 20, 1)  
    keys2 = [k3, k4]

    lookup_res = listmap_1.lookup(keys2, -1)
    assert verify_array_contents(lookup_res, [0,1,2,3,4,5,6,7,8,9,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1])

def test_additems(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    # Test Case 1
    k3 = fc.arange(2) + 10
    k4 = fc.array('f', 2, 1)
    keys2 = [k3, k4]
    new_keys1, new_values1 = listmap_1.add_items(keys2)
    assert isinstance(new_keys1, list)
    assert isinstance(new_values1, IntArrayIdentifier)
    assert len(new_keys1) == 2
    assert len(new_keys1[0]) == len(k3)
    assert len(listmap_1) == len(k1) + len(k3)
    assert verify_keys(new_keys1, [[i + 10, 1.0] for i in range(len(k3))])
    assert verify_keys(listmap_1.keys(), [[i, 1.0] for i in range(len(listmap_1))])
    assert verify_array_contents(new_values1, [10, 11])
    
    # Test Case 2
    try:
        listmap_1.add_items(keys2)
        assert False
    except ValueError as ex:
        assert True
    
    # Test Case 3
    try:
        listmap_1.add_items([k4, k4])
        assert False
    except ValueError as ex:
        assert True
        
    # Test Case 4: Check listmap as input
    listmap_2 = fc.listmap([k1+ 20, k2])
    new_keys2, new_values2 = listmap_1.add_items(listmap_2)
    assert isinstance(new_keys2, list)
    assert isinstance(new_values2, IntArrayIdentifier)
    assert len(listmap_1) == len(k1) * 2 + len(k3)
  
def test_mergeitems(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    k3 = fc.arange(5) + 8
    k4 = fc.array('f', 5, 1)
    keys2 = [k3, k4]
    
    # Test Case 1
    new_keys1, new_values1 = listmap_1.merge_items(keys2)
    assert isinstance(new_keys1, list)
    assert isinstance(new_values1, IntArrayIdentifier)
    assert len(new_keys1) == 2
    assert len(new_keys1[0]) == 3
    assert verify_keys(new_keys1, [[i + 10, 1.0] for i in range(3)])
    assert verify_keys(listmap_1.keys(), [[i, 1.0] for i in range(13)])
    assert verify_array_contents(new_values1, [10, 11, 12])

def test_mergeitems_bytearray32(fc):
    k1 = fc.array("b32", 30)
    keys1 = [k1]
    listmap1 = fc.listmap('b32', keys1)
    k2 = fc.array("b32", 10)
    keys2 = [k2]
    new_keys1, new_values1 = listmap1.merge_items(keys2)
    assert len(new_keys1) == 1

def test_removeitems(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    k3 = fc.arange(2)
    k4 = fc.array('f', 2, 1)
    keys2 = [k3, k4]
    
    # Test Case 1
    moved_keys1, old_values1, new_values1 = listmap_1.remove_items(keys2)
    assert isinstance(moved_keys1, list)
    assert isinstance(old_values1, IntArrayIdentifier)
    assert isinstance(new_values1, IntArrayIdentifier)
    k1_len = len(k1)
    k3_len = len(k3)
    assert len(moved_keys1) == 2
    assert len(moved_keys1[0]) == k3_len
    assert len(listmap_1) == k1_len - k3_len
    assert verify_keys(moved_keys1, [[k1_len-(i+1), 1.0] for i in range(k3_len)])
    assert verify_keys(listmap_1.keys(), [[i+k3_len, 1.0] for i in range(k1_len - k3_len)])
    assert verify_array_contents(old_values1, [8, 9])
    
    # Test Case 2
    try:
        listmap_1.remove_items(keys2)
        assert False
    except ValueError as ex:
        assert True
    
    # Test Case 3
    try:
        listmap_1.remove_items([k4, k4])
        assert False
    except ValueError as ex:
        assert True
        
    # Test Case 4: Check listmap as input
    listmap_2 = fc.listmap([k1+ 20, k2])
    # TODO: add additional tests here
    moved_keys2, old_values2, new_values2 = listmap_2.remove_items(listmap_2)
    assert isinstance(moved_keys2, list)
    assert isinstance(old_values2, IntArrayIdentifier)
    assert isinstance(new_values2, IntArrayIdentifier)

def test_removeitems2(fc):
    k1 = fc.array('i', [1, 2, 3, 4])
    k2 = fc.array('f', [1.1, 2.2, 3.3, 4.4])
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    k3 = fc.array('i', [2, 3])
    k4 = fc.array('f', [2.2, 3.3])
    keys2 = [k3, k4]
    moved_keys1, old_values1, new_values1 = listmap_1.remove_items(keys2)
    assert isinstance(moved_keys1, list)
    assert isinstance(old_values1, IntArrayIdentifier)
    assert isinstance(new_values1, IntArrayIdentifier)

     
def test_discarditems(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    k3 = fc.arange(2)
    k4 = fc.array('f', 2, 1)   
    keys2 = [k3, k4]
    
    # Test Case 1: Should not error delete keys that don't exist
    listmap_1.discard_items(keys2)
    moved_keys1, old_values1, new_values1 = listmap_1.discard_items(keys2)
    assert isinstance(moved_keys1, list)
    assert isinstance(old_values1, IntArrayIdentifier)
    assert isinstance(new_values1, IntArrayIdentifier)
    
    # TODO: add test cases for function output
 
def test_intersect(fc):
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    
    k3 = fc.arange(5)
    k4 = fc.array('f', 5, 1)
    keys2 = [k3, k4]
    
    intersect_res = listmap_1.intersect_items(keys2)
    assert isinstance(intersect_res, ListMapIdentifier)
    assert len(intersect_res) == 5
    assert verify_keys(intersect_res.keys(), [[i, 1.0] for i in range(5)])
    
    # TODO: add additional test cases

def test_merge_items_empty_results(fc):
    # See issue #135
    k1 = fc.arange(5)
    listmap_1 = fc.listmap('i', [k1])
    _, v1 = listmap_1.merge_items([k1])
    v2 = fc.array('i', 5)
    v2[v1] = fc.array('i')
    assert fl.verify(v2.len() == fc.array('i', [5]))

    listmap_2 = fc.listmap('i', [k1])
    k2 = fc.array('i')
    _, v3 = listmap_2.merge_items([k2])
    v4 = fc.array('i', 5)
    v4[v3] = fc.array('i')
    assert fl.verify(v4.len() == fc.array('i', [5]))
    
def test_unflatten(fc):
    # Test case 1: check unflatten modifies in-place and runs successfully
    k1 = fc.arange(10)
    k2 = fc.array('f', 10, 1)
    keys1 = [k1, k2]
    listmap_1 = fc.listmap('if', keys1)
    listmap_2 = listmap_1.unflatten([k1 + 10, k2 + 10])
    assert listmap_1.handle() == listmap_2.handle()

    # Test Case 2: check non-list input for param 'data' raises an exception
    with pytest.raises(TypeError):
        _ = listmap_1.unflatten(k1)

def test_get_keys_single_item(fc):
    # Test case 1: Tests that a keys method returns a list for a single element composite key
    k1 = fc.array('i')
    listmap_1 = fc.listmap('i', k1)
    keys = listmap_1.keys()
    assert isinstance(keys, list)
    assert len(keys) == 1
   
# Util functions - note they only check values on coordinator
def verify_keys(keys, expected_keys):
    arrs = []
    res = []
    for i, k in enumerate(keys):
        if k.typecode() in ['i', 'f']:
            arrs.append(k.tolist())
    for i in range(len(arrs[0])):
        tmp = []
        for w in range(len(keys)):
            tmp.append(arrs[w][i])
        res.append(tmp)
    res.sort(key=lambda x: x[0])
    expected_keys.sort(key=lambda x: x[0])
    return res == expected_keys

def verify_array_contents(actual_arr, expected_list):
    fc = actual_arr.context()
    expected_arr = fc.array('i', expected_list)
    len_match = fl.verify(expected_arr.len() == actual_arr.len())
    if len_match:
        return fl.verify(actual_arr.contains(expected_arr))
    else:
        return False
    
