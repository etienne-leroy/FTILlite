#!/usr/bin/env python3
'''
FinTracer 4-account non-linear typology example. 
We consider 3 "source account" sets {A, B, C} and 1 "sink account" set D. Each {A, B, C} goes to D independently.

  A   B   C
   \  |  /  
    \ | /
      D
'''
import os
import sys
import time

pythonscripts_dir = os.path.dirname(os.path.dirname(__file__))
sys.path.insert(0, pythonscripts_dir)
sys.path.insert(0, os.path.join(pythonscripts_dir, "TypeDefinitions"))
sys.path.insert(0, os.path.join(pythonscripts_dir, "helpers"))

from pair import Pair       # type: ignore
from dictionary import *    # type: ignore
from elgamal import *       # type: ignore
from helpers import *

console = Console(highlight=True)   # type: ignore

overall_start = time.time()

app_name = "4-ACCOUNT NON-LINEAR TYPOLOGY"

typology_graphic = [
        "A    B    C",
        "\\   |   /",
        " \\  |  /",
        " \\ | /",
        " D",
]

# ────────────────────────────── Application Start ─────────────────────────────
display_banner(app_name, typology_graphic)

fc = setup_ftil_context(app_name)
# ─────────────────────────── User Input for Set Sizes ─────────────────────────
heading("QUERY CONFIGURATION")

a_size, b_size, c_size, d_size = nonlinear4_set_size_config()
# ─────────────────────────── Mapping initialisation ─────────────────────────
heading("MAPPING INITIALIZATION")
    
bank_id_mapping = run_step(
    "Setting up branch-to-node mapping", 
    get_branch2node, fc
)

console.print()
# ───────────────────────────── Cryptographic setup ──────────────────────────
heading("CRYPTOGRAPHIC SETUP")
cryptographic_setup = run_step(
        "Generating ElGamal key-pair", 
        setup_crypto, fc
)

private_key, public_key, encrypted_zero, default_zero, plain_zero = cryptographic_setup

console.print()
peer_nodes = fc.scope().difference(fc.CoordinatorID)
console.print(f"[bold magenta]    🏛️  Active peer nodes: {[n.name() for n in peer_nodes]}")
console.print()
# ───────────────────────────────── Data loading ─────────────────────────────
heading("DATA LOADING")

full_transaction_network = run_step(
    "Loading transaction data", 
    load_transactions, fc
)

account_sets = run_step(
    "Loading accounts sets", 
    load_nonlinear4,
    peer_nodes, fc, 
    a_size, b_size, 
    c_size, d_size
)

set_A, set_B, set_C, set_D = account_sets

console.print()
# ──────────────────────────  FinTracer Algorithm Execution  ───────────────────────
heading("FINTRACER ALGORITHM EXECUTION")

initialized_source_tags = run_step(
    "Creating initial encrypted tags", 
    create_initial_tags,
    account_sets[:3], public_key, fc
)

# Forward graph
propagated_tag_A = run_step(
    "FinTracer step A ➜  D", 
    fintracer_step,
    initialized_source_tags[0], 
    full_transaction_network,
    public_key, bank_id_mapping
)
propagated_tag_B = run_step(
    "FinTracer step B ➜  D", 
    fintracer_step,
    initialized_source_tags[1], 
    full_transaction_network,
    public_key, bank_id_mapping
)
propagated_tag_C = run_step(
    "FinTracer step C ➜  D", 
    fintracer_step,
    initialized_source_tags[2], 
    full_transaction_network,
    public_key, bank_id_mapping
)

restricted_tag_AD = run_step(
    "Restricting A ➜  D", 
    restrict_tag_with_set,
    propagated_tag_A, set_D,
    default_zero, peer_nodes
)

intersection_set = [restricted_tag_AD, propagated_tag_B, propagated_tag_C]
intersection_D = run_step(
    "Calculating intersection at D", 
    calc_multi_intersection,
    intersection_set,
    public_key, private_key,
    encrypted_zero, plain_zero, fc
)

# Reversed graph
transactions_rev = Pair(full_transaction_network.second, full_transaction_network.first)

propagated_tag_D_rev = run_step(
    "FinTracer Reverse Step D ➜ ", 
    fintracer_step, 
    intersection_D, transactions_rev, 
    public_key, bank_id_mapping
)

console.print()
# ─────────────────────────────────  Results  ────────────────────────────────
heading("FINTRACER RESULTS")

results_A = read_results(
    "A", propagated_tag_D_rev, 
    set_A, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_B = read_results(
    "B", propagated_tag_D_rev, 
    set_B, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_C = read_results(
    "C", propagated_tag_D_rev, 
    set_C, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_D = read_results(
    "D", intersection_D, 
    set_D, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)

# Combine all results for investigation
all_results = {**results_A, **results_B, **results_C, **results_D}

saved_file = prompt_for_save_results(all_results)

if prompt_for_investigation(all_results):
    console.print()
    heading("DETAILED INVESTIGATION")
    
    report_dir = investigate_suspicious_accounts(all_results)
    
    console.print(f"\n[bold green]📁 Investigation complete! Reports saved to:[/bold green]")
    console.print(f"[bold blue]{report_dir}[/bold blue]")
else:
    if saved_file:
        console.print(f"\n[bold blue]💡 To run investigation later, use:[/bold blue]")
        console.print(f"[dim]python ../run_investigation.py -f {saved_file}[/dim]")

heading( "FINTRACER EXECUTION COMPLETE" )
elapsed = time.time() - overall_start
out_line(elapsed)