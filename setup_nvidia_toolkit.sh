#!/bin/bash
set -e # Exit immediately if a command exits with a non-zero status.

echo "Starting NVIDIA Container Toolkit setup..."

echo "Installing prerequisites (curl, gnupg, dnf-utils)..."
sudo dnf install -y curl gnupg dnf-utils --allowerasing

echo "Cleaning up any old NVIDIA Docker repository configurations..."
sudo rm -f /etc/yum.repos.d/nvidia-docker.repo
sudo rm -f /etc/yum.repos.d/nvidia-container-toolkit.repo # Ensure clean slate

echo "Setting up NVIDIA Container Toolkit repository..."
# Define distribution - for Amazon Linux 2023, we'll use the generic approach
# which is generally more robust.
DISTRIBUTION_ID=$(. /etc/os-release; echo $ID)
DISTRIBUTION_VERSION=$(. /etc/os-release; echo $VERSION_ID)

# For Amazon Linux 2023, which is Fedora-based, the RHEL/CentOS paths are often compatible.
# NVIDIA's generic repo setup is preferred.
NVIDIA_REPO_URL="https://nvidia.github.io/libnvidia-container/stable/rpm/nvidia-container-toolkit.repo"

echo "Adding NVIDIA Container Toolkit repository from ${NVIDIA_REPO_URL}..."
curl -s -L "${NVIDIA_REPO_URL}" | sudo tee /etc/yum.repos.d/nvidia-container-toolkit.repo
if [ $? -ne 0 ]; then
    echo "Error: Failed to add NVIDIA Container Toolkit repository."
    exit 1
fi
echo "NVIDIA Container Toolkit repository added."

echo "Importing NVIDIA GPG key..."
# The GPG key URL is often referenced within the .repo file or NVIDIA's documentation.
# Let's use the one commonly cited.
NVIDIA_GPG_KEY_URL="https://nvidia.github.io/libnvidia-container/gpgkey"
NVIDIA_GPG_KEY_PATH="/tmp/nvidia_gpgkey"

curl -fsSL "${NVIDIA_GPG_KEY_URL}" -o "${NVIDIA_GPG_KEY_PATH}"
if [ ! -s "${NVIDIA_GPG_KEY_PATH}" ]; then
    echo "Error: Failed to download GPG key from ${NVIDIA_GPG_KEY_URL}."
    # You can try with verbose curl to debug:
    # curl -vL "${NVIDIA_GPG_KEY_URL}"
    exit 1
fi

sudo rpm --import "${NVIDIA_GPG_KEY_PATH}"
if [ $? -ne 0 ]; then
    echo "Error: Failed to import GPG key."
    rm -f "${NVIDIA_GPG_KEY_PATH}"
    exit 1
fi
rm -f "${NVIDIA_GPG_KEY_PATH}"
echo "NVIDIA GPG Key imported successfully."

echo "Cleaning dnf cache..."
sudo dnf clean expire-cache
sudo dnf makecache # Regenerate cache with the new repo

echo "Installing NVIDIA Container Toolkit..."
# It's good practice to update package lists before installing
sudo dnf check-update || true # Allow to proceed even if there are updates, just to refresh
sudo dnf install -y nvidia-container-toolkit
if [ $? -ne 0 ]; then
    echo "Error: Failed to install nvidia-container-toolkit."
    exit 1
fi
echo "NVIDIA Container Toolkit installed successfully."

echo "Configuring Docker to use the NVIDIA runtime..."
sudo nvidia-ctk runtime configure --runtime=docker
if [ $? -ne 0 ]; then
    echo "Error: nvidia-ctk runtime configure command failed."
    exit 1
fi
echo "Docker runtime configured by nvidia-ctk."

echo "Verifying Docker daemon configuration (/etc/docker/daemon.json)..."
if [ -f /etc/docker/daemon.json ]; then
    echo "Contents of /etc/docker/daemon.json:"
    sudo cat /etc/docker/daemon.json
    if sudo grep -q '"nvidia"' /etc/docker/daemon.json; then
        echo "NVIDIA runtime appears to be configured in daemon.json."
    else
        echo "Warning: NVIDIA runtime not found in daemon.json. Configuration might have failed or uses a different method."
    fi
else
    echo "Warning: /etc/docker/daemon.json not found. nvidia-ctk might have failed to create/modify it."
fi

echo "Restarting Docker service..."
sudo systemctl restart docker
if [ $? -ne 0 ]; then
    echo "Error: Failed to restart Docker service."
    exit 1
fi
echo "Docker service restarted."

echo "Waiting a few seconds for Docker to stabilize..."
sleep 5

echo "Performing final verification: Running nvidia-smi in a CUDA container..."
# Use the same CUDA version as your project for consistency in testing
TEST_CUDA_IMAGE="nvidia/cuda:12.2.2-base-ubuntu22.04"
if sudo docker run --rm --gpus all "${TEST_CUDA_IMAGE}" nvidia-smi; then
    echo "SUCCESS: nvidia-smi ran successfully inside a Docker container."
    echo "NVIDIA Container Toolkit setup appears to be correct."
else
    echo "ERROR: Failed to run nvidia-smi inside a Docker container."
    echo "There might still be an issue with the NVIDIA Docker integration."
    echo "Check Docker logs ('sudo journalctl -u docker.service') and dmesg for errors."
    exit 1
fi

echo "NVIDIA Container Toolkit setup and verification complete."
echo "A system reboot is recommended to ensure all changes are effective."
echo "After rebooting, try 'make test_py_gpu' again."
