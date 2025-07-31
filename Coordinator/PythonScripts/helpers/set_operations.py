"""
Set Operations for FinTracer
"""

import ftillite as fl
import os
import sys
from math import ceil, floor, exp
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../TypeDefinitions'))
from dictionary import Dict # type: ignore
from .fintracer import diff_priv_amount
from elgamal import elgamal_encrypt, elgamal_refresh, elgamal_sanitise, ElGamalCipher # type: ignore

def noise_amount(fc, epsilon, delta, m):
    """Calculate noise amount for differential privacy"""
    k = max(0, ceil(m/delta - 1/(1-exp(-epsilon/m))))
    if min(k, m-1) > 0:
        rho = floor(fc.randomarray('f', 1, 0, 1) * m / delta / min(k, m-1))
        rand1 = fc.randomarray('i', 1, 0, min(k, m-1) -1)
    else:
        rho = fc.array('i', 1)
        rand1 = fc.array('i', 1)
    if k < m:
        rand2 = k + fl.floor(-fl.log(fc.randomarray('f', 1, 0, 1))*m/epsilon)
    else:
        rand2 = m-1 + diff_priv_amount(fc, epsilon/m, delta/(m-(m-1)*delta))
    return fl.mux(rho == 1, rand1, rand2)


def calc_multi_negation(arr, pub_key, priv_key, zero, plain_zero, fc, epsilon=0.001, delta=0.001):
    """Calculate multi-negation for set operations"""
    
    peer_nodes = fc.scope().difference(fc.CoordinatorID)
    with fl.on(peer_nodes):
        zeros = fc.array('i', 1+noise_amount(fc, epsilon, delta, len(arr)))
        ones = fc.array('i', 1+noise_amount(fc, epsilon, delta, len(arr)), 1)
        padding = zeros.len() + ones.len()
        zeros = elgamal_encrypt(zeros, pub_key)
        ones = elgamal_encrypt(ones, pub_key)
        datavec = [d.values() for d in arr]
        datavec.extend([zeros, ones])
        arrlen = len(datavec)
        datalen = fc.array('i', 1)
        for d in datavec:
            datalen += d.len()
        dataitems = ElGamalCipher(fc.array('E', datalen), fc.array('E', datalen))
        datalen[0] = 0
        for d in datavec:
            newlen = datalen + d.len()
            dataitems[datalen[0]:newlen[0]] = d
            datalen = newlen
        elgamal_sanitise(dataitems)
        elgamal_refresh(dataitems, pub_key)
        perm = fc.randomperm(datalen)
        dataitems = dataitems[perm]
    valuesmap = fl.transmit({fc.CoordinatorID : dataitems})
    with fl.on(fc.CoordinatorID):
        dec = {n : valuesmap[n].decrypt(priv_key) for n in valuesmap}
        del valuesmap
        one_plaintext_array = fc.array('i', 1, 1)
        one = elgamal_encrypt(one_plaintext_array, pub_key)
        neg = {}
        nodes = list(dec.keys())
        for n in nodes:
            neg[n] = fl.mux(dec[n] == plain_zero, one, zero)
            del dec[n]
            elgamal_refresh(neg[n], pub_key)
    valuesmap = fl.transmit(neg)
    with fl.on(fc.CoordinatorID):
        del neg
    narr = []
    with fl.on(peer_nodes):
        dataitems[perm] = valuesmap[fc.CoordinatorID]
        datalen[0] = 0
        for d in arr:
            newlen = datalen + d.len()
            narr.append(Dict(d.keys(), dataitems[datalen[0]:newlen[0]]))
            datalen = newlen
    return narr

def calc_negation(a, pub_key, priv_key, zero, plain_zero, fc):
    """Calculate negation of a set"""
    return calc_multi_negation([a], pub_key, priv_key, zero, plain_zero, fc)[0]

