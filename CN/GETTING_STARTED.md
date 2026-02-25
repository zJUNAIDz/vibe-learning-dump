# Getting Started — Environment Setup

---

## What You Need

This curriculum assumes you have a **Linux machine** (or VM) with root/sudo access.

### Minimum Setup
- Linux (Ubuntu 20.04+ recommended, Debian/Fedora also fine)
- Root or sudo access
- Internet connection
- Terminal emulator

### Recommended Tools (install all)

```bash
# Core networking tools
sudo apt update
sudo apt install -y \
  iproute2 \
  iputils-ping \
  traceroute \
  mtr \
  tcpdump \
  net-tools \
  dnsutils \
  curl \
  wget \
  netcat-openbsd \
  nmap \
  iperf3 \
  wireshark-cli \
  bridge-utils \
  ethtool \
  conntrack

# For network simulation/virtualization modules
sudo apt install -y \
  docker.io \
  socat \
  hping3
```

### Optional but Useful
- **Wireshark** (GUI) — for the packet analysis module
- **VS Code** — for reading these notes alongside terminal
- A **second machine or VM** — for testing connectivity between hosts

---

## Verify Your Setup

Run these commands to verify everything is working:

```bash
# Check ip tool
ip addr show

# Check ss (socket statistics)
ss -tuln

# Check tcpdump
sudo tcpdump --version

# Check DNS tools
dig google.com

# Check network namespaces support
sudo ip netns add test_ns && sudo ip netns delete test_ns && echo "Namespaces work"

# Check iperf3
iperf3 --version

# Check traceroute
traceroute --version
```

If any of these fail, install the missing package before proceeding.

---

## A Note on Permissions

Many networking commands require **root access** (`sudo`). This is because:
- Reading raw packets (`tcpdump`) requires access to the network interface
- Creating network namespaces requires kernel privileges
- Modifying routing tables requires administrative access

Throughout this curriculum, commands that need `sudo` will always show it explicitly.

---

## Terminal Setup Tips

```bash
# Add useful aliases to ~/.bashrc or ~/.zshrc
alias ll='ls -la'
alias ss='ss -tuln'
alias tcpd='sudo tcpdump -i any -nn'
alias routes='ip route show'
alias ifs='ip -br addr show'
```

You're ready. Start with [00-orientation/01-what-is-a-network.md](00-orientation/01-what-is-a-network.md).
