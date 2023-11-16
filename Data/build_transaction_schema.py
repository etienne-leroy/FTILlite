# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from sqlalchemy import Column, Integer, String, Date, Numeric
from sqlalchemy import create_engine, inspect, schema
from sqlalchemy.sql import text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy import exc
from sqlalchemy_utils import database_exists, create_database, drop_database
import os
from os.path import join, dirname
from dotenv import load_dotenv
from datetime import datetime
import click

if __name__ == "__main__":

    # Load node names from .env file
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

    RE_NAME_LEN = max([len(n) for n in node_names])

    # Config
    username = os.environ.get("DB_DEFAULT_USER", "postgres")
    password = os.environ.get("DB_DEFAULT_PW", "postgres")
    server = os.environ.get("DB_HOST", "localhost")
    port = os.environ.get("DB_LOCAL_PORT", "5432")
    address = f"{server}:{port}"
    database = os.environ.get("DB_DEFAULT_DB", "ftillite")

    admin_engine = create_engine(f'postgresql://{username}:{password}@{address}')
    engine = create_engine(f'postgresql://{username}:{password}@{address}/{database}')
   
    # Rename old database
    backup_database = f"{database}_{datetime.now().strftime('%d_%m_%y_%H_%M_%S')}"
    if click.confirm(f'\nNote: Running this script will delete the database "{database}" if it exists.\nDo you want to back this up to "{backup_database}"?', default=True):
        if database_exists(engine.url):
            admin_engine.execute(f"ALTER DATABASE {database} RENAME TO {backup_database}")
    else:
        if database_exists(engine.url):
            drop_database(engine.url)
    # exit(0)
    create_database(engine.url)

    # Declare and create base tables
    Base = declarative_base()

    class Transactions(Base):
        __tablename__ = 'transactions'
        id = Column(Integer, primary_key=True)
        origin_re = Column(String(RE_NAME_LEN))
        origin_id = Column(Integer)
        origin_bsb = Column(Integer)
        dest_re = Column(String(RE_NAME_LEN))
        dest_id = Column(Integer)
        dest_bsb = Column(Integer)
        total_amount = Column(Numeric(13,2))
        transaction_date = Column(Date)
        
    class Accounts(Base):
        __tablename__ = 'accounts'
        re = Column(String(RE_NAME_LEN))
        account_id = Column(Integer, primary_key=True)
        bsb = Column(Integer, primary_key=True)
        account_open_date = Column(Date)
        account_type = Column(String)
        cust_occupation = Column(String)


    inspector = inspect(engine)

    # Create schemas
    for idx, s in enumerate(node_schemas):
        engine.execute(schema.CreateSchema(s))

    Base.metadata.create_all(engine)

    # Create users
    for usr, pw in zip(node_usernames, node_passwords):
        cur = engine.execute(f"SELECT COUNT(*) FROM pg_catalog.pg_roles WHERE rolname = '{usr}'")
        n, = cur.fetchone()
        user_exists = True
        if n == 0:
            # Create User
            engine.execute(f"CREATE USER {usr} WITH PASSWORD '{pw}';")
        else:
            # Update User
            engine.execute(f"ALTER USER {usr} WITH PASSWORD '{pw}';")

    # Load transcations

    # Add indexes
    try:
        engine.execute('''CREATE INDEX IF NOT EXISTS idx_transaction
        ON public.transactions USING btree
        (origin_id ASC NULLS LAST, origin_bsb ASC NULLS LAST, dest_id ASC NULLS LAST, dest_bsb ASC NULLS LAST)
        TABLESPACE pg_default;''')
        print("Transaction Indexes Created")

        engine.execute('''CREATE INDEX idx_account
        ON public.accounts USING btree
        (account_id ASC NULLS LAST, bsb ASC NULLS LAST);''')
        print("Account Indexes Created")
    except exc.ProgrammingError as ex:
        pass

    # Create Views
    for idx, p in enumerate(node_names):
        
        p_lc = p.lower()

        print(f"{p} - Account View and Index Creation Started")

        # Account views
        definition = text(f"""SELECT * FROM public.accounts WHERE "re"='{p}'""")
        engine.execute(f'CREATE MATERIALIZED VIEW {p_lc}.accounts AS {definition}')
        print(f"{p} - Account View Created")
        engine.execute(f'''CREATE INDEX {p_lc}_idx_account
        ON {p_lc}.accounts USING btree
        (account_id ASC NULLS LAST, bsb ASC NULLS LAST);''')
        print(f"{p} - Account Index Created")

        # Transaction Views
        definition = text(f"""SELECT * FROM public.transactions WHERE "origin_re"='{p}' or "dest_re"='{p}'""")
        engine.execute(f'CREATE MATERIALIZED VIEW {p_lc}.transactions AS {definition}')
        print(f"{p} - Transaction View Created")
        engine.execute(f'''CREATE INDEX IF NOT EXISTS {p_lc}_idx_transaction
        ON {p_lc}.transactions USING btree
        (origin_id ASC NULLS LAST, origin_bsb ASC NULLS LAST, dest_id ASC NULLS LAST, dest_bsb ASC NULLS LAST)
        TABLESPACE pg_default;''')
        print(f"{p} - Transaction Index Created")
        
        # Transaction Edges Views
        definition = text(f"""SELECT DISTINCT origin_bsb, origin_id, dest_bsb, dest_id FROM public.transactions WHERE "origin_re"='{p}' or "dest_re"='{p}'""")
        engine.execute(f'CREATE MATERIALIZED VIEW {p_lc}.edges AS {definition}')
        print(f"{p} - Edges View Created")
        engine.execute(f'''CREATE INDEX {p_lc}_idx_edges
        ON {p_lc}.edges USING btree
        (origin_bsb ASC NULLS LAST, origin_id ASC NULLS LAST, dest_bsb ASC NULLS LAST, dest_id ASC NULLS LAST)
        ;''')
        print(f"{p} - Edges Index Created")
        
        # Grant acess to user for all views 
        engine.execute(f'GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA {p_lc} TO {p_lc}_user;')
        
        # Grant acess to user for schema
        engine.execute(f'GRANT USAGE ON SCHEMA {p_lc} TO {p_lc}_user;')

        print(f"{p} - Account View and Index Creation Finished")

    print(f"Schema Creation Finished")