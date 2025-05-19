#!/usr/bin/env bash

# We target the Turing NVIDIA architecture by default as this is what we're using in AWS,
# see https://arnon.dk/matching-sm-architectures-arch-and-gencode-for-various-nvidia-cards/
# for list of support values.
if [[ -z "${CUDA_ARCH}" ]]; then
  CUDA_ARCH=${1-sm_61}
else
  CUDA_ARCH=${1-$CUDA_ARCH}
fi

if [[ -z "${CUDA_VERSION}" ]]; then
  CUDA_VERSION=12.2
else
  CUDA_VERSION="${CUDA_VERSION}"


set -e

rm -f libftcrypto.so
rm -f ftcrypto.h

cd libftcrypto

mkdir -p ./out
#read -p "Press any key to resume ..."

# You can switch between using docker or a local install of nvcc by un/commenting the following lines.
# The target environment must be running the same (or newer) version of the CUDA library as we use to compile here.

#docker run --rm -v"$PWD:/src" "nvidia/cuda:${CUDA_VERSION}-devel-ubuntu22.04" nvcc --compiler-options '-fPIC' --shared -arch "${CUDA_ARCH}" -o /src/out/libftcrypto.so /src/src/ftcrypto.cu

# requires: nvidia-cuda-toolkit, build-essential
nvcc --compiler-options '-fPIC' --shared -arch "${CUDA_ARCH}" -o ./out/libftcrypto.so ./src/ftcrypto.cu

cd ..
mv ./libftcrypto/out/libftcrypto.so ./
cp ./libftcrypto/src/ftcrypto.h ./