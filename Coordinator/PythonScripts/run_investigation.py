"""
Standalone Investigation Runner for FinTracer Results
Runs detailed analysis on previously saved FinTracer results without re-running the detection.
"""

import os
import sys
import json
import argparse
from datetime import datetime
from pathlib import Path

# Add helpers to path
sys.path.append(os.path.join(os.path.dirname(__file__), 'helpers'))

from helpers import *
from further_investigation import investigate_suspicious_accounts
from rich.console import Console
from rich.panel import Panel

console = Console()

def load_fintracer_results(results_file):
    """Load FinTracer results from JSON file"""
    try:
        with open(results_file, 'r') as f:
            data = json.load(f)
        
        # Handle both old format (direct dict) and new format (with metadata)
        if 'results' in data and 'metadata' in data:
            console.print(f"[blue]Loaded results with metadata: {data['metadata']['total_suspicious_accounts']} suspicious accounts[/blue]")
            if 'parameters' in data['metadata']:
                params = data['metadata']['parameters']
                console.print(f"[dim]Algorithm: {data['metadata'].get('algorithm', 'unknown')}, k={params.get('minimum_hits_k', 'unknown')}, iterations={params.get('num_iterations', 'unknown')}[/dim]")
            return data['results']
        else:
            # Old format - return as is
            return data
            
    except Exception as e:
        console.print(f"[bold red]Error loading results file: {e}[/bold red]")
        return None

def list_available_results(results_dir="fintracer-eu/fintracer_results"):
    """List available FinTracer result files"""
    if not os.path.exists(results_dir):
        console.print(f"[yellow]Results directory '{results_dir}' not found.[/yellow]")
        return []
    
    result_files = []
    for file in os.listdir(results_dir):
        if file.endswith('.json'):
            filepath = os.path.join(results_dir, file)
            try:
                mtime = os.path.getmtime(filepath)
                date_str = datetime.fromtimestamp(mtime).strftime('%Y-%m-%d %H:%M:%S')
                result_files.append((file, filepath, date_str))
            except:
                result_files.append((file, filepath, "Unknown"))
    
    return sorted(result_files, key=lambda x: x[2], reverse=True)

def main():
    parser = argparse.ArgumentParser(description='Run investigation analysis on saved FinTracer results')
    parser.add_argument('--results-file', '-f', type=str, 
                       help='Path to FinTracer results JSON file')
    parser.add_argument('--results-dir', '-d', type=str, default='fintracer-eu/fintracer_results',
                       help='Directory containing FinTracer results (default: fintracer-eu/fintracer_results)')
    parser.add_argument('--output-dir', '-o', type=str, default='fintracer-eu/investigation_reports',
                       help='Output directory for investigation reports (default: fintracer-eu/investigation_reports)')
    parser.add_argument('--list', '-l', action='store_true',
                       help='List available result files')
    
    args = parser.parse_args()
    
    display_banner("Deeper Investigation Runner",
                   typology_graphic=[
                       
                   ])
    
    # List available results if requested
    if args.list:
        console.print("\n[bold green]Available FinTracer Result Files:[/bold green]")
        available_files = list_available_results(args.results_dir)
        
        if not available_files:
            console.print("[yellow]No result files found.[/yellow]")
            return
        
        for i, (filename, filepath, date_str) in enumerate(available_files, 1):
            console.print(f"  {i}. [cyan]{filename}[/cyan] ({date_str})")
        
        console.print(f"\nTo run investigation on a file, use:")
        console.print(f"[dim]python run_investigation.py -f {args.results_dir}/FILENAME[/dim]")
        return
    
    # Load results file
    if args.results_file:
        if not os.path.exists(args.results_file):
            console.print(f"[bold red]Results file not found: {args.results_file}[/bold red]")
            return
        
        console.print(f"[green]Loading results from: {args.results_file}[/green]")
        results_dict = load_fintracer_results(args.results_file)
        
    else:
        # Interactive mode - let user choose from available files
        available_files = list_available_results(args.results_dir)
        
        if not available_files:
            console.print(f"[bold red]No result files found in '{args.results_dir}'.[/bold red]")
            console.print("Use --help for usage information.")
            return
        
        console.print("\n[bold green]Available FinTracer Result Files:[/bold green]")
        for i, (filename, filepath, date_str) in enumerate(available_files, 1):
            console.print(f"  {i}. [cyan]{filename}[/cyan] ({date_str})")
        
        while True:
            try:
                choice = input(f"\nSelect file to analyze (1-{len(available_files)}): ")
                choice_idx = int(choice) - 1
                if 0 <= choice_idx < len(available_files):
                    selected_file = available_files[choice_idx][1]
                    break
                else:
                    console.print("[red]Invalid selection. Please try again.[/red]")
            except (ValueError, KeyboardInterrupt):
                console.print("\n[yellow]Operation cancelled.[/yellow]")
                return
        
        console.print(f"[green]Loading results from: {selected_file}[/green]")
        results_dict = load_fintracer_results(selected_file)
    
    if results_dict is None:
        return
    
    # Validate results format
    if not isinstance(results_dict, dict):
        console.print("[bold red]Invalid results format. Expected dictionary.[/bold red]")
        return
    
    # Count total suspicious accounts
    total_accounts = sum(len(accounts) for accounts in results_dict.values() if accounts)
    console.print(f"[blue]Found results for {len(results_dict)} nodes with {total_accounts} total suspicious accounts[/blue]")
    
    if total_accounts == 0:
        console.print("[yellow]No suspicious accounts found in results. Nothing to investigate.[/yellow]")
        return
    
    # Run investigation
    console.print(f"[green]Starting investigation analysis...[/green]")
    report_dir = investigate_suspicious_accounts(results_dict, args.output_dir)
    
    console.print(f"\n[bold green]✅ Investigation completed![/bold green]")
    console.print(f"📁 Reports saved to: [blue]{report_dir}[/blue]")

if __name__ == "__main__":
    main()
