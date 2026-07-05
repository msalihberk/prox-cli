<div align="center">

# prox-cli

### Lightweight, modular, and extensible CLI toolkit for developers and security researchers

A modern Go-based command-line suite for everyday tasks such as encoding, password generation, and network-related checks.

> This project is currently in beta and is being actively shaped around a modular architecture that makes it easy to extend.

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
[![Status](https://img.shields.io/badge/Status-Beta-yellow?style=for-the-badge)]()

<br><br>

**⚡ Fast** • **🧩 Modular** • **🔐 Utility-focused** • **🛠 Developer-friendly**

</div>

---

> [!NOTE]
>
> prox-cli is designed as a lightweight and extensible toolkit. Its current focus is on practical CLI utilities that are easy to extend as the project grows.

# 🚀 Overview

prox-cli is a modular command-line application written in Go. It combines a small set of useful utilities into a single, easy-to-use tool for developers, pentesters, and researchers who want quick access to common operations directly from the terminal.

The project is built around a simple command registry, making it straightforward to add new functionality without rewriting the CLI structure. This modular design is one of the core strengths of the project, and new commands can be introduced by adding a small command file under the commands directory.

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
├── LICENSE       # Apache 2.0 License
├── NOTICE        # Third-party software copyrights and legal notices
└── README.md     # Project documentation
```

# 📦 Installation

prox-cli requires Go 1.26 or newer.

You can also download prebuilt binaries from the Releases section of the repository when available.

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

On Windows, the compiled executable can also be run as:

```bash
./prox.exe help
```

You can also run it directly without building:

```bash
go run . help
```

# 💻 Usage

Here are a few common examples:

```bash
# Generate hash
 ./prox b64 hash --sha256 -s test

# Base64 encode
 ./prox b64 encode "hello"

# Portscan
 ./prox b64 portscan SAMPLE_PORT -p 1-500

# Generate a secure random key
 ./prox keygen 16

# Show the current public IP helper
 ./prox myip

# Show version information
 ./prox version
```

# 🧰 Available Commands

| Command | Description |
| :--- | :--- |
| `help`     | Display the available commands |
| `b64`      | Encode or decode strings to and from Base64 |
| `keygen`   | Generate random secure keys |
| `myip`     | Display public IP information |
| `version`  | Show the current version of the CLI |
| `hash`     | Compute the hash of a given input string or file securely |
| `portscan` | Scan a target host for open ports concurrently |
| `lorem `   | Generate dummy Lorem Ipsum text for testing and placeholders |

# 🧩 Adding Your Own Module

Adding a new module is intentionally simple. Create a new Go file in the commands folder, implement the command interface, and register it in the initializer.

If you enjoy the project while it is still in development, you can help by starring the repository, opening pull requests, or sharing your own modules with the community.

You can also use the built-in argument parser from [commands/getargs.go](commands/getargs.go) to handle positional arguments and flags more cleanly.

**Module Template ⬇️**

```go
/* Copyright 2026 Mustafa Salih Berk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package commands

type NameCommand struct{}

func (v NameCommand) Execute(args []string) error {
	return nil
}

func (v NameCommand) Description() string {
	return "Description"
}

func init() {
	register("name", NameCommand{})
}
```

# 📝 License

This project is licensed under the Apache License 2.0.

See [LICENSE](LICENSE) for more details.
