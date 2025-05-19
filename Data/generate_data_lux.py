"""Script to generate synthetic Luxembourg transaction data (using IBAN) to assist
with FinTracer prototype development - INCERT ID Fraud Detection project.

    - Modified functions are marked with "---- Modified ----"

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
origin_re,origin_iban,dest_re,dest_iban,amount,datetime


The schema of the transactions.csv is:

originRE
        dtype: string (bank identifier)
        example: "Peer_1"
originIBAN
        dtype: string (IBAN) 
        example: LU31 0011 2345 6789 0123
destRE
        dtype: string (peer identifier)
        example: "PEER_1"
DestIBAN
        dtype: string 
        example: LU31 0011 2345 6789 0123
amount
        dtype: string
        example: 136.82
datetime
        dtype: string (in YYYY-MM-DD HH:MM:SS format)
        example: "2020-06-16T14:38:23"

originRE    originIBAN                   destRE      DestIBAN                     amount      datetime
--------    ------------------------     ------      ------------------------     -------     -------------------
PEER_1      LU31 0011 2345 6789 0123     PEER_1      LU31 0011 2345 6789 0123     136.82      2020-06-16T14:38:23



The schema of the accounts.csv is:

PEER
        dtype: string (peer identifier)
        example: "PEER_1"
IBAN
        dtype: string
        example: LU31 0011 2345 6789 0123
accountOpenDate
        dtype: string (YYYY-MM-DD format)
        example: 1990-05-11
accountType
            dtype: string (options, and their distribution are defined in params.ini)
            example: Business
custOccupation
            dtype: string (options, and their distribution are defined in params.ini)
            example: Student


RE         IBAN                         accountOpenDate     accountType     custOccupation
------     ------------------------     ---------------     -----------     ---------------
PEER_1     LU31 0011 2345 6789 0123     1990-05-11          Business        Student

"""

import networkx as nx
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import powerlaw
from datetime import datetime
import sys
import itertools

import configparser
import os
import shutil
import sys 

# ---- Modified ----
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

    #generate_account_numbers(G_undirected) - in main function

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
    G = nx.DiGraph()

    if b > 1.0:
        raise ValueError(f"Expected 0 <= b <= 1.0, but got {b}")

    for A, B in G_undirected.edges():
        
        rand = rng.random()

        if rand < b/2.0:
            G.add_edge(A, B)
        elif rand < (1 - b/2.0): 
            G.add_edge(A, B)
            G.add_edge(B, A)
        else:
            G.add_edge(B, A)
    
    return G


# ---- Modified ----
def generate_account_numbers(G):
    # Read IBAN params from config
    # IBAN form: LUkkBBBCCCCCCCCCCCCC
    chk_len     = config['Graph'].getint('check_digits_len')
    bc_len      = config['Graph'].getint('bank_code_len')
    acc_len     = config['Graph'].getint('acc_num_len')
    country     = config['Graph']['country_code']

    # Peers (banks)
    peer_names  = [p.strip() for p in config['Graph']['peer_names'].split(',')]

    # We now assign _one_ unique bank code per peer, e.g. "001","002",…
    # (zero-padded to bc_len).  Enumerate peers starting at 1.
    bank_code_map = {
        peer: f"{idx:0{bc_len}d}"
        for idx, peer in enumerate(peer_names, start=1)
    }
    

    # Potential Bug
    for n in G.nodes():
        # Randomly pick which peer (bank) this node belongs to
        peer = str(rng.choice(peer_names))
        G.nodes[n]['Peer'] = peer

        # Use that peer's one-and-only bank code
        bc = bank_code_map[peer]

        # Random 2-digit check digits
        # IRL: Are NOT random, but in our case these are irrelevant
        chk = rng.integers(10**(chk_len-1), 10**chk_len)

        # Random acc_num_len-digit account number
        acc = rng.integers(10**(acc_len-1), 10**acc_len)

        # Assemble the IBAN
        iban = str(f"{country}{chk:0{chk_len}d}{bc}{acc:0{acc_len}d}")
        G.nodes[n]['IBAN'] = iban


