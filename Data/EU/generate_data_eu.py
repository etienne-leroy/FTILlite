# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

"""Script to generate synthetic domestic transaction data to assist
with FinTracer prototype development.

The main purpose of the script is to generate two csv files:
'transactions.csv' and 'accounts.csv'.  The 'transactions.csv' file is
also split into four csv files corresponding to the transactions
relevant to each Peer.

Due to the large number of parameters associated with defining the
various aspects of the transaction network, the account attributes and
various probability distributions, the input parameters are contained
in a config file called 'params.ini'.

In addition to the two main csv files, several other files and images
are generated containing description of the statistical properties of
the output dataset.


The schema of the transactions.csv is:

originRE
        dtype: string (three letter bank identifier)
        example: "CBA"
originID
        dtype: Integer
        example: 854768321582345
destRE
        dtype: string (three letter bank identifier)
        example: "NAB"
DestID
        dtype: Integer (6 digit BSB + 9 digit Acc Number)
        example: 591778636289458
amount
        dtype: string
        example: 136.82
datetime
        dtype: string (in YYYY-MM-DD HH:MM:SS format)
        example: "2020-06-16T14:38:23"

originRE    originID            destRE  DestID              amount      datetime
--------    ---------------     ------  ---------------     -------     -------------------
CBA         854768321582345     NAB     591778636289458     136.82      2020-06-16T14:38:23



The schema of the accounts.csv is:

RE
        dtype: string (three letter bank identifier)
        example: "CBA"
accountID
        dtype: Integer,
        example: 854768321582345
accountOpenDate
        dtype: string (YYYY-MM-DD format)
        example: 1990-05-11
accountType
            dtype: string (options, and their distribution are defined in params.ini)
            example: Business
custOccupation
            dtype: string (options, and their distribution are defined in params.ini)
            example: Student


RE      accountID           accountOpenDate     accountType     custOccupation
---     ----------------    ---------------     -----------     ---------------
CBA     854768321582345     1990-05-11          Business        Student

"""

import networkx as nx
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import powerlaw
from datetime import datetime
import sys
import itertools


def generate_transaction_network(num_nodes, m, p, num_remove_nodes, b=0, seed=None):
    """Generates a directed scale-free graph as the skeleton for
    generating domestic retail transactions.  The output graph is used
    in a later step for generating transactional data.

    Parameters
    ----------
    num_nodes : int

        the number of nodes

    m : int

        The graph is gown one node at a time.  Each time a node is
        added, m random edges are added between the newly added node
        and existing nodes.  Each edge is attached from the new node
        to an existing node with the probability proportional to its
        degree.  This step is known as preferential attachment, and
        leads to a power-law degree distribution.

    p : float

        The probability of performing the triad step in the Holme and
        Kim algorithm.  If in the preferential attachment step an edge
        is added from a new node (v) to an existing node (w), then in
        the triad step, one more edge is added from the new node (v)
        to a randomly chosen neighbour of w, call it u.  The triad
        step creates a "triangle" of edges connecting the new node (v)
        to both u and w, which were already neighbours.

    num_remove_nodes: int

        The number of high degree nodes to remove from graph (used to
        increase the average path length of the generated graph)

    b: float,

        The fraction of node-pairs (0<b<1) that do not have
        bi-directional edges.  When converting from an undirected
        graph to a directed graph, a fraction b/2 of edges are
        converted to (A->B) edges only, another fraction b/2 of edges
        are converted to (B->A) edges only, and a fraction (1-b) are
        converted to bi-directional (that is, both A->B and B->A).

    seed : integer, random_state, or None (default)

        Indicator of random number generation state.

    Notes
    -----
    Currently this function generates only a single "layer". In order
    to more accurately represent a real-world situation we may want to
    add different agents, and different layers (i.e. business,
    non-business, trusts).  For now we just use the Holme and Kim
    algorithm for growing graphs with powerlaw degree distribution
    (which can only generate degree distributions with exponent -3)
    and allows tuning of the average clustering. It can only generate
    undirected graphs however, and the slope (-3) is a bit high, so
    may be also worth tring one of the other scale-free graph
    generators.
    """
    G_undirected = nx.powerlaw_cluster_graph(num_nodes, m, p, seed)

    generate_account_numbers(G_undirected)

    # Now remove a few of the nodes with the highest degree to
    # increase the average path length between node pairs.  (This
    # might be something we'd do in the real world anyway, to avoid
    # connecting the whole graph.)

    nodes_sorted_by_degree = sorted(
        G_undirected.degree, key=lambda x: x[1], reverse=True)

    for i in range(num_remove_nodes):
        G_undirected.remove_node(nodes_sorted_by_degree[i][0])

    # Now convert G to a directed graph, making each edge into two
    # (bi-directional) edges.
    G = nx.DiGraph(G_undirected)

    # Then remove some of the directional edges, so as to make some
    # edges "one-way" rather than bi-directional (keeping at least one
    # directional edge) This is necessary because the
    # powerlaw_cluster_graph only generates undirected graphs, and
    # making some edges "one-way" may be more realistic and also
    # serves to increase the average path length

    if b > 1.0:
        raise ValueError(f'Expected 0 <= b <= 0.5 but it was {b}')

    for A, B in G_undirected.edges():

        rand = np.random.rand()
        if rand < b/2.0:
            G.add_edge(A, B)
        elif b/2.0 < rand < (1.0 - b/2.0):
            G.add_edge(A, B)
            G.add_edge(B, A)
        else:
            G.add_edge(B, A)

    return G


