"""
Investigation Helper Functions for FinTracer
Provides detailed analysis and reporting on suspicious accounts identified by FinTracer.
"""

import os
import sys
import json
import pandas as pd
from datetime import datetime
from collections import defaultdict, Counter
from sqlalchemy import create_engine, text
from dotenv import load_dotenv
from os.path import join, dirname
import random

from rich.console import Console    # type: ignore
from rich.panel import Panel        # type: ignore
from rich.table import Table        # type: ignore
from rich.text import Text          # type: ignore

console = Console()

def setup_database_connection():
    """Setup database connection using environment variables"""
    dotenv_path = join(dirname(__file__), '../../..', '.env')
    load_dotenv(dotenv_path)
    
    username = os.environ.get("DB_DEFAULT_USER", "postgres")
    password = os.environ.get("DB_DEFAULT_PW", "postgres")
    server = os.environ.get("DB_HOST", "localhost")
    port = os.environ.get("DB_LOCAL_PORT", "5432")
    database = os.environ.get("DB_DEFAULT_DB", "ftillite")
    
    engine = create_engine(f'postgresql://{username}:{password}@{server}:{port}/{database}')
    return engine

def get_suspicious_accounts_list(results_dict):
    """Extract all suspicious accounts from results dictionary"""
    suspicious_accounts = []
    for node, accounts in results_dict.items():
        if accounts:
            for bank_id, account_id in accounts:
                suspicious_accounts.append({
                    'node': node,
                    'bank_id': bank_id,
                    'account_id': account_id,
                    'iban': format_account_as_iban(bank_id, account_id)
                })
    return suspicious_accounts

def format_account_as_iban(bank_id, account_id):
    """Format account as IBAN string"""
    check_digits = f"{random.randint(10, 99)}"
    bank_code = f"{bank_id:03d}"
    account_number = f"{account_id:013d}"
    
    # Construct IBAN
    iban = f"LU{check_digits}{bank_code}{account_number}"
    return iban

def get_account_details(engine, suspicious_accounts):
    """Get detailed account information for suspicious accounts"""
    if not suspicious_accounts:
        return pd.DataFrame()
    
    # Create query to get account details
    account_conditions = []
    for acc in suspicious_accounts:
        account_conditions.append(f"(bank_id = {acc['bank_id']} AND account_id = {acc['account_id']})")
    
    query = f"""
    SELECT 
        re,
        bank_id,
        account_id,
        account_open_date,
        account_type,
        cust_occupation
    FROM accounts 
    WHERE {' OR '.join(account_conditions)}
    ORDER BY re, bank_id, account_id
    """
    
    with engine.connect() as conn:
        account_details = pd.read_sql(query, conn)
    
    return account_details

def get_transaction_details(engine, suspicious_accounts, limit_per_account=1000):
    """Get transactions involving suspicious accounts with limits to prevent crashes"""
    if not suspicious_accounts:
        return pd.DataFrame()
    
    # Get transactions in batches to avoid memory issues
    all_transactions = []
    
    # Process accounts in smaller batches
    batch_size = 10  # Process 10 accounts at a time
    for i in range(0, len(suspicious_accounts), batch_size):
        batch = suspicious_accounts[i:i + batch_size]
        
        account_conditions = []
        for acc in batch:
            account_conditions.append(f"(origin_bank_id = {acc['bank_id']} AND origin_id = {acc['account_id']})")
            account_conditions.append(f"(dest_bank_id = {acc['bank_id']} AND dest_id = {acc['account_id']})")
        
        # Limit transactions per batch to prevent memory issues
        query = f"""
        SELECT 
            origin_re,
            origin_bank_id,
            origin_id,
            dest_re,
            dest_bank_id,
            dest_id,
            total_amount,
            transaction_date
        FROM transactions 
        WHERE {' OR '.join(account_conditions)}
        ORDER BY transaction_date DESC, total_amount DESC
        LIMIT {limit_per_account * len(batch)}
        """
        
        with engine.connect() as conn:
            batch_transactions = pd.read_sql(query, conn)
            all_transactions.append(batch_transactions)
    
    # Combine all batches
    if all_transactions:
        return pd.concat(all_transactions, ignore_index=True)
    else:
        return pd.DataFrame()

