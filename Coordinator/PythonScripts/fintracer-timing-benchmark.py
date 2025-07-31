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

'''
FinTracer Timing Benchmark Script
Measures performance of key operations without processing results.
'''
import ftillite as fl

from Coordinator.PythonScripts.helpers.pair import *
from Coordinator.PythonScripts.helpers.dictionary import *
from Coordinator.PythonScripts.helpers.elgamal import *

import os
import sys 
import time
import logging

from datetime import datetime
from math import exp, log, ceil, floor
from colorama import Fore, Back, Style, init
from logging.handlers import RotatingFileHandler

from Coordinator.PythonScripts.helpers.helper import *

init(autoreset=True)

banner = f"""
{Fore.RED}{Style.BRIGHT}
Hello, welcome to 
███████╗ ██╗ ███╗   ██╗████████╗ ██████╗   █████╗  ███████╗ ███████╗ ██████╗ 
██╔════╝ ██║ ████╗  ██║╚══██╔══╝ ██╔══██╗ ██╔══██╗ ██╔════╝ ██╔════╝ ██╔══██╗
█████╗   ██║ ██╔██╗ ██║   ██║    ██████╔╝ ███████║ ██║      █████╗   ██████╔╝
██╔══╝   ██║ ██║╚██╗██║   ██║    ██╔══██╗ ██╔══██║ ██║      ██╔══╝   ██╔══██╗
██║      ██║ ██║ ╚████║   ██║    ██║  ██║ ██║  ██║ ███████╗ ███████╗ ██║  ██║
╚═╝      ╚═╝ ╚═╝  ╚═══╝   ╚═╝    ╚═╝  ╚═╝ ╚═╝  ╚═╝ ╚══════╝ ╚══════╝ ╚═╝  ╚═╝
{Style.RESET_ALL}
{Fore.YELLOW}{Style.BRIGHT}╔════════════════════════════════════════════════════════════════════════════╗
║{'FINTRACER TIMING BENCHMARK'.center(76)}║
╚════════════════════════════════════════════════════════════════════════════╝
{Style.RESET_ALL}
"""
print(banner)

app_name = "FinTracer Timing Benchmark"

fc = setup_ftil_context(app_name)

start = time.time()

print(f"{Fore.CYAN}{Style.BRIGHT}⚙️  Initializing benchmark environment...{Style.RESET_ALL}")

# Setup
branch2node = get_branch2node(fc)
priv_key, pub_key, zero, plain_zero = setup_crypto(fc)
peer_nodes = fc.scope().difference(fc.CoordinatorID)

# Load data
transactions = load_transactions(fc)

'''
# Count total transactions
print(f"{Fore.YELLOW}📊 Counting transactions...{Style.RESET_ALL}")
with fl.on(peer_nodes):
    transaction_count = transactions.first.size()
total_transactions = sum(fc.read(transaction_count).values())
'''

# Load account sets
with fl.on(peer_nodes):
    # Count total accounts
    all_accounts = Pair(fc.array('i'), fc.array('i'))
    all_accounts.auxdb_read("SELECT DISTINCT bsb, account_id FROM accounts")
    #account_count = all_accounts.first.size()
    
    # Load test sets
    set_A = Pair(fc.array('i'), fc.array('i'))
    set_A.auxdb_read("""
    SELECT bsb, account_id FROM (
    SELECT ROW_NUMBER () OVER (
      ORDER BY bsb, account_id
    ) RowNum,
    bsb, account_id FROM (SELECT DISTINCT bsb, account_id FROM accounts) s
    ) t WHERE RowNum >= 1 AND RowNum <= 2;
    """)
    
    set_B = Pair(fc.array('i'), fc.array('i'))
    set_B.auxdb_read("""
    SELECT bsb, account_id FROM (
    SELECT ROW_NUMBER () OVER (
      ORDER BY bsb, account_id
    ) RowNum,
    bsb, account_id FROM (SELECT DISTINCT bsb, account_id FROM accounts) s
    ) t WHERE RowNum >= 3 AND RowNum <= 4;
    """)
    
    set_C = Pair(fc.array('i'), fc.array('i'))
    set_C.auxdb_read("""
    SELECT bsb, account_id FROM (
    SELECT ROW_NUMBER () OVER (
      ORDER BY bsb, account_id
    ) RowNum,
    bsb, account_id FROM (SELECT DISTINCT bsb, account_id FROM accounts) s
    ) t WHERE RowNum >= 5 AND RowNum <= 6;
    """)

#total_accounts = sum(fc.read(account_count).values())

print(f"{Fore.GREEN}✅ Setup complete{Style.RESET_ALL}")
#print(f"{Fore.BLUE}📈 Total Accounts: {total_accounts}{Style.RESET_ALL}")
#print(f"{Fore.BLUE}📈 Total Transactions: {total_transactions}{Style.RESET_ALL}")

