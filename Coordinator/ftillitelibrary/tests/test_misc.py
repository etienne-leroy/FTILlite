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

def test_log_message(fc):
    fc.log_message("Hello")

def test_log_message_with_spaces(fc):
    fc.log_message("Test spaces")

def test_log_value(fc):
    v1 = fc.array("i", [1, 2, 3, 4, 5])
    v1.log_value()
    v1.log_value("Test")

# Test for issue 141
def test_log_value_withspaces(fc):
    v1 = fc.array("i", [1, 2, 3, 4, 5])
    v1.log_value()
    v1.log_value("Test with spaces")

def test_log_stats(fc):
    fc.log_stats()
    fc.log_stats(True)
    fc.log_stats(False)