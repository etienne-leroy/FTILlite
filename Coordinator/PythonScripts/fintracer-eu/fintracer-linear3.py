#!/usr/bin/env python3
"""
FinTracer 3-account linear typology example.
We consider a "source account" set A which feeds into an "intermediary account" set B, 
and feeds set B to a "sink account" set C.

A -> B -> C
"""
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

app_name = "3-ACCOUNT LINEAR TYPOLOGY"
typology_graphic = [
    "A",
    "|",
    "B",
    "|",
    "C",
]

# ────────────────────────────── Application Start ─────────────────────────────
display_banner(app_name, typology_graphic)

fc = setup_ftil_context(app_name)
# ─────────────────────────── User Input for Set Sizes ─────────────────────────
heading("QUERY CONFIGURATION")

a_size, b_size, c_size = linear3_set_size_config()
# ─────────────────────────── Mapping initialisation ─────────────────────────
heading("MAPPING INITIALIZATION")

bank_id_mapping = run_step(
    "Setting up branch-ID-mapping", 
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
console.print(f"[bold magenta]    🏛️  Active peer nodes: { [n.name() for n in peer_nodes] }")
console.print()
# ───────────────────────────────── Data loading ─────────────────────────────
heading( "DATA LOADING" )

full_transaction_network = run_step(
    "Loading transaction data", 
    load_transactions, fc 
)
source_set, intermediate_set, target_set = run_step( 
    "Loading accounts sets", 
    load_linear3, peer_nodes, fc,
    a_size, b_size, c_size
)

console.print()
# ──────────────────────────  FinTracer Algorithm Execution  ───────────────────────
heading("FINTRACER ALGORITHM EXECUTION")

initialized_source_tag = run_step(
    "Creating initial encrypted tags", 
    create_initial_tag,
    source_set, public_key, fc
)

# Forward graph
propagated_source_tag = run_step(
    "FinTracer step A ➜  B", 
    fintracer_step,
    initialized_source_tag, full_transaction_network,
    public_key, bank_id_mapping
)
restricted_source_tag = run_step(
    "Restricting A ➜  B", 
    restrict_tag_with_set,
    propagated_source_tag, intermediate_set,
    default_zero, peer_nodes
)
propagated_intermediate_tag = run_step(
    "FinTracer step B ➜  C", 
    fintracer_step,
    restricted_source_tag, full_transaction_network,
    public_key, bank_id_mapping
)

# Reversed graph
transactions_rev = Pair(full_transaction_network.second, full_transaction_network.first)

reversed_initialized_target_tag = run_step(
    "Reverse initialisation", 
    restrict_tag_with_set, 
    propagated_intermediate_tag, target_set, 
    default_zero, peer_nodes
)
propagated_target_tag = run_step(
    "FinTracer reverse step C ➜  B",
    fintracer_step,
    reversed_initialized_target_tag,
    transactions_rev, public_key, 
    bank_id_mapping 
)
restricted_target_tag = run_step(
    "Restricting C ➜  B", 
    restrict_tag_with_set, 
    propagated_target_tag, 
    intermediate_set, 
    default_zero, peer_nodes
)
intermediate_tags_intersect = run_step(
    "Calculating B intersection", 
    calc_intersection,
    restricted_source_tag, 
    restricted_target_tag, public_key, 
    private_key, encrypted_zero, plain_zero, fc 
)
propagated_intermediate_tag_rev = run_step( 
    "FinTracer reverse step B ➜  A", 
    fintracer_step,
    restricted_target_tag, 
    transactions_rev, 
    public_key, bank_id_mapping
)

console.print()
# ─────────────────────────────────  Results  ────────────────────────────────
heading( "FINTRACER RESULTS" )

results_source = read_results( 
    "A", propagated_intermediate_tag_rev, 
    source_set, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_intermediate = read_results( 
    "B", restricted_target_tag, 
    intermediate_set, 
    private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_targets = read_results( 
    "C", propagated_intermediate_tag, 
    target_set, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)

# Combine all results for investigation
all_results = {**results_source, **results_intermediate, **results_targets}

elapsed = time.time() - overall_start
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
out_line(elapsed)
