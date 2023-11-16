# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from ftillite.client import ArrayIdentifier
import pytest
import ftillite as fl
import os

@pytest.fixture(scope = 'module')
def fc():
    mq_username = os.environ.get("MQ_USER", "default_mq_user")
    mq_password = os.environ.get("MQ_PW", "default_mq_pw")
    conf = fl.FTILConf().set_app_name("nonverbose") \
                    .set_rabbitmq_conf({'user': mq_username, 
                                        'password': mq_password, 
                                        'host': 'localhost',
                                        'del_await_resp': True}) \
                    .set_peer_names(['COORDINATOR', 'PEER_1', 'PEER_2', 'PEER_3', 'PEER_4'])
    return fl.FTILContext(conf = conf)

def check_array_length(arr, expected_len):
    v1 = arr.context().array('i', 1, expected_len)
    v2 = v1 == arr.len()
    return fl.verify(v2, raise_exception=False)

def check_array_value(actual_vals, expected_vals):
    if not isinstance(expected_vals, fl.ArrayIdentifier):
        fc = actual_vals.context()
        tc = actual_vals.typecode()
        expected_vals = fc.array(tc, expected_vals)
    res = actual_vals == expected_vals
    return fl.verify(res, raise_exception=False)

def check_array(arr, expected_type, expected_vals):
    type_check = True
    len_check = True
    val_check = True
    fc = arr.context()
    
    type_check = arr.typecode() == expected_type
    len_check = check_array_length(arr, len(expected_vals))
    if expected_type not in ['i', 'f'] and not isinstance(expected_vals, ArrayIdentifier):
        expected_vals = fc.promote(expected_vals)[0]
        expected_vals = expected_vals.astype(expected_type) # Handle I, bX and E types
    val_check = check_array_value(arr, expected_vals)
    
    return type_check and len_check and val_check

def check_array_for_types(fc, input_arrays, output_array, operation, types_supported = ['i', 'f', 'b8', 'b32', 'I', 'E']):
    test_results = [True] * len(types_supported)
    for idx, tc in enumerate(types_supported):
        if tc == 'E':
            if fc._backend.gpuenabled:
                input_arrays, output_array = process_arrays_for_ed25519(input_arrays, output_array)
            else:
                print("skipping GPU test")
                continue
        try:
            in_arrs = []
            for i in input_arrays:
                if not isinstance(i, ArrayIdentifier):
                    in_arr = fc.array('i', i)
                    in_arr = in_arr.astype(tc) if tc != 'i' else in_arr
                else:
                    in_arr = i
                in_arrs.append(in_arr)
            expected_out_arr = fc.array('i', output_array)
            expected_out_arr = expected_out_arr.astype(tc) if tc != 'i' else expected_out_arr
            actual_out_arr = operation(*in_arrs)
            equals_comparison = actual_out_arr == expected_out_arr
            res = fl.verify(equals_comparison, raise_exception=False)
            test_results[idx] = res
        except Exception as ex:
            raise Exception(f"Type {tc} failed with error: {ex}")
    return test_results
        
def process_arrays_for_ed25519(in_arrs, out_arr):
    # Handles cases of negative values, to avoid ED25519 operations failing
    in_count = len(in_arrs)
    new_in_arrs = [[] for _ in range(in_count)]
    new_out_arr = []
    for idx, o in enumerate(out_arr):
        if o >= 0 and all([x[idx] >= 0 for x in in_arrs]):
            for x in range(in_count):
                new_in_arrs[x].append(in_arrs[x][idx])
            new_out_arr.append(o)
    return new_in_arrs, new_out_arr
        
    
    
        