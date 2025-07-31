# ftillite

This repository contains both the front-end (Coordinator - Python) and back-end (Peer - Golang) code to deliver the ftillite body of work.

# Folder Structure

| Folder | Description |
| ------ | ------ |
| Coordinator | Contains the **client** and **Compute Manager** code written in **Python**. The coordinator node will run a single instance of both the Client and Compute Manager. |
| Peer | Contains the **Segment Manager** code written in **golang**. Each peer node will run a Segment Manager instance. | 
| DataGenerator | Contains code to generate synthetic data and load this data into the postgreSQL db. | 
| Documentation | Contains all documentation relevant to a technical user / developer of the repo. | 

# Solution Overview Diagram

Below is a diagram that conveys the core solution design.

![Solution Overview Diagram](/Documentation/Solution-Diagram.png "Solution Overview Diagram")

# CI/CD

\<Removed\>

# AWS Environment

\<Removed\>

## Running the solution locally

The FTILLITE solution can be run using docker-compose. The docker-compose.yml file defines two profiles, `all` and `debug`. The `all` profile runs each service specified in the compose file, and the `debug` profile runs everything except the first Peer segment, to allow for this to be run in debug more within an IDE. Open a console at the root directory of the solution and enter
```
make start_all
```
This will start a similar set of containers that are configured to run in AWS. The Jupyter Notebooks and RabbitMQ admin interfaces can both be accessed in the same way as described for AWS but by substituting localhost as the IP Address. Note there may be a need to change connection strings in the notebook if you're running a hybrid between an IDE and docker containers.

To stop the containers use
```
make stop_all
```
Note that with any updates to the repository you need to run the following for those changes to apply:
```
make build_all
make start_all
```

## Full Project Regeneration

For a complete fresh start with clean Docker environment and new data, use:
```
make generate_all
```
This command performs a comprehensive project refresh including:
- Stopping all services and removing volumes/images
- Cleaning Docker builder cache
- Rebuilding libftcrypto and all Docker images
- Starting services with fresh EU transaction data

## Update database schema

To update the database schema ensure that the postgres container is running. From the Database folder run the `update-postgres-container.sh` script. Note that you may need to change the name of the target container within the script from `ftillite-postgres-1` to `ftillite_postgres_1` depending on your environment.

## Populate the database locally

### Australian Transaction Data

To populate the database with Australian transaction data, perform the following with the FTILLITE system running via the above docker-compose command. Note that you will want to use a linux environment such as WSL due to Numpy limitations on windows systems for the data generation.
```
make generate_data
```

### European Transaction Data

To populate the database with European transaction data:
```
make generate_data_eu
```

Note that you can tweak parameters on the data generation; see DataGenerator/README.MD for details.

Also note that if you have an older version of the database running or have changed any of the db details in the .env file, then you'll need to do the following in order to get the schemas and users setup correctly for each Peer node:

1. Run ```make stop_all```

2. Delete the postgres-data folder

3. Run ```make start_all```

## Financial Transaction Analysis (FinTracer)

The FTILLITE solution includes comprehensive financial transaction analysis capabilities through the FinTracer module, which implements various money laundering typologies and suspicious transaction patterns.

### Available Typologies

The following financial crime typologies are available for analysis:

| Typology | Command | Description |
| -------- | ------- | ----------- |
| Unified Analysis | `make fintracer` | Comprehensive analysis including all typologies with optional result storage and deep analysis |
| Linear (3 accounts) | `make linear3` | Simple layering through 3 sequential accounts |
| Non-linear (4 accounts) | `make nonlinear4` | Complex routing through 4 accounts with multiple paths |
| Tree (5 accounts) | `make tree5` | Hierarchical distribution pattern across 5 accounts |
| Accumulation | `make accum` | Fund aggregation from multiple sources into target accounts |

### Running FinTracer Analysis

To run the complete financial analysis suite:
```
make fintracer
```

This will execute the unified FinTracer script which provides options to:
- Run individual typology analyses
- Store results for further investigation
- Perform comprehensive cross-typology analysis
- Generate detailed reports

Individual typologies can also be run separately using their respective make commands.

### Investigation Tools

For detailed investigation of suspicious patterns identified by FinTracer:
```
make investigate
```

The FinTracer module includes helper function libraries and type definition libraries to maintain clean, modular code structure while providing comprehensive analysis capabilities.

# Developement

## Local environment setup

The following tools are expected to be installed on your machine:

- Go
- Docker
- NVIDIA toolkit 11.4 (optional)

Before the Go code can be compiled you must have built libftcrypto and have the shared object file 
and header in Peer/lib/. Building libftcrypto can be done by:

1. Initialising and updating the Git submodule (Peer/lib/libftcrypto):
    - git submodule init
    - git submodule update
2. If you don't have the NVIDIA toolkit available locally, update the build-libftcrypto.sh script to use
   docker and comment out the use of nvcc.