# ---- Not Modified ----
def generate_account_attributes(G, acc_type, acc_type_probs, cust_occupation, cust_occupation_probs):
    
    keys = list(G.nodes())
    
    # Generate account types
    # rng.choice returns a list of np.str object
    acc_type_vals = rng.choice(
        acc_type, size=G.number_of_nodes(), p=acc_type_probs)

    # Convert np.str_ objects to native Python str
    # nx.write_graphml_lxml() can export the graph and its attributes
    acc_type_vals_str = [str(item) for item in acc_type_vals]
    acc_types = dict(zip(keys, acc_type_vals_str))
    nx.set_node_attributes(G, acc_types, 'acc_type')

    # Generate customer occupations
    occupation_vals = rng.choice(
        cust_occupation, size=G.number_of_nodes(), p=cust_occupation_probs)
    # convert np.str_ to str
    occupation_vals_str = [str(item) for item in occupation_vals]
    cust_occupations = dict(zip(keys, occupation_vals_str))
    nx.set_node_attributes(G, cust_occupations, 'cust_occupation')

    # Generate Account Opening Date
    base_date = np.datetime64('1970-01-01')
    days_since_base_date = (np.datetime64('today') - base_date).astype(int)
    random_dates = (base_date + rng.integers(
        0, days_since_base_date, size=G.number_of_nodes())).astype('datetime64[D]')
    
    # convert np.str_ to str
    random_dates_str = [str(item) for item in random_dates]
    random_dates_str_dict = dict(zip(keys, random_dates_str))
    nx.set_node_attributes(G, random_dates_str_dict, 'acc_open_date')
 

# ---- New: assign three additional flags to each account ----
def assign_extra_flags(G):
    """
    Randomly tag each node with:
      
      - criminal_affiliation_flag (bool)
      - id_fraud_victim_flag      (bool)
      - watchlist_flag            (bool)

    using probabilities from the config file.
    """
    # read the probs from config
    crim_p  = config['AccountAttributes'].getfloat('criminal_affiliation_prob')
    fraud_p = config['AccountAttributes'].getfloat('id_fraud_victim_prob')
    watchlist_prob = config['AccountAttributes'].getfloat('watchlist_prob')
    
    # prepare node list
    nodes = list(G.nodes())
    n = len(nodes)
    
    # draw independent bernoulli trials per node
    crim_flags  = rng.random(n) < crim_p
    fraud_flags = rng.random(n) < fraud_p
    watchlist_flags = rng.random(n) < watchlist_prob

    crim_flags_str = [str(item) for item in crim_flags]
    fraud_flags_str = [str(item) for item in fraud_flags]
    watchlist_flags_str = [str(item) for item in watchlist_flags]
    
    # build dicts and set as node attributes
    crim_dict  = dict(zip(nodes, crim_flags_str))
    fraud_dict = dict(zip(nodes, fraud_flags_str))
    watchlist_dict = dict(zip(nodes, watchlist_flags_str))
    
    nx.set_node_attributes(G, crim_dict, 'criminal_affiliation_flag')
    nx.set_node_attributes(G, fraud_dict, 'id_fraud_victim_flag')
    nx.set_node_attributes(G, watchlist_dict, 'watchlist_flag')

    # Easily extendable to add more flags in the future


# ---- Modified ----
def generate_transactions(G, output_filename, avg_num_tx_per_pair,
                           avg_tx_amount, sigma_tx_amount,
                           start_date_str, date_range, start_time, end_time):
    '''
    Given a graph G, generate a multi-graph by creating multiple
    transactions per edge, and write the transactions to a csv file.
    '''
    
    with open(output_filename, 'w') as tx_file:
        tx_file.write("origin_re,origin_iban,dest_re,dest_iban,total_amount,transaction_date\n")
        
        #rng = np.random.default_rng(config['Graph'].getint('seed'))
        #base = np.datetime64(start_date_str)

        # For each edge, generate a number of different transactions
        # and corresponding times, and write them to file
        for source, dest in G.edges():
            n_tx, tx_amounts = get_transaction_amounts(
                avg_num_tx_per_pair, avg_tx_amount, sigma_tx_amount)
            tx_dates = get_transaction_times(
                n_tx, start_date_str, date_range, start_time, end_time) 
    
            sourceRE = G.nodes[source]['Peer']
            sourceIBAN = G.nodes[source]['IBAN']
            destRE = G.nodes[dest]['Peer']
            destIBAN = G.nodes[dest]['IBAN']
            
            """# number & amounts
            n_tx = rng.poisson(avg_num_tx_per_pair) + 1
            tx_amounts = rng.lognormal(mean=np.log(avg_tx_amount), sigma=np.log(sigma_tx_amount), size=n_tx)
            tx_dates   = (base
                          + rng.integers(0, date_range, size=n_tx).astype('timedelta64[D]')
                          + rng.integers(start_time, end_time, size=n_tx).astype('timedelta64[s]'))
            src = G.nodes[source]; dst = G.nodes[dest]

            write_transactions(source, dest,
                               src['Peer'], src['IBAN'],
                               dst['Peer'], dst['IBAN'],
                               tx_amounts, tx_dates, tx_file)"""
            
            # write_transactions() will write n_tx transactions from
            # source to dest to the file-hanle tx_file
            write_transactions(
                source, dest, sourceRE, sourceIBAN, destRE, 
                destIBAN, tx_amounts, tx_dates, tx_file)

        # Text files should end with a new line character
        tx_file.write('\n')