def analyze_connections(transactions_df, suspicious_accounts):
    """Analyze connections between suspicious accounts and external accounts"""
    suspicious_set = {(acc['bank_id'], acc['account_id']) for acc in suspicious_accounts}
    
    # Separate internal (suspicious-to-suspicious) vs external transactions
    internal_connections = []
    external_transactions = []
    connection_summary = defaultdict(lambda: {'out_count': 0, 'in_count': 0, 'out_amount': 0, 'in_amount': 0})
    
    for _, tx in transactions_df.iterrows():
        origin_key = (tx['origin_bank_id'], tx['origin_id'])
        dest_key = (tx['dest_bank_id'], tx['dest_id'])
        
        # Check if both accounts are suspicious (internal connections)
        if origin_key in suspicious_set and dest_key in suspicious_set:
            internal_connections.append({
                'origin_iban': format_account_as_iban(tx['origin_bank_id'], tx['origin_id']),
                'dest_iban': format_account_as_iban(tx['dest_bank_id'], tx['dest_id']),
                'origin_re': tx['origin_re'],
                'dest_re': tx['dest_re'],
                'amount': tx['total_amount'],
                'date': tx['transaction_date'],
                'connection_type': 'suspicious_to_suspicious'
            })
            
            # Update connection summary for internal connections
            origin_iban = format_account_as_iban(tx['origin_bank_id'], tx['origin_id'])
            dest_iban = format_account_as_iban(tx['dest_bank_id'], tx['dest_id'])
            
            connection_summary[origin_iban]['out_count'] += 1
            connection_summary[origin_iban]['out_amount'] += float(tx['total_amount'])
            connection_summary[dest_iban]['in_count'] += 1
            connection_summary[dest_iban]['in_amount'] += float(tx['total_amount'])
        else:
            # External transaction (one side is suspicious, other is not)
            external_transactions.append({
                'origin_iban': format_account_as_iban(tx['origin_bank_id'], tx['origin_id']),
                'dest_iban': format_account_as_iban(tx['dest_bank_id'], tx['dest_id']),
                'origin_re': tx['origin_re'],
                'dest_re': tx['dest_re'],
                'amount': tx['total_amount'],
                'date': tx['transaction_date'],
                'origin_is_suspicious': origin_key in suspicious_set,
                'dest_is_suspicious': dest_key in suspicious_set,
                'connection_type': 'external'
            })
    
    return internal_connections, external_transactions, dict(connection_summary)

def trace_money_flow(transactions_df, suspicious_accounts):
    """Trace money flow from and to suspicious accounts (optimized)"""
    suspicious_set = {(acc['bank_id'], acc['account_id']) for acc in suspicious_accounts}
    
    flow_analysis = {}
    
    # Process each suspicious account individually to avoid memory issues
    for acc in suspicious_accounts:
        acc_key = (acc['bank_id'], acc['account_id'])
        
        # Filter transactions for this specific account
        account_txs = transactions_df[
            ((transactions_df['origin_bank_id'] == acc['bank_id']) & (transactions_df['origin_id'] == acc['account_id'])) |
            ((transactions_df['dest_bank_id'] == acc['bank_id']) & (transactions_df['dest_id'] == acc['account_id']))
        ]
        
        if len(account_txs) == 0:
            continue
            
        # Analyze incoming flows
        incoming_txs = account_txs[
            (account_txs['dest_bank_id'] == acc['bank_id']) & (account_txs['dest_id'] == acc['account_id'])
        ]
        incoming_summary = Counter(incoming_txs['origin_re'].tolist())
        total_incoming = incoming_txs['total_amount'].sum()
        
        # Analyze outgoing flows
        outgoing_txs = account_txs[
            (account_txs['origin_bank_id'] == acc['bank_id']) & (account_txs['origin_id'] == acc['account_id'])
        ]
        outgoing_summary = Counter(outgoing_txs['dest_re'].tolist())
        total_outgoing = outgoing_txs['total_amount'].sum()
        
        flow_analysis[acc['iban']] = {
            'total_incoming': float(total_incoming),
            'total_outgoing': float(total_outgoing),
            'net_flow': float(total_incoming - total_outgoing),
            'incoming_sources': dict(incoming_summary),
            'outgoing_destinations': dict(outgoing_summary),
            'transaction_count': len(account_txs)
        }
    
    return flow_analysis

