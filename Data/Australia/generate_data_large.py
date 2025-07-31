# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

import numpy as np
from time import time
from datetime import datetime
import os
import pandas as pd
import configparser
import sys
import subprocess

def get_config_items_list(config, section, key):
    """Function to split comma delimeted string returned from
    config.get() into a list of elements.
    """
    itms = config[section].get(key).split(',')
    itms_list = [item.strip() for item in itms]
    return itms_list

if len(sys.argv) == 2:
    config_file = sys.argv[1]
else:
    print("\nError: You must include a path to a config file. \n\n  -- Usage: " + sys.argv[0] + " <path/to/config/file>.\n\n")
    sys.exit()

if not os.path.isfile(config_file):
    print("\n\n Error: The config file '" + config_file + "' does not exist. Exiting program.\n\n")
    sys.exit()

# Read parameters from Config File...

config = configparser.ConfigParser()

config.read(config_file)

node_count = config['Graph'].getint('num_nodes', fallback=300)
edge_count = config['Graph'].getint('num_edges', fallback=600)
peer_names = get_config_items_list(config, 'Graph', 'peer_names')
peer_count = len(peer_names)

# Account config
acc_num = config['Graph'].getint('account_num_len', fallback=15)

# Transaction config
start_date_str = config['TransactionDatetime']['start_date_str']
date_range = config['TransactionDatetime'].getint('date_range', fallback=10)
start_time = config['TransactionDatetime'].getint('start_time', fallback=32400)
end_time = config['TransactionDatetime'].getint('end_time', fallback=61200)

avg_tx_amount =config['TransactionAmount'].getfloat('avg_tx_amount', fallback=200.0)
sigma_tx_amount = config['TransactionAmount'].getfloat('sigma_tx_amount', fallback=4)

seed = config['Graph'].getint('seed', fallback=123)
np.random.seed(seed)

chunksize = config['Graph'].getint('chunk_size', fallback=10 ** 5)
num_bsbs = get_config_items_list(config, 'Graph', 'num_bsbs')
num_bsbs_per_peer = {}
if len(num_bsbs) == 1:
    num_bsbs_per_peer = {p: int(num_bsbs[0]) for p in peer_names}
elif len(num_bsbs) == peer_count:
    num_bsbs_per_peer = {p: int(num_bsbs[i]) for i, p in enumerate(peer_names)}
else:
    raise ValueError("num_bsbs should have a length equal to the number of peers or a single value")

# Create directory for output data
# (Note: this assumes the directory synthesised_data exists.)
time_now = datetime.today().strftime('%Y-%m-%d-%H-%M-%S')

if not os.path.exists("synthesised_data"):
    os.mkdir("synthesised_data")

dir_name = "synthesised_data/transactions_" + time_now

if not os.path.exists(dir_name):
    os.mkdir(dir_name)

# Start timer
start = time()


# Generate transactions - before RAM becomes over utilised
a_val = config['Graph'].get('a', fallback="0.57")
b_val = config['Graph'].get('b', fallback="0.19")
c_val = config['Graph'].get('c', fallback="0.19")
txt_location = f"{dir_name}/edges.csv"

subprocess.run(["./PaRMAT", 
                "-nVertices", str(node_count), 
                "-nEdges", str(edge_count), 
                "-threads", "4", 
                "-noEdgeToSelf", 
                "-noDuplicateEdges",
                "-seed", str(seed),
                "-memUsage", "0.9", 
                "-output", f"{dir_name}/edges.csv", 
                "-a", a_val, "-b", b_val, "-c", c_val])

acc_ids_count = 0
acc_ids = np.array([])

# Generate Account Ids
while acc_ids_count != node_count:
    count_dff = node_count - acc_ids_count
    tmp_acc_ids = np.random.randint(10**(acc_num-1), (10**(acc_num) - 1), size=count_dff)
    acc_ids = np.unique(np.concatenate((acc_ids, tmp_acc_ids), axis=0))
    acc_ids_count = len(acc_ids)
# node_count = len(acc_ids)

print(f"Accounts Generated: {node_count}")

# Generate RE ids
re_ids = np.random.randint(0, high=peer_count, size=node_count)
# Calculate distribution of RE ids
re_unique, re_counts = np.unique(re_ids, return_counts=True)
print(f"RE Distribution: {re_counts}")

# Generate BSBs
bsb_ids = np.empty(node_count).astype(int)
bsb_map = {}
bsb_num = config['Graph'].getint('bsb_num_len', fallback=6)
gen_count = bsb_num - len(str(peer_count))

for i, p in enumerate(peer_names):
    prefix = (i+1) * 10**(gen_count)
    bsb_map[i] = prefix + np.random.randint(
    10**(gen_count-1), high=(10**(gen_count) - 1), size=num_bsbs_per_peer[p])