# ---- New Function ----  Simulate risk scores based on account attributes and activity --------- Could be buggy, shouldnt iterate over G.nodes(), can cause problems
def generate_risk_scores(G, tx_df, today=None):

    if today is None:
        today = np.datetime64('today')

    # Precompute age normalization ---
    open_dates = np.array([np.datetime64(G.nodes[n]['acc_open_date'])
                           for n in G.nodes()])
    ages_days = (today - open_dates).astype(int)
    max_age = ages_days.max() or 1

    # Build per-IBAN metrics from the DataFrame 
    grp = (
        tx_df
        .groupby('origin_iban')['total_amount']
        .agg(tx_count='count', avg_amt='mean')
    )
    max_tx = grp['tx_count'].max() or 1
    max_amt = grp['avg_amt'].max() or 1

    # Assign risk per node
    for n in G.nodes():
        iban = G.nodes[n]['IBAN']
        pep = 1.0 if G.nodes[n]['cust_occupation']=='PEP' else 0.0

        # normalized age
        age_days = (today - np.datetime64(G.nodes[n]['acc_open_date'])).astype(int)
        age_norm = age_days / max_age

        # lookup tx metrics (0 if missing)
        if iban in grp.index:
            vel_norm = grp.loc[iban, 'tx_count'] / max_tx
            amt_norm = grp.loc[iban, 'avg_amt'] / max_amt
        else:
            vel_norm = 0.0
            amt_norm = 0.0

        noise = rng.random()

        score = 0.20*pep + 0.10*age_norm + 0.15*amt_norm \
                + 0.15*vel_norm + 0.40*noise

        G.nodes[n]['risk_score']      = str(score)
        G.nodes[n]['high_risk_flag']  = bool(score > 0.45)


# ---- Modified ----
def write_transactions(source, dest,
                       sourceRE, sourceIBAN,
                       destRE, destIBAN,
                       tx_amounts, tx_dates, tx_file):
    
    for tx_amount, tx_date in zip(tx_amounts, tx_dates):
        tx_file.write(sourceRE + ',' 
                    #  + str(int(source)) + ','
                      + str(sourceIBAN) + ','
                      + destRE + ','
                    #  + str(int(dest)) + ','
                      + str(destIBAN) + ','
                      + str(tx_amount) + ','
                      + np.datetime_as_string(tx_date, unit='D')
                        + '\n')


# ---- Modified ----
def write_accounts(G, dir_name, acc_type, acc_type_probs,
                   cust_occ, cust_occ_probs):
    # Write out file containing account numbers for each RE
    files = {RE: open(dir_name + '/{}_accounts.csv'.format(RE), 'w') 
             for RE in peer_names}
    files['all'] = open(dir_name + '/all_accounts.csv', 'w')

    header = "re,iban,acc_open_date,acc_type,cust_occupation,watchlist_flag, criminal_affiliation_flag, id_fraud_victim_flag, risk_score,high_risk_flag\n"

    for f in files.values(): 
        f.write(header)

    for n in G:
        row = (G.nodes[n]['Peer'] + ','
            #   + str(int(n)) + ','
               + str(G.nodes[n]['IBAN']) + ','
               + G.nodes[n]['acc_open_date'] + ','
               + G.nodes[n]['acc_type'] + ','
               + G.nodes[n]['cust_occupation'] + ','
               + str(G.nodes[n]['watchlist_flag']) + ','
               + str(G.nodes[n]['criminal_affiliation_flag']) + ','
               + str(G.nodes[n]['id_fraud_victim_flag']) + ','
               + G.nodes[n]['risk_score'] + ','
               + str(G.nodes[n]['high_risk_flag']) + "\n"
               )
        
        # Write the row to the file for this peer, and the all peers
        files['all'].write(row)
        files[G.nodes[n]['Peer']].write(row)

    for f in files.values():
        f.write('\n')
        f.close()


