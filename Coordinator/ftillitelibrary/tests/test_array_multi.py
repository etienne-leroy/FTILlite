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

def test_ed25516_mul_ed25519Int(fc):
    # Test Case 1: Relates to Issue 66
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    v1 = fc.array("i", [1, 5, 7])
    v2 = v1.astype("E")
    v3 = v1.astype("I")
    v4 = v2 * v3
    v5 = fc.array("i", [1, 25, 49]).astype("E")
    v6 = v4 == v5
    assert fl.verify(v6, False)

def test_ed25516Int_mul_ed25519(fc):
    # Test Case 1: Relates to Issue 66
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    v1 = fc.array("i", [1, 5, 7])
    v2 = v1.astype("E")
    v3 = v1.astype("I")
    v4 = v3 * v2
    v5 = fc.array("i", [1, 25, 49]).astype("E")
    v6 = v4 == v5
    assert fl.verify(v6, False)

def test_ed25516Int_mul_ed25519Int(fc):
    # Test Case 1: Relates to Issue 66
    v1 = fc.array("i", [1, 5, 7])
    v2 = v1.astype("I")
    v3 = v1.astype("I")
    v4 = v2 * v3
    v5 = fc.array("i", [1, 25, 49]).astype("I")
    v6 = v4 == v5
    assert fl.verify(v6, False)

def test_ed25516_mul_int(fc):
    if not fc._backend.gpuenabled:
        pytest.skip("GPU not enabled")

    # Test Case 1: i * E - Relates to Issue 137
    v1 = fc.array("i", [1, 5, 7])
    v2 = v1.astype("E")
    v3 = v1 * v2
    v4 = fc.array("i", [1, 25, 49]).astype("E")
    v5 = v3 == v4

    # Test Case 2: E * i - Relates to Issue 137
    v3 = v2 * v1
    v4 = fc.array("i", [1, 25, 49]).astype("E")
    v5 = v3 == v4

    assert fl.verify(v5, False)