#!/usr/bin/env python3
"""
FinTracer Unified Interface - Simple model selector and executor
"""
import os
import sys
import subprocess

from rich.console import Console    # type: ignore
from rich.panel import Panel        # type: ignore
console = Console(highlight=True)   # type: ignore

def display_menu(models):
    """Display the model selection menu"""

    console.print()
    title = "FinTracer Unified Interface - Model Selector"
    panel_width = 60
    title_width = len(title)
    padding = (panel_width - title_width - 4) // 2  # -4 for panel borders
    centered_title = " " * padding + title
    
    console.print(
        Panel(
            f"[bold magenta]{centered_title}",
            width=panel_width,
            border_style="bold magenta",
        ),
        justify="left",
    )
    console.print()

    model_names = [
        '[bold green]LINEAR - 3 ACCOUNT[/bold green]',
        '[bold yellow]NONLINEAR - 4 ACCOUNT[/bold yellow]',
        '[bold blue]NON-LINEAR TREE - 5 ACCOUNT[/bold blue]',
        '[bold red]ACCUMULATION ACCOUNT[/bold red]',
        ]
    
    menu_content = []
    for i in range(len(model_names)):
        menu_content.append(f"{i+1:2d}. {model_names[i]}")
    
    menu_content.append(f"{len(models)+1:2d}. [bold white]Exit[/bold white]")
    
    console.print(
        Panel(
            "\n".join(menu_content),
            title="[bold cyan]Available Model Typologies[/bold cyan]",
            border_style="cyan",
            width=60,
        ),
        justify="left",
    )

def get_user_choice(max_choice):
    """Get valid user choice"""
    while True:
        try:
            console.print(f"\n[white]Please select a model to execute ([bold cyan]1-{max_choice}[/]):[/white]")
            choice = input(f"\n    -¬> ").strip()
            choice_num = int(choice)
            if 1 <= choice_num <= max_choice:
                return choice_num
            else:
                print(f"Please enter a number between 1 and {max_choice}")
        except ValueError:
            print("Please enter a valid number")
        except KeyboardInterrupt:
            print("\nExiting...")
            sys.exit(0)

def execute_model(model_file):
    """Execute the selected model file"""
    current_dir = os.path.dirname(os.path.abspath(__file__))
    model_path = os.path.join(current_dir, model_file)
    
    console.print(f"\nExecuting {model_file}...")
    print("="*60)
    
    try:
        # Execute the model file
        result = subprocess.run([sys.executable, model_path], 
                              cwd=current_dir,
                              check=True)
        
    except subprocess.CalledProcessError as e:
        print("="*60)
        print(f"Error executing {model_file}: {e}")
        return False
    except KeyboardInterrupt:
        print("\nExecution interrupted by user")
        return False
    
    return True

def main():
    """Main program loop"""
    while True:
        
        models = [
        'LINEAR - 3 ACCOUNT',
        'NONLINEAR - 4 ACCOUNT',
        'NON-LINEAR TREE - 5 ACCOUNT',
        'ACCUMULATION ACCOUNT',
        ]
        
        file_names = [
        'fintracer-linear3.py',
        'fintracer-nonlinear4.py',
        'fintracer-tree5.py',
        'fintracer-accum.py',
        ]

        display_menu(models)
        choice = get_user_choice(len(models) + 1)
        
        if choice == len(models) + 1:  # Exit option
            print("Goodbye!")
            break
        
        selected_model = file_names[choice - 1]
        execute_model(selected_model)
        
        # Ask if user wants to run another model
        print("\n" + "="*60)
        while True:
            continue_choice = input("Run another model? (y/n): ").strip().lower()
            if continue_choice in ['y', 'yes']:
                break
            elif continue_choice in ['n', 'no']:
                print("\nGoodbye!\n")
                return
            else:
                print("Please enter 'y' or 'n'")

if __name__ == "__main__":
    main()