# ---- Not Modified ----
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
    n_tx = rng.poisson(lam=avg_num_tx_per_pair)

    # This is just to make sure all edges get at least one transaction...
    n_tx = n_tx + 1

    # For each transaction, sample the transaction amount from a
    # lognormal distribution
    tx_amounts = rng.lognormal(
        mean=np.log(avg_tx_amount), sigma=np.log(sigma_tx_amount), size=n_tx)

    return n_tx, tx_amounts


# ---- Not Modified ----
def get_transaction_times(n_tx, start_date_str, date_range, start_time, end_time):
    # For each transaction, sample the transaction date/time from a
    # uniform distribution over N business days (with times between
    # 9am - 5pm).
    start_date = np.datetime64(start_date_str)
    date_deltas = rng.integers(0, date_range, n_tx)
    time_deltas = rng.integers(
        start_time, end_time, n_tx).astype('timedelta64[s]')
    datetimes = np.busday_offset(
        start_date, date_deltas, roll='forward') + time_deltas

    return datetimes


# ---- Modified ---- nx.info deprecated? - wouldn't work
def print_graph_info(G, dir_name):
    with open(dir_name + "/graph_info.txt", 'w') as info_file:
        """info_file.write(nx.info(G) + '\n')
        info_file.write('Graph Density: ' + str(nx.density(G)) + '\n')
        info_file.write('Average Clustering Coefficient: '
                        + str(nx.average_clustering(G)) + '\n')
        """
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

    plt.savefig(dir_name + "/degree_distribution.png", format='png', dpi=500, bbox_inches='tight')
    plt.show()


# ---- Not Modified ----
def get_config_items_list(config, section, key):
    """Function to split comma delimeted string returned from
    config.get() into a list of elements.
    """
    itms = config[section].get(key).split(',')
    itms_list = [item.strip() for item in itms]
    return itms_list



