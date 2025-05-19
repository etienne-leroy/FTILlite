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

test_py_gpu:
	docker-compose --profile all up -d
	( sleep 10 )
	( cd Coordinator/ftillitelibrary/tests/ ; pytest . )
	( docker-compose --profile all down )

build_all:
	docker-compose --profile all build

start_all:
	docker-compose --profile all up -d

stop_all:
	docker-compose --profile all down

build_libcrypto:
	( cd Peer/lib ; ./build-libftcrypto.sh )

build_schema:
	( cd Data ; python build_transaction_schema.py )
	( cd Data ; python build_pickle_schema.py )

generate_data:
	( cd Data ; python generate_data_lux.py params_lux.ini )
	( cd Data ; python load_data.py )

generate_data_large:
	( cd Data ; python generate_data_large.py params_large.ini )
	( cd Data ; python load_data.py )

generate_binaries:
	( export CUDA_ARCH=sm_70 ; export CUDA_VERSION=11.6.2 ; cd Peer/lib ; ./build-libftcrypto.sh ) 
	docker-compose --profile all build
	docker-compose --profile all up -d
	docker cp $$(docker ps -aqf "name=peer0"):/app/ftillite-peer .
	docker cp $$(docker ps -aqf "name=peer0"):/lib/libftcrypto.so .