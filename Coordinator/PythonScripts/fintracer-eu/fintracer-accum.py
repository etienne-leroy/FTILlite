#!/usr/bin/env python3
'''
FinTracer Accumulation Typology

We consider a source set A of accounts (accounts sharing an implicit relationship),
and we aim to find B where each account has >=k hits in the N subsets of A. 

k = minimum number of hits for an account to be considered suspicious
N = number of subsets = (k^2)/2 
'''
import os
import sys
import time

pythonscripts_dir = os.path.dirname(os.path.dirname(__file__))
sys.path.insert(0, pythonscripts_dir) 
sys.path.insert(0, os.path.join(pythonscripts_dir, 'TypeDefinitions'))
sys.path.insert(0, os.path.join(pythonscripts_dir, 'helpers'))

import ftillite as fl
from pair import *          # type: ignore
from dictionary import *    # type: ignore
from elgamal import *       # type: ignore
from helpers import *

console = Console(highlight=True)   # type: ignore  

overall_start = time.time()

app_name = f"ACCUMULATION TYPOLOGY"

typology_graphic = [
        "\\ | /",
        "-- A --",
        "/ | \\",
]

# ────────────────────────────── Application Start ─────────────────────────────
display_banner(app_name, typology_graphic)

fc = setup_ftil_context(app_name)
# ─────────────────────────── User Input for Query ─────────────────────────
heading("QUERY CONFIGURATION")

num_rows, minimum_num_hits_k, num_algorithm_iterations = accum_config()

num_source_subsets = (minimum_num_hits_k**2) // 2   # Number of subsets of the source set 
# ─────────────────────────── Mapping initialisation ─────────────────────────
heading("MAPPING INITIALIZATION")

bank_id_mapping = run_step(
    "Setting up bank-ID-mapping", 
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
heading("DATA LOADING")

full_transactions_network = run_step(
    "Loading transaction data", 
    load_transactions, fc
)
source_set = run_step(
    "Loading accounts sets", 
    load_accum, 
    num_rows, peer_nodes, fc
)

console.print()
# ──────────────────────────  FinTracer Algorithm Execution  ───────────────────────
heading("FINTRACER ALGORITHM EXECUTION")

all_B_iterations_results = []
    
for iteration in range(num_algorithm_iterations):

    iteration_start = time.time()
    console.print()
    console.rule(f"\n[bold magenta]🔄 Starting Algorithm Iteration {iteration + 1}/{num_algorithm_iterations}", align="left", style="bold magenta")
        
    subsets = run_step(
        f"Partitioning A into {num_source_subsets} random subsets", 
        partition_source_into_subsets, 
        source_set, peer_nodes, 
        num_source_subsets, fc
    )

    all_A_subset_tags = run_step(
        f"Creating {num_source_subsets} initial encrypted tags", 
        create_initial_tags,
        subsets, public_key, fc
    )

    all_B_tags = run_step(
        f"Performing {num_source_subsets} FinTracer steps", 
        propagate_all_tags,  
        all_A_subset_tags, public_key, 
        full_transactions_network, bank_id_mapping
    )

    normalized_tags = run_step(
        f"Creating {num_source_subsets} 0-1 tags of B", 
        calc_multi_normalize, 
        all_B_tags, public_key, private_key, 
        encrypted_zero, plain_zero, fc
    )

    B_prime_tags = run_step(
        "Summing B tags to get B'", 
        calc_multi_union, 
        normalized_tags, fc
    )

    B_prime_tags_temp = run_step(
        f"Removing accounts with <{minimum_num_hits_k} hits", 
        subtract_low_hits, 
        B_prime_tags, public_key, 
        peer_nodes, minimum_num_hits_k, fc
    )

    B_result = run_step(
        f"Computing B (iteration {iteration + 1})", 
        compute_final_B, 
        B_prime_tags_temp, 
        public_key, private_key, 
        encrypted_zero, plain_zero, fc
    )
        
    all_B_iterations_results.append(B_result)
    
    iteration_end = time.time()
    elapsed = iteration_end - iteration_start
    description = f"Completed iteration {iteration + 1}/{num_algorithm_iterations}"
    console.print(f"\n    [bold green]✅  {description:<48} [purple]->     [bold cyan]{elapsed:.2f}s")

console.print()

final_B_res = run_step(
    f"Aggregating results from {num_algorithm_iterations} iterations", 
    calc_multi_union, 
    all_B_iterations_results, fc
)

console.print()
# ─────────────────────────────────  Results  ────────────────────────────────
heading("FINTRACER RESULTS")

results_B = read_results("B", final_B_res, source_set, private_key, public_key, encrypted_zero, plain_zero, fc)

saved_file = prompt_for_save_results(results_B)

if prompt_for_investigation(results_B):
    console.print()
    heading("DETAILED INVESTIGATION")
    
    report_dir = investigate_suspicious_accounts(results_B)
    
    console.print(f"\n[bold green]📁 Investigation complete! Reports saved to:[/bold green]")
    console.print(f"[bold blue]{report_dir}[/bold blue]")
else:
    if saved_file:
        console.print(f"\n[bold blue]💡 To run investigation later, use:[/bold blue]")
        console.print(f"[dim]python ../run_investigation.py -f {saved_file}[/dim]")

heading("FINTRACER EXECUTION COMPLETE")
elapsed = time.time() - overall_start
out_line(elapsed)