# ---- Modified ----
if __name__ == '__main__':
    """Create a set of fake transactions to be used for dev/test work on
    the FinTracer Prototype.

    Also calculate some descriptive statistics for the synthesized
    dataset and store in the same location as the dataset itself.
    """

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
    
    with open(sys.argv[1], 'r', encoding='utf-8') as f:
        config.read_file(f)

    # Graph params
    num_nodes   = config['Graph'].getint('num_nodes', fallback =300)
    m           = config['Graph'].getint('m', fallback=15)
    p           = config['Graph'].getfloat('p', fallback=0.3)
    b           = config['Graph'].getfloat('b', fallback=0.1)
    num_remove_nodes  = config['Graph'].getint('num_remove_nodes', fallback=5)
    seed        = config['Graph'].getint('seed', fallback=123)
    peer_names  = get_config_items_list(config, 'Graph', 'peer_names')
    
    #[s.strip() for s in config['Graph']['peer_names'].split(',')]
    
    # Transaction params
    avg_num_tx_per_pair = config['TransactionsPerEdge'].getfloat('avg_num_tx_per_pair', fallback=10)
    avg_tx_amount       = config['TransactionAmount'].getfloat('avg_tx_amount', fallback=200.0)
    sigma_tx_amount     = config['TransactionAmount'].getfloat('sigma_tx_amount', fallback=4)
    
    start_date_str      = config['TransactionDatetime']['start_date_str']
    date_range          = config['TransactionDatetime'].getint('date_range', fallback=10)
    start_time          = config['TransactionDatetime'].getint('start_time', fallback=32400)
    end_time            = config['TransactionDatetime'].getint('end_time', fallback=61200)
    # Account attributes
    acc_type            = get_config_items_list(config, 'AccountAttributes', 'acc_type')
    acc_type_probs      = get_config_items_list(config, 'AccountAttributes', 'acc_type_probs')
    cust_occ            = get_config_items_list(config, 'AccountAttributes', 'cust_occupation')
    cust_occ_probs      = get_config_items_list(config, 'AccountAttributes', 'cust_occupation_probs')

    # Print out the parameters used to generate the data in terminal
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

    # Set the random seed for numpy.random functions - to ensure consistent reproducible results across runs 
    rng = np.random.default_rng(config['Graph'].getint('seed'))

    # Create transaction network via graph generator
    G = generate_transaction_network(num_nodes, m, p, num_remove_nodes, b, seed)
    
    generate_account_numbers(G)

    generate_account_attributes(G, acc_type, acc_type_probs, cust_occ, cust_occ_probs)

    assign_extra_flags(G)

    # Generate output filename and path in which to store the synthetic data
    # (Note: this assumes the directory synthesised_data exists.)
    time_now = datetime.now().strftime('%Y-%m-%d-%H-%M-%S')

    if not os.path.exists("synthesised_data"):
        os.mkdir("synthesised_data")

    dir_name = "synthesised_data/transactions_" + time_now

    if not os.path.exists(dir_name):
        os.mkdir(dir_name)

    output_filename = dir_name + "/all_transactions.csv"

    # Create transactions (this function writes transactions to a file)
    generate_transactions(G, output_filename,
                          avg_num_tx_per_pair,
                          avg_tx_amount, sigma_tx_amount,
                          start_date_str, date_range,
                          start_time, end_time)
    
    # Split into individual Peer transaction datasets
    for p in peer_names:
        #peer_output_filename = dir_name + f"/{p}_transactions.csv"
        peer_output_filename = dir_name + f"\\{p}_transactions.csv"

        os.system(f"echo 'origin_re,origin_iban,dest_re,dest_iban,total_amount,transaction_date' > {peer_output_filename}")
        os.system(f"grep {p} {output_filename} >> {peer_output_filename}")

    df = pd.read_csv(output_filename)
    df['total_amount'] = df['total_amount'].astype(float)   # pandas df for risk computation

    # simulate risk
    generate_risk_scores(G, df) # nodes have 'risk_score' ∈ [0,1] and a boolean 'high_risk_flag'
    
        
    write_accounts(G, dir_name,
                   acc_type, acc_type_probs,
                   cust_occ, cust_occ_probs)

    # copy the config file (passed on the command line when launching the script) to this directory as well,
    # so that we can keep track of what config parameters were used to create the specific dataset...
    shutil.copy2(sys.argv[1], dir_name)
    print(f"Generated data in {dir_name}")

    # Now read in transactions.csv and plot distributions

    df = pd.read_csv(output_filename)

    bin_edges = np.logspace(-1.0, 3.5, 300)

    df['total_amount'].astype(float).plot.hist(bins=bin_edges, alpha=0.5)

    plt.title("Distribution of Transaction Amounts")
    plt.xlabel("Transaction Amount (Euros)")

    plt.savefig(dir_name + "/transaction_amount_distribution.png", format='png', dpi=500, bbox_inches='tight')
    plt.close()

    # Generate plot of shortest path length for randomly selected node pairs

    path_lengths = dict(nx.shortest_path_length(G))

    all_pairs_shortest_path_lengths = list(itertools.chain.from_iterable(d.values() for d in path_lengths.values()))

    bin_edges = np.linspace(0.5, max(all_pairs_shortest_path_lengths) + 0.5, max(all_pairs_shortest_path_lengths) + 1)

    plt.figure(figsize=(9, 6))
    plt.hist(all_pairs_shortest_path_lengths, bins=bin_edges, 
             weights = np.ones_like(all_pairs_shortest_path_lengths)/float(len(all_pairs_shortest_path_lengths)))

    plt.xlabel("Shortest Path Length (Number of hops)", fontsize=18)
    plt.ylabel("Fraction of node-pairs per bin", fontsize=18)
    plt.title("Distribution of shortest path lengths", fontsize=18)
    plt.xticks(fontsize=18)
    plt.yticks(fontsize=18)

    plt.savefig(dir_name + "/shortest_path_lengths_distribution.png", format='png', dpi=500, bbox_inches='tight')
    plt.close()
    # Save graph in graphml format so it can be visualised with eg. Gephi

    nx.write_graphml_lxml(G, dir_name + "/graph.graphml")

    # Generate some descriptive stats about the generated data, and put it in the directory...

    print_graph_info(G, dir_name)

    scores = [
    float(G.nodes[n]['risk_score'])
    for n in G.nodes()
]

# Plot histogram of risk scores
plt.figure(figsize=(8, 6))
plt.hist(scores, bins=50)             
plt.xlabel('Risk Score')
plt.ylabel('Number of Accounts')
plt.title('Distribution of Account Risk Scores')
plt.tight_layout()

# Save to file
chart_path = os.path.join(dir_name, "risk_score_distribution.png")
plt.savefig(chart_path, format='png', dpi=500, bbox_inches='tight')
plt.close()
print(f"Saved risk‐score distribution chart to {chart_path}")
