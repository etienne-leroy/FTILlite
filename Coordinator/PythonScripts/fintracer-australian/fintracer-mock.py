#!/usr/bin/env python
# coding: utf-8

# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################


# In[1]:


import ftillite as fl

from pair import *
from dictionary import *
from mock_elgamal import *

import logging
from datetime import datetime
from logging.handlers import RotatingFileHandler
import sys 
import time

app_name = "test_cases"
def create_logger(x, app_name):
    logger = logging.getLogger(x)
    logger.setLevel(logging.INFO)
    # formatter = logging.Formatter('%(asctime)s,%(msecs)d %(name)s %(levelname)s %(message)s')
    formatter = logging.Formatter('%(asctime)s %(levelname)s: %(message)s')
    fn = f'logs/LOG-{x}-{app_name}-{ datetime.now().strftime("%Y_%m_%d-%H:%M:%S:%f")}'
    
    file_handler = RotatingFileHandler(fn, maxBytes=200000000, backupCount=5)
    file_handler.setFormatter(formatter)
    file_handler.setLevel(logging.INFO)
    logger.addHandler(file_handler)
    
    stdout_handler = logging.StreamHandler(sys.stdout)
    stdout_handler.setLevel(logging.WARNING)
    stdout_handler.setFormatter(formatter)
    logger.addHandler(stdout_handler)
    return logger

logger_all = create_logger('ALL', app_name)
logger_client = create_logger('CLIENT', app_name)
logger_compute_mgr = create_logger('COMPUTE MGR', app_name)
logger_segment_client = create_logger('SEGMENT CLIENT', app_name)

conf = fl.FTILConf().set_app_name("nonverbose") \
                    .set_rabbitmq_conf({'user': 'ftillite', 'password': 'ftillite', 'host': 'rabbitmq.ftillite-gpu', \
                                        'AUSTRAC':'0', 'ANZ':'1', 'CBA':'2', 'NAB':'3', 'WPC':'4'})\
                    .set_client_logger(logger_client)\
                    .set_compute_manager_logger(logger_compute_mgr) \
                    .set_segment_client_logger(logger_segment_client)
                    # .set_all_loggers(logger_all)
    

fc = fl.FTILContext(conf = conf)


# In[2]:


start = time.time()

def get_branch2node(fc):
  branch2node_first = fc.array('i', 1000000, -1)
  branch2node_last = fc.array('i', 1000000, -1)
  all_accounts = Pair(fc.array('i'), fc.array('i'))
  peer_nodes = fc.scope().difference(fc.CoordinatorID)
  with fl.on(peer_nodes):
    bsbs = fc.array('i')
    bsbs.auxdb_read("SELECT DISTINCT bsb FROM accounts;")
  bsbs = fl.transmit({i : bsbs for i in fc.scope()})
  k = bsbs.keys()
  for i in k:
    branch2node_last[bsbs[i]] = i.num()
  for i in reversed(k):
    branch2node_first[bsbs[i]] = i.num()
  fl.verify(branch2node_first == branch2node_last)
  return branch2node_first

branch2node = get_branch2node(fc)


# In[3]:


def distribute_tag(tag, pub_key):
  out_scope = tag.context().scope().difference([tag.context().CoordinatorID])
  map = {}
  with fl.on(tag.scope()):
    for n in out_scope:
      # This can be done more efficiently using "discard_items".
      ind = (branch2node[tag.keys().first] == n.num()).index()
      map[n] = Dict(tag.keys()[ind], tag.values()[ind])
  # The next loop performs refreshing, prior to transmitting.
  for n in out_scope:
    with fl.on(tag.scope().difference(set([n]))):
      elgamal_refresh(map[n].values(), pub_key)
  return fl.transmit(map)



def fintracer_step(tag, txs, pub_key):
  fc = fl.get_context([tag, txs, pub_key])
  with fl.on(tag.scope()):
    zero = fc.array('i', 1)
    nonce = elgamal_encrypt(zero, pub_key)
    next_tag = tag.stub()
    with fl.massop():
      next_tag.reduce_sum(txs.second, tag.lookup(txs.first, nonce))
  map = distribute_tag(next_tag, pub_key)
  scope = set()
  for n in map:
    scope.update(map[n].scope())
  with fl.on(scope):
    # We need to create here a new empty tag, but can't use ".stub()" because
    # it may have a different scope than any existing tag.
    # In a real implementation, we would have had a "tag" type, and all this
    # would have happened automatically.
    accounts_stub = Pair(fc.array('i'), fc.array('i'))
    cipher_stub = ElGamalCipher(fc)
    rc = Dict(accounts_stub, cipher_stub)
    for n in map:
      with fl.on(map[n].scope()):
        rc += map[n]
  return rc


# In[4]:


epsilon = 0.001
delta = 0.001

from math import exp, log, ceil, floor