def generate_investigation_report(results_dict, output_dir="v2.0/investigation_reports"):
    """Generate comprehensive investigation report"""
    console.print("\n🔍 [bold blue]Starting Investigation Report Generation...[/bold blue]")
    
    # Create output directory
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    timestamp = datetime.now().strftime('%Y-%m-%d_%H-%M-%S')
    report_dir = os.path.join(output_dir, f"investigation_{timestamp}")
    os.makedirs(report_dir, exist_ok=True)
    
    # Setup database
    engine = setup_database_connection()
    
    # Get suspicious accounts
    suspicious_accounts = get_suspicious_accounts_list(results_dict)
    
    if not suspicious_accounts:
        console.print("[bold red]No suspicious accounts found to investigate.[/bold red]")
        return
    
    console.print(f"[bold green]Found {len(suspicious_accounts)} suspicious accounts to investigate[/bold green]")
    
    # Get detailed data
    console.print("📊 Gathering account details...")
    account_details = get_account_details(engine, suspicious_accounts)
    
    console.print("💰 Gathering transaction data...")
    transactions = get_transaction_details(engine, suspicious_accounts)
    
    console.print("🔗 Analyzing connections...")
    internal_connections, external_transactions, connection_summary = analyze_connections(transactions, suspicious_accounts)
    
    console.print("🌊 Analyzing money flows...")
    flow_analysis = trace_money_flow(transactions, suspicious_accounts)
    
    # Generate reports
    console.print("📝 Generating detailed reports...")
    
    # 1. Summary Report
    summary_report = generate_summary_report(suspicious_accounts, account_details, 
                                            transactions, internal_connections, external_transactions, 
                                            flow_analysis)
    
    # 2. Save detailed data to files
    account_details.to_csv(os.path.join(report_dir, "suspicious_accounts_details.csv"), index=False)
    transactions.to_csv(os.path.join(report_dir, "all_transactions.csv"), index=False)
    
    # 3. Save connections data
    if internal_connections:
        internal_df = pd.DataFrame(internal_connections)
        internal_df.to_csv(os.path.join(report_dir, "internal_suspicious_connections.csv"), index=False)
    
    if external_transactions:
        external_df = pd.DataFrame(external_transactions)
        external_df.to_csv(os.path.join(report_dir, "external_transactions.csv"), index=False)
    
    # 4. Save analysis results as JSON
    analysis_results = {
        'suspicious_accounts_summary': [
            {
                'iban': acc['iban'],
                'node': acc['node'],
                'bank_id': acc['bank_id'],
                'account_id': acc['account_id']
            } for acc in suspicious_accounts
        ],
        'connection_summary': connection_summary,
        'flow_analysis': flow_analysis,
        'investigation_metadata': {
            'generated_at': timestamp,
            'total_suspicious_accounts': len(suspicious_accounts),
            'total_transactions_analyzed': len(transactions),
            'internal_suspicious_connections': len(internal_connections),
            'external_transactions': len(external_transactions)
        }
    }
    
    with open(os.path.join(report_dir, "investigation_analysis.json"), 'w') as f:
        json.dump(analysis_results, f, indent=2, default=str)
    
    # 5. Generate human-readable summary
    with open(os.path.join(report_dir, "investigation_summary.txt"), 'w') as f:
        f.write(summary_report)
    
    # Display summary to console
    display_investigation_summary(suspicious_accounts, account_details, 
                                 transactions, internal_connections, external_transactions, 
                                 flow_analysis)
    
    return report_dir

