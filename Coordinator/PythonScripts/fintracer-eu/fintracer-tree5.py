#!/usr/bin/env python3
'''
FinTracer 5-account non-linear (tree) typology example . 

We consider 3 "source accounts" {C, D, E} one "intermediary accounts" {F} and one "accumulation account" G.

{C,D} feed into F
{E,F} then feed into G

            C   D
             \ /
    E         F
     \       /
      \     /
       \   /
         G
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

app_name = "5-ACCOUNT NON-LINEAR (TREE) TYPOLOGY"
    
typology_graphic = [
        "         C   D",
        "         \\ /",
        "E         F",
        "\\       /",
        "\\     /",
        "\\   /",
        "\\ /",
        "G",
]

#
# OUTPUT GIVING ERRORS -------------------------------- Investigate why----------------------------------------
#

# ────────────────────────────── Application Start ─────────────────────────────    
display_banner(app_name, typology_graphic)

fc = setup_ftil_context(app_name)
# ─────────────────────────── User Input for Set Sizes ─────────────────────────
heading("SET SIZE CONFIGURATION")

c_size, d_size, e_size, f_size, g_size = tree5_set_size_config()
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

full_transactions_network = run_step(
    "Loading transaction data", 
    load_transactions, fc
)
account_sets = run_step(
        "Loading accounts sets", 
        load_tree5, 
        peer_nodes, fc,
        c_size, d_size, 
        e_size, f_size, g_size
)

set_C, set_D, set_E, set_F, set_G = account_sets

console.print()
# ──────────────────────────  FinTracer Algorithm Execution  ───────────────────────
heading("FINTRACER ALGORITHM EXECUTION")

initialized_source_tags = run_step(
    "Creating initial encrypted tags", 
    create_initial_tags,
    account_sets[:3], public_key, fc
)

propagated_tag_C = run_step(
    "FinTracer step C ➜  F", 
    fintracer_step, 
    initialized_source_tags[0], 
    full_transactions_network, 
    public_key, bank_id_mapping
)
propagated_tag_D = run_step(
    "FinTracer step D ➜  E", 
    fintracer_step, 
    initialized_source_tags[1], 
    full_transactions_network, 
    public_key, bank_id_mapping
)
restricted_tag_CF = run_step(
    "Restricting C ➜  F", 
    restrict_tag_with_set,
    propagated_tag_C, set_F, 
    default_zero, peer_nodes
)
intersected_tag_CDF = run_step(
    "Calculating intersection at F", 
    calc_intersection,
    restricted_tag_CF, 
    propagated_tag_D, 
    public_key, private_key, 
    encrypted_zero, plain_zero, fc
)
propagated_tag_E = run_step(
    "FinTracer step E ➜  F", 
    fintracer_step, 
    initialized_source_tags[2], 
    full_transactions_network, 
    public_key, bank_id_mapping
)
propagated_tag_F = run_step(
    "FinTracer step F ➜  G",
    fintracer_step, 
    intersected_tag_CDF, 
    full_transactions_network, 
    public_key, bank_id_mapping
)
restricted_tag_EG = run_step(
    "Restricting E ➜  G", 
    restrict_tag_with_set,
    propagated_tag_E, set_G,    
    default_zero, peer_nodes
)
intersected_tag_EFG = run_step(     # Final result G
    "Calculating intersection at G", 
    calc_intersection,
    restricted_tag_EG, 
    propagated_tag_F, 
    public_key, private_key, 
    encrypted_zero, plain_zero, fc
)

# Reversed graph
transactions_rev = Pair(full_transactions_network.second, full_transactions_network.first)

propagated_tag_G_rev = run_step(
    "FinTracer reverse step G ➜ E / F", 
    fintracer_step, 
    intersected_tag_EFG, 
    transactions_rev, 
    public_key, bank_id_mapping
)
restricted_tag_GE = run_step(       # Final result E
    "Restricting G ➜  E", 
    restrict_tag_with_set,
    propagated_tag_G_rev, set_E,
    default_zero, peer_nodes
)
restricted_tag_GF = run_step(
    "Restricting G ➜  F", 
    restrict_tag_with_set,
    propagated_tag_G_rev, set_F,
    default_zero, peer_nodes
)
intersected_tag_F = run_step(       # Final result F
    "Calculating intersection at F", 
    calc_intersection,
    restricted_tag_GF, 
    intersected_tag_CDF, 
    public_key, private_key, 
    encrypted_zero, plain_zero, fc
)
propagated_tag_F_rev = run_step(
    "FinTracer reverse step F ➜ C / D", 
    fintracer_step, 
    intersected_tag_F, 
    transactions_rev, 
    public_key, bank_id_mapping
)

console.print()
# ─────────────────────────────────  Results  ────────────────────────────────
heading("FINTRACER RESULTS")

results_C = read_results(
    "C", propagated_tag_F_rev, 
    set_C, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_D = read_results(
    "D", propagated_tag_F_rev, 
    set_D, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_E = read_results(
    "E", restricted_tag_GE, 
    set_E, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_F = read_results(
    "F", intersected_tag_F, 
    set_F, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)
results_G = read_results(
    "G", intersected_tag_EFG, 
    set_G, private_key, public_key, 
    encrypted_zero, plain_zero, fc
)

# Combine all results for investigation
all_results = {**results_C, **results_D, **results_E, **results_F, **results_G}

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

heading("FINTRACER EXECUTION COMPLETE")
elapsed = time.time() - overall_start
out_line(elapsed)