def generate_account_numbers(G):
    # Read params from config
    acc_num_len = config['Graph'].getint('account_num_len', fallback=13)
    
    # Generate 13-digit account numbers that fit in PostgreSQL bigint
    max_account_id = min(10**acc_num_len - 1, 9223372036854775807)  # PostgreSQL bigint max
    min_account_id = 10**(acc_num_len-1)  # Ensure 13 digits
    acc_num_sample = np.random.randint(
        min_account_id, high=(max_account_id + 1), size=G.number_of_nodes())
    
    label_mapping = {n: acc_num_sample[n] for n in G}
    nx.relabel_nodes(G, label_mapping, copy=False)
    
    # Generate 3-digit bank_ids deterministically per peer
    bank_id_map = {}
    for i, p in enumerate(peer_names):
        # Assign deterministic bank_id: PEER_1 = 001, PEER_2 = 002, etc.
        bank_id_base = (i + 1) * 100  # 100, 200, 300, 400
        bank_id_list = []
        for j in range(num_bank_ids_per_peer[p]):
            bank_id = bank_id_base + j + 1  # 101, 201, 301, 401 (if num_bank_ids=1)
            bank_id_list.append(bank_id)
        bank_id_map[p] = bank_id_list
    
    # Randomly assign a peer name for each graph node
    for n in G:
        # use str() to convert from np.str_ to str
        G.nodes[n]['Peer'] = str(np.random.choice(peer_names))
        G.nodes[n]['bank_id'] = np.random.choice(bank_id_map[G.nodes[n]['Peer']])


def generate_account_attributes(G, acc_type, acc_type_probs, cust_occupation, cust_occupation_probs):

    keys = list(G.nodes())
    
    # Generate account types
    # np.random.choice returns a list of np.str objects.
    acc_type_vals = np.random.choice(
        acc_type, size=G.number_of_nodes(), p=acc_type_probs)

    # The following converts np.str_ objects to regular python str
    # objects so we can use nx.write_graphml_lxml() to write out the
    # graph structure and attributes.
    acc_type_vals_str = [str(item) for item in acc_type_vals]
    acc_types = dict(zip(keys, acc_type_vals_str))
    nx.set_node_attributes(G, acc_types, 'acc_type')

    # Generate customer occupations
    occupation_vals = np.random.choice(
        cust_occupation, size=G.number_of_nodes(), p=cust_occupation_probs)
    # convert np.str_ to str
    occupation_vals_str = [str(item) for item in occupation_vals]
    cust_occupations = dict(zip(keys, occupation_vals_str))
    nx.set_node_attributes(G, cust_occupations, 'cust_occupation')

    # Generate Account Opening Date
    base_date = np.datetime64('1970-01-01')
    days_since_base_date = (np.datetime64('today') - base_date).astype(int)
    random_dates = (base_date + np.random.randint(
        0, days_since_base_date, size=G.number_of_nodes())).astype('datetime64[D]')
    # convert np.str_ to str
    random_dates_str = [str(item) for item in random_dates]
    random_dates_str_dict = dict(zip(keys, random_dates_str))
    nx.set_node_attributes(G, random_dates_str_dict, 'acc_open_date')