3. Run build-libftcrypto.sh to build the shared object file and copy the header file to Peer/lib
* Make sure to set the `CUDA_ARCH` and `CUDA_VERSION` environment variables to match your system's specifications.

Note: If you're using Visual Studio Code, you may need to open the Peer/segment/types/ed25519array_gpu.go
      file and click the "regenerate cgo definitions" link at the top of the file to stop VSCode from 
      showing errors in the codebase.


## Updating libftcrypto

There are two changes required when updates are available from libftcrypto, both should be done to keep everything consistent.

### ftillite repository

Updating libftcrypto in the repository is required so that local builds of ftillite Peer nodes are linked against the same libftcrypto library.

To update the repository:

1. Create a new branch/merge request for the update change.
2. From within the submodule path, `Peer/lib/libftcrypto`, run `git checkout master && git pull` to grab the latest from upstream master.
3. Navigate back out of the submodule path and commit the change.
4. Run `Peer/lib/build-libftcrypto.sh` to rebuild libftcrypto and update your local library (i.e., `Peer/lib/libftcrypto.so`).

# Testing

## Local Environment Setup

### Python Tests
To run the python tests, you will need to install all packages, including the ftillite library firstly with:
```
cd Coordinator
pip install -r requirements.txt
pip install -e ftillitelibrary
```

Then to run the tests, run the either of the following in the project root directory:
* `make test_py_non_gpu` - for non-GPU environments
* `make test_py_gpu` - for GPU environments

#### Python Profiling

To profile the code, you can either run the following:
```
python -m cProfile <Python Module name>.py
```

Or you can add code programatically, which allows you to output data to a CSV. 
See `Coordinator/ftillitelibrary/tests/check_listmap.py` for an example of this.

Note that `cumtime` and `ncalls` are the most crucial metrics, as all others are obscured due to the heavy lifting being done on the Go-Side.

### Go Tests

To run the tests, run the either of the following in the project root directory:
* `make test_go_non_gpu` - for non-GPU environments
* `make test_go_gpu` - for GPU environments

**Note** that it is recommended not to use the VsCode GUI as test cases that throw a panic will silently fail.

#### Additional Notes
In order to run test cases for the Ed25519 type, you will either need to run you test cases with the flag `--gpu=true` or in Vscode, add a settings.json to the root directory `.vscode` of this repo with the below contents:

```
{
    "go.testEnvVars": {
        "gpu": "true"
    }
}
```

You may also need to extend the VScode setting `Test Timeout` to `30m` instead of the default `30s`.

Also when adding a python test, you will need to:

* Name the Python module / file `test_<rest of name>.py` and
* functions within this module will also need to be defined as `def test_<test name>()` for pytest to register these new test cases.

## Writing a new Python command backed by Go

The following steps can be used to implement a new FTILLite command. We'll implement the `sha3_256` as an example, as it requires changes to both the Python library and the Go backend. 

### Python Library

The `Coordinator/ftillitelibrary/ftillite` directory contains the `ftillite` Python module, in which is the `client.py`
file containing the class hierarchy described in the FTILLite documentation. 

`SHA3_256` is defined as a top-level function of the `ftillite` module, so the following code can be placed anywhere outside 
any of the class blocks in the `client.py` file. To keep the file organised, these are generally placed at the bottom of the file.

```python
def sha3_256(data):
    return data.context()._exec_command(f'sha3_256 1 {data.handle()}')
```

Explanation of the code:

- The `data` parameter is Bytearray Array whose elements will be hashed. The result will be a Bytearray Array of the same length, but strictly 32 bytes wide, containing the hashes.
- `.context()` returns the FTILContext associated with the FTILLite variable (`data`).
- `_exec_command` is a function defined the FTILContext which sends the command to the message broker.
- `"sha3_256 1 {data.handle()}"` is the command string sent to the Peers (i.e., the Go backend). Before the command string is sent, the following preprocessing steps are applied:
    - `sha3_256` will be prefixed with `command_`
    - `1` will be replaced with a new variable handle. `0` can be used to indicate that no handle should be created, which are typically used for in-place commands.
    - `{data.handle()}` is replaced with the handle of the `data` variable.
    
    The exact format of the command string will vary between commands, however the important thing is that both the Python and Go code agree on what the format is.

### Go

The `segment` package in `Peer/segment` contains the Go implementations of the commands. To define a new command in Go:

1. Define a new command constant in `command_constants.go`, e.g., `CommandSHA3256 = "command_sha3_256"`. The constants value must match the command name as sent by the Python library.
2. Create a new Go function which will contain the logic of the command. The type of the function is `func ([]string) (string, error)`, which has:
    - An array of parameters sent by Python, in the case of `sha3_256`, the first parameter will be the result handle, and the second the `data` variable handle.
    - A string result, or an error result. The format of the string result can be found in the `FTILContext :: _parse_result` Python function in the `client.py` file.
3. Register the function in `command_registration.go`, e.g., `s.RegisterCommand(CommandSHA3256, s.sha3_256)`.
4. Write unit tests.

# Deployment

\<Removed\>