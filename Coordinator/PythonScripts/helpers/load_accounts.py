"""
Loading accounts from SQL Database for FinTracer queries
"""
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../TypeDefinitions'))

import ftillite as fl
from TypeDefinitions.pair import *
from rich.console import Console    # type: ignore
from rich.prompt import IntPrompt   # type: ignore

console = Console(highlight=True) 

def load_linear3(peer_nodes, fc, a_size=50, b_size=50, c_size=50):
    with fl.on(peer_nodes):
        source_set       = Pair(fc.array( "i" ), fc.array( "i" ) ) 
        intermediate_set = Pair( fc.array( "i" ), fc.array( "i" ) )
        target_set       = Pair( fc.array( "i" ), fc.array( "i" ) )

        source_set.auxdb_read(
            f"""
            SELECT bank_id, account_id FROM (
                SELECT ROW_NUMBER() OVER (ORDER BY bank_id, account_id) AS RowNum,
                    bank_id, account_id
                FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
            ) t WHERE RowNum BETWEEN 1 AND {a_size};
            """
        )
        intermediate_set.auxdb_read(
            f"""
            SELECT bank_id, account_id FROM (
                SELECT ROW_NUMBER() OVER (ORDER BY bank_id, account_id) AS RowNum,
                    bank_id, account_id
                FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
            ) t WHERE RowNum BETWEEN {b_size+1} AND {a_size + b_size};
            """
        )
        target_set.auxdb_read(
            f"""
            SELECT bank_id, account_id FROM (
                SELECT ROW_NUMBER() OVER (ORDER BY bank_id, account_id) AS RowNum,
                    bank_id, account_id
                FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
            ) t WHERE RowNum BETWEEN {a_size+b_size+1} AND {a_size + b_size + c_size};
            """
        )
        return source_set, intermediate_set, target_set


def linear3_set_size_config():
    console.print("📊 [white]Configure the linear typology:")
    console.print()
    console.print("[bold green] source   [bold yellow]intermediate   [bold red]target")
    console.print()

    a_size = IntPrompt.ask(
        "[bold green]     Enter size for source set A",
        default=50,
        show_default=True
    )

    b_size = IntPrompt.ask(
        "[bold yellow]     Enter size for intermediate set B",
        default=50,
        show_default=True
    )

    c_size = IntPrompt.ask(
        "[bold red]     Enter size for target set C",
        default=50,
        show_default=True
    )

    console.print()

    return a_size, b_size, c_size

    
def load_nonlinear4(peer_nodes, fc, a_size=50, b_size=50, c_size=50, d_size=100):
    with fl.on(peer_nodes):
        set_A = Pair(fc.array('i'), fc.array('i'))
        set_B = Pair(fc.array('i'), fc.array('i'))
        set_C = Pair(fc.array('i'), fc.array('i'))
        set_D = Pair(fc.array('i'), fc.array('i'))

        set_A.auxdb_read(
                f"""
                SELECT bank_id, account_id FROM (
                    SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                        bank_id, account_id 
                    FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                ) t WHERE RowNum >= 1 AND RowNum <= {a_size};
                """
        )    

        set_B.auxdb_read(
                f"""
                SELECT bank_id, account_id FROM (
                    SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                        bank_id, account_id 
                    FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                ) t WHERE RowNum >= {a_size+1} AND RowNum <= {a_size + b_size};
                """
        )

        set_C.auxdb_read(
                f"""
                SELECT bank_id, account_id FROM (
                    SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                        bank_id, account_id 
                    FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                ) t WHERE RowNum >= {a_size+b_size+1} AND RowNum <= {a_size + b_size + c_size};
                """
        )
            
        set_D.auxdb_read(
                f"""
                SELECT bank_id, account_id FROM (
                    SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                        bank_id, account_id 
                    FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                ) t WHERE RowNum >= {a_size + b_size + c_size + 1} AND RowNum <= {a_size + b_size + c_size + d_size}    ;
                """
        )
        return set_A, set_B, set_C, set_D


def nonlinear4_set_size_config():
    console.print("📊 [white]Configure the nonlinear typology:")
    console.print()
    console.print("[bold green] source   [bold yellow]intermediate   [bold red]target")
    console.print()

    a_size = IntPrompt.ask(
        "[bold green]     Enter size for source set A",
        default=50,
        show_default=True
    )

    b_size = IntPrompt.ask(
        "[bold green]     Enter size for source set B",
        default=50,
        show_default=True
    )

    c_size = IntPrompt.ask(
        "[bold green]     Enter size for source set C",
        default=50,
        show_default=True
    )

    d_size = IntPrompt.ask(
        "[bold red]     Enter size for target set D",
        default=100,
        show_default=True
    )

    console.print()

    return a_size, b_size, c_size, d_size


