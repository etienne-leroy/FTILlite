# =====================================
#
# Copyright (c) 2023, AUSTRAC Australian Government
# All rights reserved.
#
# Licensed under BSD 3 clause license
#  
########################################

from sqlalchemy import Column, Integer, String, Date, Numeric, BigInteger
from sqlalchemy import create_engine, inspect, schema
from sqlalchemy.sql import text
from sqlalchemy.orm import declarative_base  # Updated import
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
            # Use connection instead of engine.execute
            with admin_engine.connect() as connection:
                connection.execute(text(f'ALTER DATABASE "{database}" RENAME TO "{backup_database}"'))
                connection.commit()
    else:
        if database_exists(engine.url):
            drop_database(engine.url)
    
    create_database(engine.url)

    # Declare and create base tables
    Base = declarative_base()

    class Transactions(Base):
        __tablename__ = 'transactions'
        id = Column(Integer, primary_key=True)
        origin_re = Column(String(RE_NAME_LEN))
        origin_id = Column(BigInteger)  # Use BigInteger for 13-digit account numbers
        origin_bank_id = Column(Integer)
        dest_re = Column(String(RE_NAME_LEN))
        dest_id = Column(BigInteger)    # Use BigInteger for 13-digit account numbers
        dest_bank_id = Column(Integer)
        total_amount = Column(Numeric(13,2))
        transaction_date = Column(Date)
        
    class Accounts(Base):
        __tablename__ = 'accounts'
        re = Column(String(RE_NAME_LEN))
        account_id = Column(BigInteger, primary_key=True)  # Use BigInteger for 13-digit account numbers
        bank_id = Column(Integer, primary_key=True)
        account_open_date = Column(Date)
        account_type = Column(String)
        cust_occupation = Column(String)

    inspector = inspect(engine)

    # Create schemas using connection
    with engine.connect() as connection:
        for idx, s in enumerate(node_schemas):
            connection.execute(schema.CreateSchema(s))
        connection.commit()

    Base.metadata.create_all(engine)

    # Create users using connection
    with engine.connect() as connection:
        for usr, pw in zip(node_usernames, node_passwords):
            cur = connection.execute(text(f"SELECT COUNT(*) FROM pg_catalog.pg_roles WHERE rolname = '{usr}'"))
            n, = cur.fetchone()
            if n == 0:
                # Create User
                connection.execute(text(f"CREATE USER \"{usr}\" WITH PASSWORD '{pw}';"))
            else:
                # Update User
                connection.execute(text(f"ALTER USER \"{usr}\" WITH PASSWORD '{pw}';"))
        connection.commit()

    # Add indexes using connection
    try:
        with engine.connect() as connection:
            connection.execute(text('''CREATE INDEX IF NOT EXISTS idx_transaction
            ON public.transactions USING btree
            (origin_id ASC NULLS LAST, origin_bank_id ASC NULLS LAST, dest_id ASC NULLS LAST, dest_bank_id ASC NULLS LAST)
            TABLESPACE pg_default;'''))
            print("Transaction Indexes Created")

            connection.execute(text('''CREATE INDEX IF NOT EXISTS idx_account
            ON public.accounts USING btree
            (account_id ASC NULLS LAST, bank_id ASC NULLS LAST);'''))
            print("Account Indexes Created")
            connection.commit()
    except exc.ProgrammingError as ex:
        print(f"Index creation error: {ex}")
        pass

    # Create Views using connection
    with engine.connect() as connection:
        for idx, schema_name in enumerate(node_schemas):
            node_name = node_names[idx]
            
            print(f"{node_name} - View Creation Started for schema {schema_name}")

            # Create views that filter for the correct peer based on the node name
            # This maps the node name directly to the peer name in the data
            if node_name == 'COORDINATOR':
                # Create empty views for coordinator (regulatory body)
                print(f"Creating empty views for {schema_name} (regulatory coordinator)")
                connection.execute(text(f'CREATE MATERIALIZED VIEW {schema_name}.accounts AS SELECT * FROM public.accounts WHERE false'))
                connection.execute(text(f'CREATE MATERIALIZED VIEW {schema_name}.transactions AS SELECT * FROM public.transactions WHERE false'))
                connection.execute(text(f'CREATE MATERIALIZED VIEW {schema_name}.edges AS SELECT DISTINCT origin_bank_id, origin_id, dest_bank_id, dest_id FROM public.transactions WHERE false'))
            else:
                # For peer nodes, use the node name directly as the peer identifier
                actual_peer = node_name  # This should be PEER_1, PEER_2, etc.
                print(f"Creating data views for {schema_name} filtering for peer {actual_peer}")
                connection.execute(text(f'CREATE MATERIALIZED VIEW {schema_name}.accounts AS SELECT * FROM public.accounts WHERE re = \'{actual_peer}\''))
                connection.execute(text(f'CREATE MATERIALIZED VIEW {schema_name}.transactions AS SELECT * FROM public.transactions WHERE origin_re = \'{actual_peer}\' OR dest_re = \'{actual_peer}\''))
                connection.execute(text(f'CREATE MATERIALIZED VIEW {schema_name}.edges AS SELECT DISTINCT origin_bank_id, origin_id, dest_bank_id, dest_id FROM public.transactions WHERE origin_re = \'{actual_peer}\' OR dest_re = \'{actual_peer}\''))

            # Create indexes for all schemas (even empty ones)
            connection.execute(text(f'''CREATE INDEX {schema_name}_idx_account
            ON {schema_name}.accounts USING btree
            (account_id ASC NULLS LAST, bank_id ASC NULLS LAST);'''))
            print(f"{node_name} - Account Index Created")

            connection.execute(text(f'''CREATE INDEX {schema_name}_idx_transaction
            ON {schema_name}.transactions USING btree
            (origin_id ASC NULLS LAST, origin_bank_id ASC NULLS LAST, dest_id ASC NULLS LAST, dest_bank_id ASC NULLS LAST);'''))
            print(f"{node_name} - Transaction Index Created")
            
            connection.execute(text(f'''CREATE INDEX {schema_name}_idx_edges
            ON {schema_name}.edges USING btree
            (origin_bank_id ASC NULLS LAST, origin_id ASC NULLS LAST, dest_bank_id ASC NULLS LAST, dest_id ASC NULLS LAST);'''))
            print(f"{node_name} - Edges Index Created")
            
            # Grant access to user for all views 
            user_name = node_usernames[idx]
            connection.execute(text(f'GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA {schema_name} TO "{user_name}";'))
            connection.execute(text(f'GRANT USAGE ON SCHEMA {schema_name} TO "{user_name}";'))

            print(f"{node_name} - View Creation Finished")
        
        connection.commit()

    print(f"Schema Creation Finished")