def diff_priv_amount(epsilon, delta):
    gamma = 1 - exp(-epsilon)
    Y = max(0, ceil(log(gamma*(gamma-delta)/(delta*(1-exp(-2*epsilon))) + 1) / epsilon))
    t = 1 + (delta - 1) * exp(-epsilon) - delta * exp((Y - 1)*epsilon)
    r = fc.randomarray('f', 1, 1-gamma/t, 1)
    x = Y + floor(fl.mux(r > 0, -fl.log(r)/epsilon, fl.log(1+ r*t/(delta*exp((Y-1)*epsilon)))/epsilon))
    return x


# In[5]:


with fl.on(fc.CoordinatorID): # On the coordinator node...

  # Creating an ElGamal key pair.
  (priv_key, local_pub_key) = elgamal_keygen(fc)

# Back on all nodes: distributing the public key across all nodes.
pub_key = fl.transmit({i : local_pub_key for i in fc.scope()})[fc.CoordinatorID]

# On the peer nodes...
peer_nodes = fc.scope().difference(fc.CoordinatorID)


# In[6]:


# plain_zero = fc.array('E', 1)
# zero = elgamal_encrypt(plain_zero, pub_key)

zero = ElGamalCipher(fc)
zero.set_length(1)

with fl.on(fc.CoordinatorID):
    local_plain_zero = zero.decrypt(priv_key)
    
plain_zero = fl.transmit({n : local_plain_zero for n in fc.scope()})[fc.CoordinatorID]

def read_tag(tag, targets, priv_key, pub_key):
    with fl.on(peer_nodes):
        target_values = tag.lookup(targets, zero)
        elgamal_sanitise(target_values)
        x = diff_priv_amount(epsilon, delta)
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
    return {n.name() : list(v) for n,v in rc_local.items()}


# In[ ]:


accounts = Pair(fc.array('i'), fc.array('i'))

transactions = Pair(accounts, accounts)

with fl.on(peer_nodes):
  transactions.auxdb_read("SELECT DISTINCT origin_bsb, origin_id, dest_bsb, dest_id FROM transactions;")

# def double_size(x):
#     xl = x.len()
#     x.set_length(2*xl)
#     x[xl:] = x[:xl]

# with fl.on(peer_nodes):
#     transactions.auxdb_read("SELECT DISTINCT origin_bsb, origin_id, dest_bsb, dest_id FROM transactions;")
#     double_size(transactions)
#     double_size(transactions)
#     double_size(transactions)
#     double_size(transactions)
#     double_size(transactions)
#     double_size(transactions)
    

with fl.on(peer_nodes):
  # accounts.auxdb_read("SELECT DISTINCT bsb, account_id FROM accounts LIMIT 2;")
  # sources = accounts[0]
  # targets = accounts[1]
  sources = Pair(fc.array('i'), fc.array('i'))
  sources.auxdb_read("""
SELECT bsb, account_id FROM (
SELECT ROW_NUMBER () OVER (
  ORDER BY bsb, account_id
) RowNum,
bsb, account_id FROM (SELECT DISTINCT bsb, account_id FROM accounts) s
) t WHERE RowNum <= 3;
""")
  targets = Pair(fc.array('i'), fc.array('i'))
  targets.auxdb_read("""
SELECT bsb, account_id FROM (
SELECT ROW_NUMBER () OVER (
  ORDER BY bsb, account_id
) RowNum,
bsb, account_id FROM (SELECT DISTINCT bsb, account_id FROM accounts) s
) t WHERE RowNum > 3 AND RowNum <= 6;
""")
  
  values = fc.array('i', sources.len(), 1)
  ciphertext = elgamal_encrypt(values, pub_key)
  dist_tag = Dict(sources, ciphertext)


# In[8]:


print("Step start.")
step_start = time.time()
dist_tag = fintracer_step(dist_tag, transactions, pub_key)
step_end = time.time()
print("Step end.")

print(f"Time for FinTracer step: {step_end - step_start}")


# In[9]:


print("Read start")
read_start = time.time()
rc = read_tag(dist_tag, targets, priv_key, pub_key)
read_end = time.time()
print("Read end")

print(f"Return value: {rc}")

print(f"Read time: {read_end - read_start}")


# In[10]:


print("Step start.")
step_start = time.time()
dist_tag = fintracer_step(dist_tag, transactions, pub_key)
step_end = time.time()
print("Step end.")

print(f"Time for FinTracer step: {step_end - step_start}")


# In[11]:


print("Read start")
read_start = time.time()
rc = read_tag(dist_tag, targets, priv_key, pub_key)
read_end = time.time()
print("Read end")

print(f"Return value: {rc}")

print(f"Read time: {read_end - read_start}")


# In[12]:


print("Done.")
end = time.time()
print(f"Time for the full computation: {end - start}")

