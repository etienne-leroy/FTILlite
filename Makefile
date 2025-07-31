include .env
export

test_go_non_gpu:
	( cd Peer ; go test ./... -v )

test_go_gpu:
	( cd Peer ; go test ./... -v --gpu )

test_go_coverage:
	( cd Peer; go test -coverprofile=coverage.out -coverpkg ./... ./...; go tool cover -html=coverage.out )

test_py_non_gpu:
	( export GPU_ENABLED=false; docker-compose --profile all_nogpu up -d )
	( sleep 10 )
	( cd Coordinator/ftillitelibrary/tests/ ; pytest . )
	( docker-compose --profile all_nogpu down )

generate_binaries:
	( export CUDA_ARCH=sm_75 ; export CUDA_VERSION=12.2.2 ; cd Peer/lib ; ./build-libftcrypto.sh ) 
	docker-compose --profile all build
	docker-compose --profile all up -d
	docker cp $$(docker ps -aqf "name=peer0"):/app/ftillite-peer .
	docker cp $$(docker ps -aqf "name=peer0"):/lib/libftcrypto.so .

test_py_gpu:
	( cd Coordinator/ftillitelibrary/tests/ ; pytest . )


build_all:
	docker-compose --profile all build

start_all:
	docker-compose --profile all up -d

stop_all:
	docker-compose --profile all down

build_libcrypto:
	( cd Peer/lib ; ./build-libftcrypto.sh )

build_schema:
	( source .venv/bin/activate ; cd Data ; python build_transaction_schema.py )
	( source .venv/bin/activate ; cd Data ; python build_pickle_schema.py )

build_schema_eu:
	( source .venv/bin/activate ; cd Data ; python build_transaction_schema_eu.py )
	( source .venv/bin/activate ; cd Data ; python build_pickle_schema_eu.py )

generate_data:
	@echo "Building schema..."
	$(MAKE) build_schema
	@echo "Generating data files..."
	( source .venv/bin/activate ; cd Data ; python generate_data.py params.ini )
	@echo "Loading data into database..."
	( source .venv/bin/activate ; cd Data ; python load_data.py )

generate_data_eu:
	@echo "Building EU schema..."
	$(MAKE) build_schema_eu
	@echo "Generating EU data files..."
	( source .venv/bin/activate ; cd Data ; python generate_data_eu.py params_eu.ini )
	@echo "Loading EU data into database..."
	( source .venv/bin/activate ; cd Data ; python load_data_eu.py )

generate_data_mini:
	@echo "Generating mini data files..."
	( cd Data ; python generate_data.py params_mini.ini )
	@echo "Building database schema if necessary (will prompt if db exists)..."
	@echo "Loading mini data into database..."
	( cd Data ; python load_data.py )

generate_data_large:
	( cd Data ; python generate_data_large.py params_large.ini )
	( cd Data ; python load_data.py )

generate_all:
	@echo "--- Starting Full Project Regeneration ---"
	
	@echo "[1/8] Stopping all services, removing volumes, and service images..."
	docker-compose --profile all down --rmi all -v || true
	
	@echo "[2/8] Removing local postgres-data directory..."
	sudo rm -rf postgres-data
	
	@echo "[3/8] Pruning Docker builder cache..."
	docker builder prune -a -f
	
	@echo "[4/8] Building libftcrypto..."
	$(MAKE) build_libcrypto
	
	@echo "[5/8] Building all Docker images without cache..."
	docker-compose --profile all build
	
	@echo "[6/8] Starting all services..."
	$(MAKE) start_all
	
	@echo "[7/8] Building database schema..."
	@echo "Generating and loading data..."
	$(MAKE) generate_data_eu
	
	@echo "--- Full Project Regeneration Complete [8/8] ---"
	@echo "All services are up and running with fresh data."



fintracer:
	@( cd Coordinator ; source .venv/bin/activate ; cd PythonScripts ; cd v2.0 ; python fintracer-unified.py )

linear3:
	@( cd Coordinator ; source .venv/bin/activate ; cd PythonScripts ; cd v2.0 ; python fintracer-linear3.py )

nonlinear4:
	@( cd Coordinator ; source .venv/bin/activate ; cd PythonScripts ; cd v2.0 ; python fintracer-nonlinear4.py )

tree5:
	@( cd Coordinator ; source .venv/bin/activate ; cd PythonScripts ; cd v2.0 ; python fintracer-tree5.py ) 

accum:
	@( cd Coordinator ; source .venv/bin/activate ; cd PythonScripts ; cd v2.0 ; python fintracer-accum.py )

investigate:
	@( cd Coordinator ; source .venv/bin/activate ; cd PythonScripts ; python run_investigation.py )


restart_services:
	@echo "Restarting all services with clean state..."
	docker-compose --profile all down
	docker-compose --profile all up -d
	@echo "Waiting for services to initialize..."
	sleep 15
	@echo "Services restarted"