def generate_transactions(G, output_filename, avg_num_tx_per_pair, avg_tx_amount, sigma_tx_amount, start_date_str, date_range, start_time, end_time):
    '''
    Given a graph G, generate a multi-graph by creating multiple
    transactions per edge, and write the transactions to a csv file.
    '''

    with open(output_filename, 'w') as tx_file:
        tx_file.write("origin_re,origin_id,origin_bank_id,dest_re,dest_id,dest_bank_id,total_amount,transaction_date\n")

        # For each edge, generate a number of different transactions
        # and corresponding times, and write them to file
        for source, dest in G.edges():
            n_tx, tx_amounts = get_transaction_amounts(
                avg_num_tx_per_pair, avg_tx_amount, sigma_tx_amount)
            tx_dates = get_transaction_times(
                n_tx, start_date_str, date_range, start_time, end_time) 
            sourceRE = G.nodes[source]['Peer']
            sourceBankID = G.nodes[source]['bank_id']
            destRE = G.nodes[dest]['Peer']
            destBankID = G.nodes[dest]['bank_id']
            # write_transactions() will write n_tx transactions from
            # source to dest to the file-hanle tx_file
            write_transactions(
                source, dest, sourceBankID, destBankID, sourceRE, destRE, tx_amounts, tx_dates, tx_file)
        # Text files should end with a new line character
        tx_file.write('\n')


def write_transactions(source, dest, sourceBankID, destBankID, sourceRE, destRE, tx_amounts, tx_dates, tx_file):
    for tx_amount, tx_date in zip(tx_amounts, tx_dates):
        tx_file.write(sourceRE + ','
                      + str(int(source)) + ','
                      + str(int(sourceBankID)) + ','
                      + destRE + ','
                      + str(int(dest)) + ','
                      + str(int(destBankID)) + ','
                      + str(tx_amount) + ','
                      + np.datetime_as_string(tx_date, unit='D')
                      + '\n')


def write_accounts(G, dir_name, acc_type, acc_type_probs, cust_occupation, cust_occupation_probs):
    # Write out file containing account numbers for each RE
    files = {RE: open(dir_name + '/{}_accounts.csv'.format(RE), 'w')
             for RE in peer_names}
    files['all'] = open(dir_name + '/all_accounts.csv', 'w')

    header = "re,account_id,bank_id,account_open_date,account_type,cust_occupation\n"

    for f in files.values():
        f.write(header)

    for n in G:
        row = (G.nodes[n]['Peer'] + ','
               + str(int(n)) + ','
               + str(int(G.nodes[n]['bank_id'])) + ','
               + G.nodes[n]['acc_open_date'] + ','
               + G.nodes[n]['acc_type'] + ','
               + G.nodes[n]['cust_occupation'] + '\n')

        files['all'].write(row)

        RE = G.nodes[n]['Peer']
        files[RE].write(row)

    for f in files.values():
        f.write('\n')
        f.close()


def get_transaction_amounts(avg_num_tx_per_pair, avg_tx_amount, sigma_tx_amount):
    '''Returns an array of transaction amounts.

    The length of the array (that is, the number of distinct
    transactions to be created) is sampled from a Poisson distribution
    The transaction amounts themselves are sampled from a log-normal
    distribution.

    References: On the Distribution of Price and Quality, Alex Coad,
    2009, https://link.springer.com/article/10.1007/s00191-009-0142-z
    '''

    # Determine the number of transactions generate, by sampling from
    # a Poisson distribution
    n_tx = np.random.poisson(lam=avg_num_tx_per_pair)

    # This is just to make sure all edges get at least one transaction...
    n_tx = n_tx + 1

    # For each transaction, sample the transaction amount from a
    # lognormal distribution
    tx_amounts = np.random.lognormal(
        mean=np.log(avg_tx_amount), sigma=np.log(sigma_tx_amount), size=n_tx)

    return n_tx, tx_amounts


def get_transaction_times(n_tx, start_date_str, date_range, start_time, end_time):
    # For each transaction, sample the transaction date/time from a
    # uniform distribution over N business days (with times between
    # 9am - 5pm).
    start_date = np.datetime64(start_date_str)
    date_deltas = np.random.randint(0, date_range, n_tx)
    time_deltas = np.random.randint(
        start_time, end_time, n_tx).astype('timedelta64[s]')
    datetimes = np.busday_offset(
        start_date, date_deltas, roll='forward') + time_deltas

    return datetimes