def calc_union(a, b, fc):
    """Calculate union of two sets"""
    peer_nodes = fc.scope().difference(fc.CoordinatorID)
    with fl.on(peer_nodes):
        c = a.stub()
        c += a
        c += b
    return c

def calc_multi_union(tags, fc):
    """Calculate union of multiple sets"""
    if not tags:
        raise ValueError("Tags list cannot be empty")
    
    result = tags[0]
    for tag in tags[1:]:
        result = calc_union(result, tag, fc)
    return result

def calc_intersection(a, b, pub_key, priv_key, zero, plain_zero, fc):
    """Calculate intersection of two sets"""
    narr = calc_multi_negation([a,b], pub_key, priv_key, zero, plain_zero, fc)
    nc = calc_union(narr[0], narr[1], fc)
    return calc_negation(nc, pub_key, priv_key, zero, plain_zero, fc)

def calc_multi_normalize(arr, pub_key, priv_key, zero, plain_zero, fc):
    """Normalize multiple sets"""
    peer_nodes = fc.scope().difference(fc.CoordinatorID)
    # Negation of tags B1 to B12
    all_neg_B = calc_multi_negation(arr, pub_key, priv_key, zero, plain_zero, fc)

    # Compute 1 - NOT(Bi) for each subset
    with fl.on(peer_nodes):
        ones_encrypted = elgamal_encrypt(fc.array('i', all_neg_B[0].len(), 1), pub_key)
        ones_dict = Dict(all_neg_B[0].keys(), ones_encrypted)

        # Create tags by computing 1 - NOT(Bi) = ones - neg_Bi
        tags = []
        for neg_B in all_neg_B:
            tag = ones_dict.stub()
            tag += ones_dict
            tag -= neg_B
            tags.append(tag)
        return tags
    
def restrict_tag_with_set(tag, account_set, default_zero, peer_nodes):
    with fl.on(peer_nodes):
        cipher_AB = tag.lookup(account_set, default_zero)
        return Dict(account_set, cipher_AB)


def calc_multi_intersection(sets, pub_key, priv_key, zero, plain_zero, fc):
    """Calculate intersection of multiple sets"""
    if not sets:
        raise ValueError("Sets list cannot be empty")
    
    narr = calc_multi_negation(sets, pub_key, priv_key, zero, plain_zero, fc)
    nc = calc_multi_union(narr, fc)
    
    return calc_negation(nc, pub_key, priv_key, zero, plain_zero, fc)

def subtract_low_hits(B_prime_tags, public_key, peer_nodes, minimum_num_hits_k, fc):
        with fl.on(peer_nodes):

            ones_encrypted = elgamal_encrypt(fc.array('i', B_prime_tags.len(), 1), public_key)
            ones_dict      = Dict(B_prime_tags.keys(), ones_encrypted)
                
            B_prime_tags_temp = []
            for i in range(minimum_num_hits_k):
                # Create a fresh Dict for each iteration
                B_temp_i = B_prime_tags.stub()
                B_temp_i += B_prime_tags

                # For i > 0, subtract ones_dict i times using a different approach
                for _ in range(i):
                    # Create a multiplier for the ones
                    B_temp_i -= ones_dict
                    
                B_prime_tags_temp.append(B_temp_i)
            return B_prime_tags_temp

def compute_final_B(B_prime_tags_temp, public_key, private_key, encrypted_zero, plain_zero, fc):
            # Apply calc_negation to all B_temps using a loop
            B_temp_negs = calc_multi_negation(B_prime_tags_temp, public_key, private_key, encrypted_zero, plain_zero, fc)

            # Perform unions using a loop - fix USE CALC MULTI UNION
            B_sum = B_temp_negs[0]
            for B_temp_neg in B_temp_negs[1:]:
                B_sum = calc_union(B_sum, B_temp_neg, fc)
            
            return calc_negation(B_sum, public_key, private_key, encrypted_zero, plain_zero, fc)




