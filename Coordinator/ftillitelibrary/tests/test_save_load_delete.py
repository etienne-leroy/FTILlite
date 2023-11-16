# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from ftillite.client import NodeSet
from tests.test_listmap import verify_keys
from .util import fc
import ftillite as fl
import pytest

def test_save_load_intarray(fc):
    v1 = fc.array("i", [1, 2, 3, 4, 5])
    destination = "pytesting"
    fc.save(v1, destination)
    v2 = fc.load(destination)
    v3 = v1 == v2
    assert fl.verify(v3)

def test_save_load_floatarray(fc):
    v1 = fc.array("f", [1.1, 2.2, 3.3, 4.4, 5.5])
    destination = "pytesting"
    fc.save(v1, destination)
    v2 = fc.load(destination)
    v3 = v1 == v2
    assert fl.verify(v3)

def test_save_load_bytearrayarray(fc):
    i1 = fc.array("i", [1, 2, 3, 4, 5])
    b1 = i1.astype("b8")
    destination = "pytesting"
    fc.save(b1, destination)
    i2 = fc.load(destination)
    i3 = i1 == i2
    assert fl.verify(i3)

def test_save_load_listmap(fc):
    k1 = fc.arange(10)
    k2 = fc.arange(10) + 5
    k3 = fc.array('f', 10, 1)
    k4 = fc.array('b20', 10)
    k5 = fc.array('I', 10)
       
    keys1 = [k1, k2, k3, k4, k5]
        
    # ed255129ints
    listmap_1 = fc.listmap('iifb20I', keys1)
    
    destination = "pytesting"
    fc.save(listmap_1, destination)
    listmap_2 = fc.load(destination)

    assert listmap_2.listmap_contains_all_keys(keys1)
    
def test_save_load_properties(fc):
    v1 = fc.array("i", [1, 2, 3, 4, 5])
    v1.k = 1
    destination = "pytesting"
    fc.save(v1, destination)
    v2 = fc.load(destination)
    v3 = v1 == v2
    assert hasattr(v2, 'k')
    assert v2.k == 1
    assert fl.verify(v3)    
    assert isinstance(v2._scope, NodeSet)
    assert fl.verify((v2 + v1) == (2 * v1))
      
def test_save_load_empty_array(fc):
    v1 = fc.array("i", 0)
    destination = "pytesting"
    fc.save(v1, destination)
    v2 = fc.load(destination)
    v3 = v1 == v2
    assert fl.verify(v3)
    
def test_save_load_empty_listmap(fc):
    k1 = fc.array('i', 0)
    k2 = fc.array('f', 0)
    k3 = fc.array('b20', 0)
    k4 = fc.array('I', 0)
       
    keys1 = [k1, k2, k3, k4]
        
    # ed255129ints
    listmap_1 = fc.listmap('ifb20I', keys1)
    
    destination = "pytesting"
    fc.save(listmap_1, destination)
    listmap_2 = fc.load(destination)

    assert listmap_2.listmap_contains_all_keys(keys1)
      