for i in range(0, peer_count):
    itemindex = np.where(re_ids == i)[0]
    rand_bsbs = np.random.choice(bsb_map[i], re_counts[i])
    bsb_ids[itemindex] = rand_bsbs

print("BSBs Generated")

# Collate all account data
all_accounts= np.column_stack((re_ids, acc_ids, bsb_ids))
print("All Account Data Generated")

# Generate Transaction amounts
tx_amounts = ((np.random.lognormal(
    mean=np.log(avg_tx_amount), sigma=np.log(sigma_tx_amount), size=edge_count) + 1)).astype(int)
# We add 1 to avoid transactions with $0
print("Transaction Amounts Generated")

# Generate Transaction times
start_date = np.datetime64(start_date_str)
date_deltas = np.random.randint(0, date_range, edge_count)
time_deltas = np.random.randint(
    start_time, end_time, edge_count).astype('timedelta64[s]')
tx_datetimes = (np.busday_offset(
    start_date, date_deltas, roll='forward') + time_deltas).astype(int)   
print("Transaction Timings Generated")

transaction_file_loc = lambda x: f'{dir_name}/tmp_transactions_{x}.csv'

init = True
chunk_counter = 0
with pd.read_csv(txt_location, chunksize=chunksize, dtype=(np.int32, np.int32), header=None) as reader:
    rows_processed = 0
    for idx, chunk in enumerate(reader):
        transaction_df = chunk.to_numpy()
        
        row_count = chunk.shape[0]
        
        transaction_to_df = np.column_stack((all_accounts[transaction_df[:,0]], all_accounts[transaction_df[:,1]], tx_amounts[rows_processed:row_count+rows_processed], tx_datetimes[rows_processed:row_count+rows_processed]))
        np.savetxt(transaction_file_loc(chunk_counter), transaction_to_df, fmt="%d", delimiter=',', header='origin_re_tmp,origin_id,origin_bsb,dest_re_tmp,dest_id,dest_bsb,amount,datetime', comments='')
        del transaction_to_df
        
        chunk_counter += 1
        rows_processed += row_count
        
        print(f"Transaction Chunk {idx+1} Done")

print(f"Transaction Written To Disk")

# Write accounts to file and clear
account_file_loc = f'{dir_name}/tmp_all_accounts.csv'
np.savetxt(account_file_loc, all_accounts, fmt="%d", delimiter=',', header='re_tmp,account_id,bsb', comments='')
del all_accounts

print(f"Accounts Written To Disk")

# Replace RE id with RE string
def rename_in_chunks(in_fn, out_fn, old_cols, new_cols, out_col_order, is_transactions=False):
    init = True
    with pd.read_csv(in_fn, chunksize=chunksize) as reader:
        for idx, chunk in enumerate(reader):
            for i, old_c in enumerate(old_cols):
                new_c = new_cols[i]
                for idx, r in enumerate(peer_names):
                    chunk.loc[chunk[old_c] == idx, new_c] = r
            df = chunk.drop(old_cols, axis=1)

            if is_transactions:
                df.rename(columns={'amount': 'total_amount', 'datetime': 'transaction_date'}, inplace=True)
                df['transaction_date'] = df['transaction_date'].astype('datetime64[s]').dt.date
            
            df.to_csv(out_fn, mode='a', header=init, index=False, columns=out_col_order)

            # Make a CSV copy of each RE's data
            for r in peer_names:
                query = " or ".join([f'{i} == "{r}"' for i in new_cols])
                re_filename = out_fn.replace('.', f'_{r}_re_copy.')
                df.query(query).to_csv(re_filename, mode='a', header=init, index=False, columns=out_col_order)
            
            init = False
            print(f"Renaming Chunk {idx+1} Done")
        

rename_in_chunks(account_file_loc, account_file_loc.replace('tmp_', ''), ['re_tmp'], ['re'], ['re', 'account_id', 'bsb'])

for i in range(0, chunk_counter):
    transaction_file_loc = account_file_loc = f'{dir_name}/tmp_transactions_{i}.csv'
    rename_in_chunks(transaction_file_loc, transaction_file_loc.replace('tmp_', ''), 
                        ['origin_re_tmp', 'dest_re_tmp'], 
                        ['origin_re', 'dest_re'], 
                        ['origin_re', 'origin_id', 'origin_bsb', 'dest_re', 'dest_id', 'dest_bsb', 'total_amount', 'transaction_date'], 
                        True)

    print(f"Renaming Chunk {i+1} of {chunk_counter} Done")
end = time()
print(f"Time for the full computation: {end - start}")