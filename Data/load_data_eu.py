# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from sqlalchemy import create_engine
import os
from os.path import join, dirname, isfile
from dotenv import load_dotenv
from time import time

if __name__ == "__main__":

    # Start timer
    start = time()

    # Util functions
    def get_csv_header(fn):
        with open(fn) as f:
            return ','.join([f'"{i}"' for i in f.readline().rstrip('\n').split(",")])

    # Load node names from .env file
    dotenv_path = join(dirname(__file__), '..', '.env')
    load_dotenv(dotenv_path)

    node_names = []
    node_schemas = []

    node_count = sum(1 for i in os.environ.keys() if 'DB_SCHEMA' in i)

    # Get config items for each node
    for i in range(node_count):
        try:
            node_names.append(os.environ[f"N{i}_NAME"])
            node_schemas.append(os.environ[f"N{i}_DB_SCHEMA"])
        except KeyError as ex:
            print(f"config missing item with error: {ex}")
            exit(0)

    print(f"Node Names: {node_names},\nNode Schema Names: {node_schemas}")

    # Config
    # TODO: Need more secure mechanism
    username = os.environ.get("DB_DEFAULT_USER", "postgres")
    password = os.environ.get("DB_DEFAULT_PW", "postgres")
    server = os.environ.get("DB_HOST", "localhost")
    port = os.environ.get("DB_LOCAL_PORT", "5432")
    address = f"{server}:{port}"
    database = os.environ.get("DB_DEFAULT_DB", "ftillite")
    all_subdirs = [d for d in os.listdir('synthesised_data')]
    csv_loc = f'synthesised_data/{sorted(all_subdirs)[-1]}' # Get the latest generated data

    engine = create_engine(f'postgresql://{username}:{password}@{address}/{database}')

    # Load transcations
    data_files = [join(csv_loc, f) for f in os.listdir(csv_loc) if isfile(join(csv_loc, f)) and 'transactions' in f and 'tmp' not in f and 're_copy' not in f]

    csv_header = get_csv_header(data_files[0])

    # Currently need this for csv files
    def remove_blanklines_from_file(file_loc, clean_file_loc):
        with open(file_loc,'r') as fr, open(clean_file_loc,'w') as fw:
            for line in fr:
                if not line.strip(): continue
                fw.write(line)

    print("Data Loader Started")
    # Start timer
    start = time()

    conn = engine.raw_connection()
    cursor = conn.cursor()


    for idx, csv_file_path in enumerate(data_files):
        # csv_file_path = f'{csv_loc}/transactions.csv'
        clean_csv_file_path = csv_file_path.replace('.csv', '_tmp.csv')
        remove_blanklines_from_file(csv_file_path, clean_csv_file_path)
        with open(clean_csv_file_path, 'r') as f:    
            cmd = f'COPY transactions({csv_header}) FROM STDIN WITH (FORMAT CSV, HEADER TRUE)'
            cursor.copy_expert(cmd, f)
            conn.commit()
        print(f"Transaction Data File - {idx+1} of {len(data_files)} Loaded")

    print("Tranaction Data Loaded")

    # Load accounts
    csv_file_path = f'{csv_loc}/all_accounts.csv'
    clean_csv_file_path = f'{csv_loc}/clean_all_accounts.csv'
    remove_blanklines_from_file(csv_file_path, clean_csv_file_path)

    csv_header = get_csv_header(clean_csv_file_path)

    with open(clean_csv_file_path, 'r') as f:    
        cmd = f'COPY accounts({csv_header}) FROM STDIN WITH (FORMAT CSV, HEADER TRUE)'
        cursor.copy_expert(cmd, f)
        conn.commit()
    print("Account Data Loaded")

    # Refresh materialized views
    for s in node_schemas:
        cursor.execute(f"REFRESH MATERIALIZED VIEW {s}.accounts")
        cursor.execute(f"REFRESH MATERIALIZED VIEW {s}.transactions")
        cursor.execute(f"REFRESH MATERIALIZED VIEW {s}.edges")  # Add this line
        conn.commit()  # Add commit after each refresh

    conn.close()
    end = time()
    print(f"Data Loader Finished in {end - start} seconds")