def generate_summary_report(suspicious_accounts, account_details, transactions, 
                          internal_connections, external_transactions, flow_analysis):
    """Generate human-readable summary report"""
    report = []
    report.append("="*80)
    report.append("FINTRACER INVESTIGATION REPORT")
    report.append("="*80)
    report.append(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    report.append(f"Total Suspicious Accounts: {len(suspicious_accounts)}")
    report.append(f"Total Transactions Analyzed: {len(transactions)}")
    report.append(f"Internal Suspicious-to-Suspicious Connections: {len(internal_connections)}")
    report.append(f"External Transactions (involving suspicious accounts): {len(external_transactions)}")
    report.append("")
    
    # Account breakdown by node
    node_breakdown = defaultdict(int)
    for acc in suspicious_accounts:
        node_breakdown[acc['node']] += 1
    
    report.append("SUSPICIOUS ACCOUNTS BY NODE:")
    report.append("-" * 40)
    for node, count in sorted(node_breakdown.items()):
        report.append(f"  {node}: {count} accounts")
    report.append("")
    
    # Top flow analysis
    if flow_analysis:
        report.append("TOP MONEY FLOW ANALYSIS:")
        report.append("-" * 40)
        
        # Sort by net flow
        sorted_flows = sorted(flow_analysis.items(), 
                            key=lambda x: abs(x[1]['net_flow']), reverse=True)
        
        for iban, flow in sorted_flows[:10]:  # Top 10
            report.append(f"  Account: {iban}")
            report.append(f"    Net Flow: ${flow['net_flow']:,.2f}")
            report.append(f"    Total In: ${flow['total_incoming']:,.2f}")
            report.append(f"    Total Out: ${flow['total_outgoing']:,.2f}")
            report.append(f"    Transactions: {flow['transaction_count']}")
            report.append("")
    
    report.append("="*80)
    return "\n".join(report)

def display_investigation_summary(suspicious_accounts, account_details, transactions, 
                                internal_connections, external_transactions, flow_analysis):
    """Display investigation summary to console"""
    console.print()
    console.print(
        Panel(
            "[bold magenta]🔍 INVESTIGATION SUMMARY",
            width=80,
            border_style="bold magenta",
        ),
        justify="center",
    )
    
    # Create summary table
    table = Table(title="Investigation Statistics", show_header=True, header_style="bold blue")
    table.add_column("Metric", style="cyan", no_wrap=True)
    table.add_column("Count", style="bold magenta")
    
    table.add_row("Suspicious Accounts", str(len(suspicious_accounts)))
    table.add_row("Suspicious Transaction Count", str(len(transactions)))
    table.add_row("External Transactions", str(len(external_transactions)))
    table.add_row("Internal Suspicious Connections", str(len(internal_connections)))

    console.print(table)
    
    # Node breakdown
    if suspicious_accounts:
        node_breakdown = defaultdict(int)
        for acc in suspicious_accounts:
            node_breakdown[acc['node']] += 1
        
        console.print("\n[bold blue]📊 Suspicious Accounts by Node:[/bold blue]")
        console.print(f"\n[bright green] {len(suspicious_accounts)}[white] total suspicious accounts found across[bold magenta] 4[white] nodes.\n")
        for node, count in sorted(node_breakdown.items()):
            console.print(f"  🏛️  [bold magenta]{node}[/bold magenta]: [green]{count}[/green] accounts")
    
    # Transaction breakdown
    console.print("\n[bold blue]💰 Transaction Breakdown:[/bold blue]")
    console.print(f"  📊 Total (suspicious) transactions analyzed: [blue]{len(transactions)}[/blue]")
    console.print(f"  🌐 External (involving suspicious accounts): [yellow]{len(external_transactions)}[/yellow]")
    console.print(f"  🔗 Internal (suspicious-to-suspicious): [green]{len(internal_connections)}[/green]")
    
    # Top suspicious flows
    if flow_analysis:
        console.print("\n[bold blue]💰 Top Suspicious Account Flows:[/bold blue]")
        sorted_flows = sorted(flow_analysis.items(), 
                            key=lambda x: abs(x[1]['net_flow']), reverse=True)
        
        for i, (iban, flow) in enumerate(sorted_flows[:5], 1):
            net_flow_color = "red" if flow['net_flow'] < 0 else "green"
            console.print(f"  {i}. [dim]{iban}[/dim]")
            console.print(f"     Net Flow: [{net_flow_color}]${flow['net_flow']:,.2f}[/{net_flow_color}] "
                         f"(In: ${flow['total_incoming']:,.2f}, Out: ${flow['total_outgoing']:,.2f})")
 
def load_results_from_file(filepath):
    """Load FinTracer results from saved JSON file"""
    try:
        with open(filepath, 'r') as f:
            data = json.load(f)
        
        # Handle both old format (direct dict) and new format (with metadata)
        if 'results' in data and 'metadata' in data:
            console.print(f"[blue]Loaded results with metadata: {data['metadata']['total_suspicious_accounts']} suspicious accounts[/blue]")
            return data['results']
        else:
            # Old format - return as is
            return data
            
    except Exception as e:
        console.print(f"[bold red]Error loading results file: {e}[/bold red]")
        return None

def investigate_suspicious_accounts(results_dict, output_dir="investigation_reports"):
    """Main function to investigate suspicious accounts from FinTracer results"""
    return generate_investigation_report(results_dict, output_dir)

def investigate_from_file(results_file, output_dir="investigation_reports"):
    """Investigate suspicious accounts from a saved results file"""
    console.print(f"[blue]Loading FinTracer results from: {results_file}[/blue]")
    results_dict = load_results_from_file(results_file)
    
    if results_dict is None:
        console.print("[bold red]Failed to load results file.[/bold red]")
        return None
    
    return investigate_suspicious_accounts(results_dict, output_dir)
