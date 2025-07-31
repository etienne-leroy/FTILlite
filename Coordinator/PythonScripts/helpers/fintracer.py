"""
Core FinTracer Algorithm Functions
"""

import os
import sys
import ftillite as fl
from math import exp, log, ceil, floor
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../TypeDefinitions'))

from pair import Pair                                                                   # type: ignore
from dictionary import Dict                                                             # type: ignore
from elgamal import elgamal_encrypt, elgamal_refresh, elgamal_sanitise, ElGamalCipher   # type: ignore
                                                                   

def create_initial_tag(account_set, pub_key, fc):
    """Create initial encrypted tag for an account set"""
    peer_nodes = fc.scope().difference(fc.CoordinatorID)
    with fl.on(peer_nodes):
        values = fc.array('i', account_set.len(), 1)
        ciphertext = elgamal_encrypt(values, pub_key)
        tag = Dict(account_set, ciphertext)
    return tag

def create_initial_tags(subsets, public_key, fc):
        all_A_subset_tags = []
        for sub in subsets:
            tag_Ai  = create_initial_tag(sub,  public_key, fc)
            all_A_subset_tags.append(tag_Ai)
        return all_A_subset_tags   

def distribute_tag(tag, pub_key, branch2node):
    """Distribute tags across nodes"""
    out_scope = tag.context().scope().difference([tag.context().CoordinatorID])
    map = {}
    with fl.on(tag.scope()):
        for n in out_scope:
            ind = (branch2node[tag.keys().first] == n.num()).index()
            map[n] = Dict(tag.keys()[ind], tag.values()[ind])

    for n in out_scope:
        with fl.on(tag.scope().difference(set([n]))):
            elgamal_refresh(map[n].values(), pub_key)

    return fl.transmit(map)

def fintracer_step(tag, txs, pub_key, branch2node):
    """Execute one step of FinTracer propagation"""
    fc = fl.get_context([tag, txs, pub_key])
    with fl.on(tag.scope()):
        zero = fc.array('i', 1)
        nonce = elgamal_encrypt(zero, pub_key)
        next_tag = tag.stub()

        with fl.massop():
            next_tag.reduce_sum(txs.second, tag.lookup(txs.first, nonce))

    map = distribute_tag(next_tag, pub_key, branch2node)
    scope = set()
    for n in map:
        scope.update(map[n].scope())

    with fl.on(scope):
        accounts_stub = Pair(fc.array('i'), fc.array('i'))
        cipher_stub = ElGamalCipher(fc)
        rc = Dict(accounts_stub, cipher_stub)

        for n in map:
            with fl.on(map[n].scope()):
                rc += map[n]
    return rc

def propagate_all_tags(all_A_subset_tags, public_key, transactions, bank_id_mapping):
            all_B_tags = []
            for tag_ in all_A_subset_tags:
                tag_Bi = fintracer_step(tag_, transactions, public_key, bank_id_mapping)
                all_B_tags.append(tag_Bi)
            return all_B_tags

def diff_priv_amount(fc, epsilon, delta):
    """Generate differential privacy noise amount"""
    gamma = 1 - exp(-epsilon)
    Y = max(0, ceil(log(gamma*(gamma-delta)/(delta*(1-exp(-2*epsilon))) + 1) / epsilon))
    t = 1 + (delta - 1) * exp(-epsilon) - delta * exp((Y - 1)*epsilon)
    r = fc.randomarray('f', 1, 1-gamma/t, 1)
    x = Y + floor(fl.mux(r > 0, -fl.log(r)/epsilon, fl.log(1+ r*t/(delta*exp((Y-1)*epsilon)))/epsilon))
    return x

def read_tag(tag, targets, priv_key, pub_key, zero, plain_zero, fc, epsilon=0.001, delta=0.001):
    """Read and decrypt tag results with differential privacy"""
    peer_nodes = fc.scope().difference(fc.CoordinatorID)
    with fl.on(peer_nodes):
        target_values = tag.lookup(targets, zero)
        elgamal_sanitise(target_values)
        x = diff_priv_amount(fc, epsilon, delta)
        target_values.set_length(target_values.len() + x)
        elgamal_refresh(target_values, pub_key)
        perm = fc.randomperm(target_values.len())
        v = target_values[perm]
        
    v_local = fl.transmit({fc.CoordinatorID : v})
    with fl.on(fc.CoordinatorID):
        d_local = {n : x.decrypt(priv_key) != plain_zero for n,x in v_local.items()}
            
    d = fl.transmit(d_local)[fc.CoordinatorID]
    with fl.on(peer_nodes):
        rc = targets[perm[d.index()]]
        
    rc_local = fl.transmit({fc.CoordinatorID : rc})
    result = {n.name() : list(v) for n,v in rc_local.items()}
    
    return result
