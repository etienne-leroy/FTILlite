"""
Display and UI Helper Functions for FinTracer
"""
from .fintracer import read_tag

import time
import random
from colorama import init                                                           # type: ignore
from rich.console import Console                                                    # type: ignore
from rich.panel import Panel                                                        # type: ignore
from rich.progress import Progress, SpinnerColumn, TextColumn, TimeElapsedColumn    # type: ignore
from rich.text import Text                                                          # type: ignore

init(autoreset=True)
console = Console(highlight=True)

def display_banner(title: str, typology_graphic: list = None) -> None:
    """Display the FinTracer banner with title and optional graphic"""
    art = """
███████╗ ██╗ ███╗   ██╗████████╗ ██████╗   █████╗  ███████╗ ███████╗ ██████╗ 
██╔════╝ ██║ ████╗  ██║╚══██╔══╝ ██╔══██╗ ██╔══██╗ ██╔════╝ ██╔════╝ ██╔══██╗
█████╗   ██║ ██╔██╗ ██║   ██║    ██████╔╝ ███████║ ██║      █████╗   ██████╔╝
██╔══╝   ██║ ██║╚██╗██║   ██║    ██╔══██╗ ██╔══██║ ██║      ██║      ██╔══██╗
██║      ██║ ██║ ╚████║   ██║    ██║  ██║ ██║  ██║ ███████╗ ███████╗ ██║  ██║
╚═╝      ╚═╝ ╚═╝  ╚═══╝   ╚═╝    ╚═╝  ╚═╝ ╚═╝  ╚═╝ ╚══════╝ ╚══════╝ ╚═╝  ╚═╝
"""
    console.print()
    console.print(Text(art, style="bold magenta"), justify="center")
    console.print(
        Panel(
            f"[bold magenta]{title}",
            width=80,
            border_style="bold magenta",
        ),
        justify="center",
    )
    console.print()
    
    if typology_graphic:
        console.print(Text("\n".join(typology_graphic), style="bold yellow"), justify="center")
        console.print()

def heading(title: str) -> None:
    """Display a section heading"""
    console.rule(f"\n[bold blue]{title} ", align="center")

def run_step(description, func, *args, **kwargs):
    """Execute a function with progress display"""
    spinner = SpinnerColumn(spinner_name="dots", style="bold yellow")
    text_col = TextColumn("[bold blue]{task.description}", justify="center")
    timer_col = TimeElapsedColumn()
    
    with Progress(
        spinner,
        text_col,
        timer_col,
        console=console,
        transient=True,
    ) as progress:
        task = progress.add_task(description + "......", total=None)
        start = time.perf_counter()
        result = func(*args, **kwargs)
        elapsed = time.perf_counter() - start

        progress.update(
            task,
            description=f"[bold green]{description} — done",
            completed=1
        )
        progress.refresh()

    # Format with consistent column alignment (adjust width as needed)
    console.log(f"[bold green]✔  {description:<42} [purple]->     [bold cyan]{elapsed:.2f}s")
    return result

def read_results(label, tag, the_set, priv_key, pub_key, zero, plain_zero, fc):
    """Display and read results from tag analysis"""
    console.print(f"\n📊[bold cyan] Reading {label} Results......")
    results = run_step(f"Decrypting {label} results", read_tag, tag, the_set,
                       priv_key, pub_key, zero, plain_zero, fc)
    console.print()
    total = sum(len(v) if v else 0 for v in results.values())
    
    console.print(
        Panel(
            f"[bold bright_magenta]🔍 Results: Accounts in {label} Matching Target Typology",
            width=60,
            border_style="bright_magenta",
        ),
        justify="center",
    )
    console.print()
    console.print(f"[bold green] {total} [/]total suspicious accounts found in {label}\n")
    
    for node, accounts in results.items():
        if accounts:
            console.print(f"[bold bright_magenta]🏛️  {node}")
            console.print(f"[white]   💰 Accounts Found: [bold green]{len(accounts)}")
            console.print(f"[white]   📋 Account IBANs: ")
            for i, (bsb, account_id) in enumerate(accounts):
                iban = format_account_as_iban(bsb, account_id)
                console.print(f"       [grey70 dim]{iban}")
        else:
            console.print(f"[bold red]🏛️  {node} – ❌ No matches")
            console.print()
        
        console.print("   " + "─" * 48, style="grey37 dim")

    return results

def format_account_as_iban(bsb, account_id):
    """Format account as IBAN string"""
    check_digits = f"{random.randint(10, 99)}"
    bank_code = f"{bsb:03d}"
    account_number = f"{account_id:013d}"
    
    # Construct IBAN
    iban = f"LU{check_digits}{bank_code}{account_number}"
    return iban

def out_line(elapsed):
    """Display completion summary"""
    console.print(
        f"[bold blue]\n⏱️  Total Computation Time: {int(elapsed // 60)}m {elapsed % 60:.2f}s",
        justify="center",
    )
    console.print(
        "[bold green]\n✅ FinTracer Algorithm Finished Successfully!\n",
        justify="center",
    )

def prompt_for_investigation(results):
    """Prompt user if they want to generate investigation report"""
    total_accounts = sum(len(v) if v else 0 for v in results.values())
    
    if total_accounts > 0:
        console.print()
        console.print(
            Panel(
                "[bold yellow]🔍 Deeper Investigation Available\n\n"
                "Would you like to generate a detailed investigation report?\n"
                "This will include:\n"
                "• Complete transaction history for suspicious accounts\n"
                "• Connection analysis between suspicious accounts\n"
                "• Money flow tracing and analysis\n"
                "• Detailed account information\n\n"
                "[dim]Press Enter to generate report, or 'n' to skip...[/dim]",
                width=70,
                border_style="yellow",
            ),
            justify="center",
        )
        try:
            response = input().strip().lower()
            return response != 'n'
        except KeyboardInterrupt:
            console.print("\n[yellow]Investigation skipped.[/yellow]")
            return False
    return False
