#!/bin/bash
set -e

echo "--- Peer Container Diagnostics ---"

echo "[INFO] Current PATH variable:"
echo "${PATH}"

echo "[INFO] Listing contents of /usr/local/nvidia:"
ls -la /usr/local/nvidia || echo "[WARN] Failed to list /usr/local/nvidia"

echo "[INFO] Listing contents of /usr/local/nvidia/bin:"
ls -la /usr/local/nvidia/bin || echo "[WARN] Failed to list /usr/local/nvidia/bin (directory might not exist or is empty)"

echo "[INFO] Listing contents of /usr/local/nvidia/lib:"
ls -la /usr/local/nvidia/lib || echo "[WARN] Failed to list /usr/local/nvidia/lib (directory might not exist or is empty)"

echo "[INFO] Listing contents of /usr/local/nvidia/lib64:"
ls -la /usr/local/nvidia/lib64 || echo "[WARN] Failed to list /usr/local/nvidia/lib64 (directory might not exist or is empty)"


echo "[INFO] Attempting to find nvidia-smi with 'which':"
which nvidia-smi || echo "[WARN] 'which nvidia-smi' failed or command not found in PATH."

echo "[INFO] Attempting to find nvidia-smi with 'find' in /usr/local/nvidia and /usr/bin:"
find /usr/local/nvidia /usr/bin -name nvidia-smi -ls 2>/dev/null || echo "[WARN] nvidia-smi not found by 'find' in common locations."

echo "[INFO] Attempting to run nvidia-smi using full path /usr/local/nvidia/bin/nvidia-smi:"
if [ -f /usr/local/nvidia/bin/nvidia-smi ]; then
    /usr/local/nvidia/bin/nvidia-smi || echo "[WARN] /usr/local/nvidia/bin/nvidia-smi command failed or GPU not visible."
else
    echo "[WARN] /usr/local/nvidia/bin/nvidia-smi does not exist."
fi

echo "[INFO] LD_LIBRARY_PATH:"
echo "${LD_LIBRARY_PATH}"

echo "[INFO] Checking libftcrypto.so linkage:"
if [ -f /lib/libftcrypto.so ]; then
    ldd /lib/libftcrypto.so || echo "[WARN] ldd on /lib/libftcrypto.so failed."
else
    echo "[WARN] /lib/libftcrypto.so not found."
fi

echo "[INFO] Checking nvcc version (if available in base image):"
if command -v nvcc &> /dev/null
then
    nvcc --version || echo "[WARN] nvcc --version failed."
else
    echo "[INFO] nvcc not found in PATH (expected for -base images)."
fi

echo "--- End Peer Container Diagnostics ---"

echo "[INFO] Starting ftillite-peer application..."
exec /app/ftillite-peer
