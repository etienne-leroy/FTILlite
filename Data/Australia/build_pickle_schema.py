# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

# from sqlalchemy import Column, Integer, String, Date, Numeric, DateTime, MetaData, Table
from sqlalchemy import create_engine, inspect, schema
from sqlalchemy.sql import text
from sqlalchemy.orm import declarative_base  # Updated import
import os
from os.path import join, dirname, isfile
from dotenv import load_dotenv

dotenv_path = join(dirname(__file__), '..', '.env')
load_dotenv(dotenv_path)

node_names = []
node_schemas = []
node_usernames = []
node_passwords = []

node_count = sum(1 for i in os.environ.keys() if 'DB_SCHEMA' in i)

# Get config items for each node
for i in range(node_count):
    try:
        node_names.append(os.environ[f"N{i}_NAME"])
        node_schemas.append(os.environ[f"N{i}_DB_SCHEMA"])
        node_usernames.append(os.environ[f"N{i}_DB_USER"])
        node_passwords.append(os.environ[f"N{i}_DB_PW"])
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
# all_subdirs = [d for d in os.listdir('synthesised_data')]
# csv_loc = f'synthesised_data/{sorted(all_subdirs)[-1]}' # Get the latest generated data
engine = create_engine(f'postgresql://{username}:{password}@{address}/{database}')

with engine.connect() as connection:
    for idx, s in enumerate(node_schemas):
        s_lc = s.lower()
        n_lc = node_usernames[idx].lower()

        # Pickle table
        connection.execute(text(f'''CREATE TABLE IF NOT EXISTS {s_lc}.pickle
            (
                destination character varying COLLATE pg_catalog."default",
                dtype character varying(20) COLLATE pg_catalog."default",
                handle character varying(20) COLLATE pg_catalog."default",
                opcode character varying(10) COLLATE pg_catalog."default",
                data bytea,
                elementindex integer,
                chunkindex integer,
                created timestamp without time zone
            )
            WITH (
                OIDS = FALSE
            )
            TABLESPACE pg_default;'''))

        connection.execute(text(f'''ALTER TABLE IF EXISTS {s_lc}.pickle
                OWNER TO "{username}";'''))

        connection.execute(text(f'''GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE {s_lc}.pickle TO "{n_lc}";'''))
        connection.execute(text(f'''GRANT ALL ON TABLE {s_lc}.pickle TO "{username}";'''))
        
        print(f"Pickle table setup in schema {s_lc} for user {n_lc}")
    
    connection.commit()