def print_graph_info(G, dir_name):
    with open(dir_name + "/graph_info.txt", 'w') as info_file:
        '''
        info_file.write(nx.info(G) + '\n')
        info_file.write('Graph Density: ' + str(nx.density(G)) + '\n')
        info_file.write('Average Clustering Coefficient: '
                        + str(nx.average_clustering(G)) + '\n')
        '''
        # nx.info alternative - temporary fix
        info_file.write(f"Name: {G.name}\n")
        info_file.write(f"Nodes: {G.number_of_nodes()}\n")
        info_file.write(f"Edges: {G.number_of_edges()}\n")
        info_file.write(f"Density: {nx.density(G)}\n")
        info_file.write(f"Avg clustering: {nx.average_clustering(G)}\n")

        # Get degree distribution and power-law indexes
        degrees = [G.degree(n) for n in G.nodes()]
        fit = powerlaw.Fit(degrees)

        info_file.write('Power law exponent of degree distribution: '
                        + str(fit.power_law.alpha) + '\n')

    # Now plot the power law (and best fit)
    # --- First, set figure parameters
    plt.rc('font', family='serif')
    plt.figure(figsize=(8, 6))
    plt.ylabel("p(degree)", fontsize=18)
    plt.xlabel("Degree", fontsize=18)
    plt.title("Degree Distribution with Powerlaw Fit", fontsize=18)
    plt.tick_params(axis="both", which="major", labelsize=14)
    plt.tick_params(axis="both", which="minor", labelsize=13)
    plt.tight_layout()

    # --- Now do the plotting
    fig = fit.plot_pdf(color='b', linewidth=2)
    fit.power_law.plot_pdf(color='b', linestyle='--', ax=fig)

    plt.annotate("Powerlaw exponent = " + str(fit.power_law.alpha),
                 (0.1, 0.1), xycoords='axes fraction')

    plt.savefig(dir_name + "/degree_distribution.pdf")
    plt.show()


def get_config_items_list(config, section, key):
    """Function to split comma delimeted string returned from
    config.get() into a list of elements.
    """
    itms = config[section].get(key).split(',')
    itms_list = [item.strip() for item in itms]
    return itms_list


