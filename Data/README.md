Note there are issues with numpy on Windows where an int defaults to 32 bits rather than 64:

https://stackoverflow.com/questions/36278590/numpy-array-dtype-is-coming-as-int32-by-default-in-a-windows-10-64-bit-machine

Recommended to use a linux system or WSL.

# Generate Data

To generate synthetic data, firstly install the required Python packages:
```
pip install -r requirements.txt
```
Then run the below command:
```
python generate_data.py params.ini
```
or in the root directory run the following:
```
make generate_data
```

You can tweak the config ini file; however, it is not required.

# Load Data

To load the data into the postgreSQL database, firstly install the required Python packages (as shown above).

Then spin-up the PostgreSQL server via ```docker-compose up -d``` in the root directory of this repo.

Then run:
```
python data_load_postgres.py
```
# Large Dataset Generation

For datasets that exceed 1 million nodes, it is recommended to use the generate_data_large.py file. 

This uses numpy instead of networkX (which uses dictionaries) to drastically improve performance; however, uses near-scale free graph generation with RMAT rather than guaranteed scale free with Powerlaw.

You will firstly need to clone the repo https://github.com/MattWMitchell/PaRMAT. 

Assuming you have the correct C++ compilers as instructed in this repo, run:
```
cd Release
make
cp PaRMAT <ftillite-repo-location>/DataGenerator/.
```
Which will provide a memory efficient and concurrent RMAT generator executable.

Then to use, simply tweak the params_large.ini or create a separate one for it and run:
```
python generate_data_large.py params_large.ini
```
or in the root directory run the following:
```
make generate_data_large
```

