"""
Data Loading Helper Functions for FinTracer
"""

import ftillite as fl
import os
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../TypeDefinitions'))
from pair import Pair # type: ignore

def get_branch2node(fc):
    """Setup branch-to-node mapping"""
    branch2node_first = fc.array('i', 1000000, -1)
    branch2node_last = fc.array('i', 1000000, -1)
    #all_accounts = Pair(fc.array('i'), fc.array('i'))
    peer_nodes = fc.scope().difference(fc.CoordinatorID)
    with fl.on(peer_nodes):
        bsbs = fc.array('i')
        bsbs.auxdb_read("SELECT DISTINCT bank_id FROM accounts;")
    bsbs = fl.transmit({i : bsbs for i in fc.scope()})
    k = bsbs.keys()
    for i in k:
        branch2node_last[bsbs[i]] = i.num()
    for i in reversed(k):
        branch2node_first[bsbs[i]] = i.num()
    fl.verify(branch2node_first == branch2node_last)
    return branch2node_first

def load_transactions(fc):
    """Load transaction data from databases"""
    peer_nodes = fc.scope().difference(fc.CoordinatorID)
    accounts = Pair(fc.array('i'), fc.array('i'))
    transactions = Pair(accounts, accounts)
    
    with fl.on(peer_nodes):
        transactions.auxdb_read("SELECT DISTINCT origin_bank_id, origin_id, dest_bank_id, dest_id FROM transactions;")
    
    return transactions
