"""
FinTracer Helper Package
Contains shared functions used across multiple FinTracer implementations.
"""

# Import all main functions to make them available at package level
from .display import display_banner, heading, run_step, read_results, out_line, prompt_for_investigation
from .crypto import setup_crypto
from .fintracer import  (
    create_initial_tag, create_initial_tags, 
    fintracer_step, distribute_tag,  diff_priv_amount, 
    read_tag, propagate_all_tags
)
from .set_operations import (
    calc_union, calc_intersection, 
    calc_multi_intersection, calc_negation, 
    calc_multi_union, calc_multi_negation, 
    calc_multi_normalize, noise_amount,
    restrict_tag_with_set, subtract_low_hits,
    compute_final_B
)
from .further_investigation import investigate_suspicious_accounts
from .load_transactions import load_transactions, get_branch2node
from .context import setup_ftil_context, create_logger
from .save_fintracer_results import prompt_for_save_results
from .load_accounts import (
    load_linear3, linear3_set_size_config, 
    load_nonlinear4, nonlinear4_set_size_config, 
    load_tree5, tree5_set_size_config, 
    load_accum, accum_config,
    partition_source_into_subsets
)

__all__ = [
    # Display functions
    'display_banner', 'heading', 'run_step', 'read_results', 
    'out_line', 'prompt_for_investigation',
    # Crypto functions  
    'setup_crypto',
    # FinTracer core functions
    'create_initial_tag', 'create_initial_tags', 'fintracer_step', 
    'distribute_tag', 'diff_priv_amount', 'read_tag', 
    'propagate_all_tags',  
    # Set operations
    'calc_union', 'calc_intersection', 
    'calc_multi_intersection', 'calc_negation', 
    'calc_multi_union', 'calc_multi_negation', 
    'calc_multi_normalize', 'noise_amount',
    'restrict_tag_with_set', 'subtract_low_hits',
    'compute_final_B', 
    # Data loading
    'load_transactions', 'get_branch2node',
    # Set context
    'setup_ftil_context', 'create_logger',
    # Further investigation
    'investigate_suspicious_accounts',
    # Load accounts
    'load_linear3', 'linear3_set_size_config', 
    'load_nonlinear4', 'nonlinear4_set_size_config', 
    'load_tree5', 'tree5_set_size_config', 
    'load_accum', 'accum_config',
    'partition_source_into_subsets',
    
    # Save results
    'prompt_for_save_results'
]