# Initialize tag for benchmarking
tag_A = create_initial_tag(set_A, pub_key, fc)

print(f"\n{Fore.CYAN}{Style.BRIGHT}🚀 Starting performance benchmarks...{Style.RESET_ALL}")

# Benchmark 1: Propagation Steps (5 iterations)
print(f"{Fore.YELLOW}⏱️  Benchmarking propagation steps (5 iterations)...{Style.RESET_ALL}")
propagation_times = []

for i in range(5):
    step_start = time.time()
    tag_result1 = fintracer_step(tag_A, transactions, pub_key, branch2node)
    step_end = time.time()
    propagation_times.append(step_end - step_start)
    print(f"   Propagation {i+1}: {step_end - step_start:.3f}s")

avg_propagation_time = sum(propagation_times) / len(propagation_times)

# Benchmark 2: Reverse Propagation Steps (5 iterations)
print(f"{Fore.YELLOW}⏱️  Benchmarking reverse propagation steps (5 iterations)...{Style.RESET_ALL}")
transactions_rev = Pair(transactions.second, transactions.first)
reverse_propagation_times = []

for i in range(5):
    step_start = time.time()
    tag_result2 = fintracer_step(tag_A, transactions_rev, pub_key, branch2node)
    step_end = time.time()
    reverse_propagation_times.append(step_end - step_start)
    print(f"   Reverse Propagation {i+1}: {step_end - step_start:.3f}s")

avg_reverse_propagation_time = sum(reverse_propagation_times) / len(reverse_propagation_times)

# Benchmark 3: Intersection Operations (2 iterations)
print(f"{Fore.YELLOW}⏱️  Benchmarking intersection operations (2 iterations)...{Style.RESET_ALL}")

# Create two tags for intersection testing
tag_AB = fintracer_step(tag_A, transactions, pub_key, branch2node)
tag_CB = fintracer_step(tag_A, transactions_rev, pub_key, branch2node)

with fl.on(peer_nodes):
    default_zero = elgamal_encrypt(fc.array('i', 1, 0), pub_key)
    cipher_AB = tag_result1.lookup(set_B, default_zero)
    tag_AB_restricted = Dict(set_B, cipher_AB)
    cipher_CB = tag_result1.lookup(set_B, default_zero)
    tag_CB_restricted = Dict(set_B, cipher_CB)

intersection_times = []

for i in range(2):
    step_start = time.time()
    intersection_result = calc_intersection(tag_AB_restricted, tag_CB_restricted, pub_key, priv_key, zero, plain_zero, fc)
    step_end = time.time()
    intersection_times.append(step_end - step_start)
    print(f"   Intersection {i+1}: {step_end - step_start:.3f}s")

avg_intersection_time = sum(intersection_times) / len(intersection_times)

# Calculate total time
end = time.time()
total_time = end - start

# Output Results
print(f"\n{Fore.CYAN}{Style.BRIGHT}╔════════════════════════════════════════════════════════════════════════════╗")
print(f"║{'BENCHMARK RESULTS'.center(76)}║")
print(f"╚════════════════════════════════════════════════════════════════════════════╝{Style.RESET_ALL}")

print(f"\n{Fore.GREEN}{Style.BRIGHT}📊 PERFORMANCE METRICS:{Style.RESET_ALL}")
#print(f"   {Fore.BLUE}1. Total Accounts: {Style.RESET_ALL}{total_accounts:,}")
#print(f"   {Fore.BLUE}2. Total Transactions: {Style.RESET_ALL}{total_transactions:,}")
print(f"   {Fore.BLUE}3. Propagation Step Time (avg): {Style.RESET_ALL}{avg_propagation_time:.3f}s")
print(f"   {Fore.BLUE}4. Intersection Time (avg): {Style.RESET_ALL}{avg_intersection_time:.3f}s")
print(f"   {Fore.BLUE}5. Total Program Time: {Style.RESET_ALL}{total_time:.3f}s")

print(f"\n{Fore.MAGENTA}{Style.BRIGHT}📈 DETAILED TIMING BREAKDOWN:{Style.RESET_ALL}")
print(f"   {Fore.YELLOW}Forward Propagation (avg): {Style.RESET_ALL}{avg_propagation_time:.3f}s")
print(f"   {Fore.YELLOW}Reverse Propagation (avg): {Style.RESET_ALL}{avg_reverse_propagation_time:.3f}s")
print(f"   {Fore.YELLOW}Intersection Operations (avg): {Style.RESET_ALL}{avg_intersection_time:.3f}s")

minutes = int(total_time // 60)
seconds = total_time % 60
print(f"\n{Fore.GREEN}{Style.BRIGHT}✅ Benchmark completed in {minutes}m {seconds:.2f}s{Style.RESET_ALL}")