def load_tree5(peer_nodes, fc, c_size=50, d_size=50, e_size=50, f_size=100, g_size=100):
        with fl.on(peer_nodes):
            set_C = Pair(fc.array('i'), fc.array('i'))
            set_D = Pair(fc.array('i'), fc.array('i'))
            set_E = Pair(fc.array('i'), fc.array('i'))
            set_F = Pair(fc.array('i'), fc.array('i'))
            set_G = Pair(fc.array('i'), fc.array('i'))

            set_C.auxdb_read(
                    f"""
                    SELECT bank_id, account_id FROM (
                        SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                            bank_id, account_id 
                        FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                    ) t WHERE RowNum >= 1 AND RowNum <= {c_size};
                    """
            )

            set_D.auxdb_read(
                    f"""
                    SELECT bank_id, account_id FROM (
                        SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                            bank_id, account_id 
                        FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                    ) t WHERE RowNum >= {c_size+1} AND RowNum <= {c_size + d_size};
                    """
            )

            set_E.auxdb_read(
                    f"""
                    SELECT bank_id, account_id FROM (
                        SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                            bank_id, account_id 
                        FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                    ) t WHERE RowNum >= {c_size+d_size+1} AND RowNum <= {c_size + d_size + e_size};
                    """
            )

            set_F.auxdb_read(
                    f"""
                    SELECT bank_id, account_id FROM (
                        SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                            bank_id, account_id 
                        FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                    ) t WHERE RowNum >= {c_size+d_size+e_size+1} 
                    AND RowNum <= {c_size + d_size + e_size + f_size};
                    """
            )

            set_G.auxdb_read(
                    f"""
                    SELECT bank_id, account_id FROM (
                        SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) RowNum,
                            bank_id, account_id 
                        FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                    ) t WHERE RowNum >= {c_size + d_size + e_size + f_size + 1} 
                    AND RowNum <= {c_size + d_size + e_size + f_size + g_size};
                    """
            )
            return set_C, set_D, set_E, set_F, set_G
        

def tree5_set_size_config():
    console.print("📊 [white]Configure the nonlinear tree typology:")
    console.print()
    console.print("[bold green] source   [bold yellow]intermediate   [bold red]target")
    console.print()

    c_size = IntPrompt.ask(
        "[bold green]     Enter size for source set C",
        default=50,
        show_default=True
    )

    d_size = IntPrompt.ask(
        "[bold green]     Enter size for source set D",
        default=50,
        show_default=True
    )

    e_size = IntPrompt.ask(
        "[bold green]     Enter size for source set E",
        default=50,
        show_default=True
    )

    f_size = IntPrompt.ask(
        "[bold yellow]     Enter size for intermediate set F",
        default=100,
        show_default=True
    )

    g_size = IntPrompt.ask(
        "[bold red]     Enter size for target set G",
        default=100,
        show_default=True
    )

    console.print()

    return c_size, d_size, e_size, f_size, g_size
        

def load_accum(num_rows, peer_nodes, fc):
    with fl.on(peer_nodes):
        source_set = Pair(fc.array('i'), fc.array('i'))

        source_set.auxdb_read(
            f"""
            SELECT bank_id, account_id FROM (
                SELECT ROW_NUMBER () OVER (ORDER BY bank_id, account_id) AS RowNum,
                    bank_id, account_id 
                FROM (SELECT DISTINCT bank_id, account_id FROM accounts) s
                ) t WHERE RowNum >= 1 AND RowNum <= {num_rows};
            """
        )
        return source_set
    

def accum_config():
    console.print("📊 [white]Configure the accumulation typology:")
    console.print()
    console.print("[bold green] source   [bold yellow]intermediate   [bold red]target")
    console.print()

    num_rows = IntPrompt.ask(
        "[bold green]     Enter size for source set A",
        default=125,
        show_default=True
    )
    console.print()
    minimum_num_hits_k = IntPrompt.ask(
        "[white]     Enter [bold magenta]minimum number of hits [/]for an account to be suspicious",
        default=5,
        show_default=True
    )

    num_algorithm_iterations = IntPrompt.ask(
        "[white]     Enter number of [bold magenta]algorithm iterations (x10 ~ 98%)",
        default=10,
        show_default=True
    )
    console.print()
    
    num_source_subsets = (minimum_num_hits_k**2) // 2   # Number of subsets of the source set 
    console.print(f"[bold blue]         Optimal number of subsets of the source set: N = {num_source_subsets}")
    console.print()

    return num_rows, minimum_num_hits_k, num_algorithm_iterations


# add randomness to the set A 
def partition_source_into_subsets(source_set, peer_nodes, num_source_subsets, fc):
    with fl.on(peer_nodes):
        total_len = source_set.first.len()
            
        # Generate random permutation and shuffle set A
        perm = fc.randomperm(total_len)
        shuffled = Pair(source_set.first[perm], source_set.second[perm])
        base_size = total_len //  num_source_subsets
                
        subsets = []
        for i in range(num_source_subsets):
            start_idx = base_size * i
            if i == num_source_subsets-1:  # Last subset gets remainder
                end_idx = total_len
            else:
                end_idx = base_size * (i + 1)
                    
            subset_first = shuffled.first[start_idx:end_idx] 
            subset_second = shuffled.second[start_idx:end_idx]

            subset = Pair(subset_first, subset_second)
            subsets.append(subset)

        return subsets