if __name__ == '__main__':
    """Create a set of fake transactions to be used for dev/test work on
    the FinTracer Prototype.

    Also calculate some descriptive statistics for the synthesized
    dataset and store in the same location as the dataset itself.
    """

    import configparser
    import os.path
    import shutil

    if len(sys.argv) == 2:
        config_file = sys.argv[1]
    else:
        print("\nError: You must include a path to a config file. \n\n  -- Usage: " + sys.argv[0] + " <path/to/config/file>.\n\n")
        sys.exit()

    if not os.path.isfile(config_file):
        print("\n\n Error: The config file '" + config_file + "' does not exist. Exiting program.\n\n")
        sys.exit()

    # Read parameters from Config File...

    config = configparser.ConfigParser()

    config.read(config_file)

    num_nodes = config['Graph'].getint('num_nodes', fallback=1000)
    m = config['Graph'].getint('m', fallback=15)
    p = config['Graph'].getfloat('p', fallback=0.3)
    b = config['Graph'].getfloat('b', fallback=0.1)
    num_remove_nodes = config['Graph'].getint('num_remove_nodes', fallback=5)
    seed = config['Graph'].getint('seed', fallback=123)
    peer_names = get_config_items_list(config, 'Graph', 'peer_names')
    num_bank_ids = get_config_items_list(config, 'Graph', 'num_bank_ids')
    num_bank_ids_per_peer = {}
    if len(num_bank_ids) == 1:
        num_bank_ids_per_peer = {p: int(num_bank_ids[0]) for p in peer_names}
    elif len(num_bank_ids) == len(peer_names):
        num_bank_ids_per_peer = {p: int(num_bank_ids[i]) for i, p in enumerate(peer_names)}
    else:
        raise ValueError("num_bank_ids should have a length equal to the number of peers or a single value")
    
    avg_num_tx_per_pair = config['TransactionsPerEdge'].getfloat('avg_num_tx_per_pair', fallback=10)

    avg_tx_amount = config['TransactionAmount'].getfloat('avg_tx_amount', fallback=200.0)
    sigma_tx_amount = config['TransactionAmount'].getfloat('sigma_tx_amount', fallback=4)

    start_date_str = config['TransactionDatetime']['start_date_str']
    date_range = config['TransactionDatetime'].getint('date_range', fallback=10)
    start_time = config['TransactionDatetime'].getint('start_time', fallback=32400)
    end_time = config['TransactionDatetime'].getint('end_time', fallback=61200)


    acc_type = get_config_items_list(config, 'AccountAttributes', 'acc_type')
    acc_type_probs = get_config_items_list(config, 'AccountAttributes', 'acc_type_probs')
    cust_occupation = get_config_items_list(config, 'AccountAttributes', 'cust_occupation')
    cust_occupation_probs = get_config_items_list(config, 'AccountAttributes', 'cust_occupation_probs')

    print(f'num_nodes: {num_nodes}')
    print(f'm: {m}')
    print(f'p: {p}')
    print(f'b: {b}')
    print(f'num_remove_nodes: {num_remove_nodes}')
    print(f'seed: {seed}')
    print(f'avg_num_tx_per_pair: {avg_num_tx_per_pair}')
    print(f'avg_tx_amount: {avg_tx_amount}')
    print(f'sigma_tx_amount: {sigma_tx_amount}')
    print(f'start_date_str: {start_date_str}')
    print(f'date_range: {date_range}')
    print(f'start_time: {start_time}')
    print(f'end_time: {end_time}')
    print(f'num_bank_ids_per_peer: {num_bank_ids_per_peer}')

    # Remove this hardcoded override - it's causing the inconsistency
    # peer_names = ['PEER_1', 'PEER_2', 'PEER_3', 'PEER_4']

    # Set the random seed for numpy.random functions
    np.random.seed(seed)

    # Create transaction network via graph generator
    G = generate_transaction_network(num_nodes, m, p, num_remove_nodes, b, seed=seed)
    generate_account_attributes(G, acc_type, acc_type_probs, cust_occupation, cust_occupation_probs)

    # Generate output filename and path in which to store the synthetic data
    # (Note: this assumes the directory synthesised_data exists.)
    time_now = datetime.today().strftime('%Y-%m-%d-%H-%M-%S')

    if not os.path.exists("synthesised_data"):
        os.mkdir("synthesised_data")

    dir_name = "synthesised_data/transactions_" + time_now

    if not os.path.exists(dir_name):
        os.mkdir(dir_name)

    output_filename = dir_name + "/all_transactions.csv"

    # Create transactions (this function writes transactions to a file)
    generate_transactions(G, output_filename, avg_num_tx_per_pair, avg_tx_amount, sigma_tx_amount, start_date_str, date_range, start_time, end_time)

    # Split into individual Peer transaction datasets
    for p in peer_names:
        peer_output_filename = dir_name + f"/{p}_transactions.csv"

        os.system(f"echo 'origin_re,origin_id,origin_bank_id,dest_re,dest_id,dest_bank_id,total_amount,transaction_date' > {peer_output_filename}")
        os.system(f"grep {p} {output_filename} >> {peer_output_filename}")

    write_accounts(G, dir_name, acc_type, acc_type_probs, cust_occupation, cust_occupation_probs)

    # copy the config file (passed on the command line when launching the script) to this directory as well,
    # so that we can keep track of what config parameters were used to create the specific dataset...
    shutil.copy2(sys.argv[1], dir_name)

    # Now read in transactions.csv and plot distributions

    df = pd.read_csv(output_filename)

    bin_edges = np.logspace(-1.0, 3.5, 300)

    df['total_amount'].astype(float).plot.hist(bins=bin_edges, alpha=0.5)

    plt.title("Distribution of transaction amounts.")
    plt.xlabel("Transaction Amount (Dollars).")

    plt.savefig(dir_name + "/transaction_amount_distribution.pdf")

    # Generate plot of shortest path length for randomly selected node pairs

    path_lengths = dict(nx.shortest_path_length(G))

    all_pairs_shortest_path_lengths = list(itertools.chain.from_iterable(d.values() for d in path_lengths.values()))

    bin_edges = np.linspace(0.5, max(all_pairs_shortest_path_lengths) + 0.5, max(all_pairs_shortest_path_lengths) + 1)

    plt.figure(figsize=(9, 6))
    plt.hist(all_pairs_shortest_path_lengths, bins=bin_edges, weights = np.ones_like(all_pairs_shortest_path_lengths)/float(len(all_pairs_shortest_path_lengths)))

    plt.xlabel("Shortest Path Length (Number of hops)", fontsize=18)
    plt.ylabel("Fraction of node-pairs per bin", fontsize=18)
    plt.title("Distribution of shortest path lengths", fontsize=18)
    plt.xticks(fontsize=18)
    plt.yticks(fontsize=18)

    plt.savefig(dir_name + "/shortest_path_lengths_distribution.pdf")

    # Save graph in graphml format so it can be visualised with eg. Gephi

    nx.write_graphml_lxml(G, dir_name + "/graph.graphml")

    # Generate some descriptive stats about the generated data, and put it in the directory...

    print_graph_info(G, dir_name)
    print_graph_info(G, dir_name)
