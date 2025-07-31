"""
Utility to save FinTracer results for later investigation analysis
"""

import os
import json
from datetime import datetime
from rich.console import Console

console = Console()

def save_fintracer_results(results_dict, output_dir="fintracer_results", custom_name=None):
    """
    Save FinTracer results to JSON file for later analysis
    
    Args:
        results_dict: Dictionary of FinTracer results (node -> list of (bsb, account_id))
        output_dir: Directory to save results (default: fintracer_results)
        custom_name: Custom filename (optional)
    
    Returns:
        str: Path to saved results file
    """
    # Create output directory
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    # Generate filename
    if custom_name:
        filename = f"{custom_name}.json"
    else:
        timestamp = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
        filename = f"fintracer_results_{timestamp}.json"
    
    filepath = os.path.join(output_dir, filename)
    
    # Prepare results with metadata
    results_with_metadata = {
        'metadata': {
            'generated_at': datetime.now().isoformat(),
            'total_nodes': len(results_dict),
            'total_suspicious_accounts': sum(len(accounts) for accounts in results_dict.values() if accounts),
            'nodes_with_suspicious_accounts': len([node for node, accounts in results_dict.items() if accounts])
        },
        'results': results_dict
    }
    
    # Save to JSON
    try:
        with open(filepath, 'w') as f:
            json.dump(results_with_metadata, f, indent=2, default=str)
        
        console.print(f"[green]✅ FinTracer results saved to: {filepath}[/green]")
        console.print(f"[blue]📊 Total suspicious accounts: {results_with_metadata['metadata']['total_suspicious_accounts']}[/blue]")
        
        return filepath
        
    except Exception as e:
        console.print(f"[bold red]❌ Error saving results: {e}[/bold red]")
        return None
    
def prompt_for_save_results(results_dict):
    """Prompt user to save results for later analysis"""
    try:
        save_choice = input("\n💾 Save results for later analysis? (y/n): ").lower().strip()
        if save_choice in ['y', 'yes']:
            custom_name = input("Enter custom filename (optional, press Enter for auto-generated): ").strip()
            if not custom_name:
                custom_name = None
            return save_fintracer_results(results_dict, custom_name=custom_name)
        return None
    except KeyboardInterrupt:
        console.print("\n[yellow]Skipping save...[/yellow]")
        return None

def load_fintracer_results_for_analysis(filepath):
    """
    Load FinTracer results and extract just the results dict for analysis
    
    Args:
        filepath: Path to saved results JSON file
    
    Returns:
        dict: Results dictionary suitable for further_investigation.py
    """
    try:
        with open(filepath, 'r') as f:
            data = json.load(f)
        
        # Handle both old format (direct dict) and new format (with metadata)
        if 'results' in data and 'metadata' in data:
            return data['results']
        else:
            # Old format - return as is
            return data
            
    except Exception as e:
        console.print(f"[bold red]Error loading results: {e}[/bold red]")
        return None
