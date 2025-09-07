# Ollama Installation Guide (China GPU Setup)

## Prerequisites
- NVIDIA GPU with Compute Capability 5.0+
- Latest [NVIDIA drivers](https://www.nvidia.com/en-us/drivers/unix/) installed
- Matching [CUDA Toolkit version](https://developer.nvidia.com/cuda-downloads)

## Installation Methods

### Option 1: Quick Install (VPN Required)
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

### Option 2: Manual Install (China-Friendly)

Download Binaries:

```bash
# For ARM64 (Jetson/NVIDIA ARM)
curl -L https://ollama.com/download/ollama-linux-arm64.tgz -o ollama.tgz
```

```bash
# For x86_64 (Intel/AMD)
curl -L https://ollama.com/download/ollama-linux-amd64.tgz -o ollama.tgz
```

Extract files

```bash
# For .tgz files:
sudo tar -C /usr -xzf ollama.tgz

# For .tar files:
sudo tar -C /usr -xf ollama.tar
```
---

(_Optional Ollama als einen Startup-Service registrieren_)

Creating a new user and group for ollama

```bash
sudo useradd -r -s /bin/false -U -m -d /usr/share/ollama ollama
sudo usermod -a -G ollama $(whoami)
```

Create a new service file `/etc/systemd/system/ollama.service`

```
[Unit]
Description=Ollama Service
After=network-online.target

[Service]
ExecStart=/usr/bin/ollama serve
User=ollama
Group=ollama
Restart=always
RestartSec=3
Environment="PATH=$PATH"

[Install]
WantedBy=multi-user.target
```

Start the service

```bash
sudo systemctl daemon-reload
sudo systemctl enable ollama
```
