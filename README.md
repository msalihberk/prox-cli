<div align="center">

# prox-cli

### Lightweight, modular, and extensible CLI toolkit for developers and security researchers

A modern Go-based command-line suite for everyday tasks such as encoding, password generation, and network-related checks.

<br>

<p>

[Overview](#-overview) •
[Features](#-features) •
[Installation](#-installation) •
[Usage](#-usage) •
[Commands](#-available-commands)

</p>

<br>

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=for-the-badge&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue?style=for-the-badge)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Active%20Development-orange?style=for-the-badge)]()

<br><br>

**⚡ Fast** • **🧩 Modular** • **🔐 Utility-focused** • **🛠 Developer-friendly**

</div>

---

> [!NOTE]
>
> prox-cli is designed as a lightweight and extensible toolkit. Its current focus is on practical CLI utilities that are easy to extend as the project grows.

# 🚀 Overview

prox-cli is a modular command-line application written in Go. It combines a small set of useful utilities into a single, easy-to-use tool for developers, pentesters, and researchers who want quick access to common operations directly from the terminal.

The project is built around a simple command registry, making it straightforward to add new functionality without rewriting the CLI structure.

# ✨ Features

- 🔐 Base64 encoding and decoding with optional file input/output support
- 🧪 Secure random key generation with customizable character sets
- 🌐 Public IP lookup helper
- 🧩 Extensible command architecture for future modules
- 🛠 Clean and minimal Go-based implementation

# 📁 Project Structure

```text
prox-cli/
├── commands/     # Command implementations
├── main.go       # CLI entry point
├── go.mod        # Go module definition
└── README.md     # Project documentation
```

# 📦 Installation

prox-cli requires Go 1.26 or newer.

## 1. Clone the repository

```bash
git clone https://github.com/msalihberk/prox-cli.git
cd prox-cli
```

## 2. Build the binary

```bash
go build -o prox .
```

## 3. Run the CLI

```bash
./prox help
```

You can also run it directly without building:

```bash
go run . help
```

# 💻 Usage

Here are a few common examples:

```bash
# Base64 encode
 go run . b64 encode "hello"

# Base64 decode
 go run . b64 decode SGVsbG8=

# Generate a secure random key
 go run . keygen 16

# Show the current public IP helper
 go run . myip

# Show version information
 go run . version
```

# 🧰 Available Commands

| Command | Description |
| :--- | :--- |
| `help` | Display the available commands |
| `b64` | Encode or decode strings to and from Base64 |
| `keygen` | Generate random secure keys |
| `myip` | Display public IP information |
| `version` | Show the current version of the CLI |

# 📝 License

This project is licensed under the Apache License 2.0.

See [LICENSE](LICENSE) for more details.
