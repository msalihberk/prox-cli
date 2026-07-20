<div align="center">
<p align="center">
  <img src="https://readme-typing-svg.herokuapp.com?font=JetBrains+Mono&weight=500&size=26&duration=3500&pause=1000&color=00D8FF&center=true&vCenter=true&width=850&lines=prox+ai+cmd+%27list+files+in+the+current+directory%27;prox+lorem+--words+100;prox+ai+explain+%27ERROR%27;prox+myip"/>
</p>

# prox-cli

### Lightweight, modular, and extensible CLI toolkit for developers and security researchers

A modern Go-based command-line suite for everyday tasks such as encoding, password generation, hashing, port scanning, sample text generation, and AI-assisted terminal help.

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

prox-cli is a modular command-line application written in Go. It combines a set of useful utilities into a single, easy-to-use tool for developers, pentesters, and researchers who want quick access to common operations directly from the terminal.

The project is built around a simple command registry, making it straightforward to add new functionality without rewriting the CLI structure. This modular design is one of the core strengths of the project, and new commands can be introduced by adding a small command file under the commands directory.

# ✨ Features

- 🔐 Base64 encoding and decoding with optional file input/output support
- 🧪 Secure random key generation with customizable character sets
- 🧮 Hash generation for MD5, SHA1, SHA256, and SHA512
- 🌐 Port scanning for a target host and a custom port range
- 📝 Lorem Ipsum text generation for testing and placeholders
- 🤖 AI-assisted command generation, command discovery, and log explanation
- 🧩 Extensible command architecture for future modules
- 🛠 Clean and minimal Go-based implementation

# 📁 Project Structure

```text
prox-cli/
├── commands/     # Command implementations
├── core/         # Shared CLI parser and helpers
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

## 3. Setup Prox

```bash
./prox setup
```

## 4. Run the CLI

```bash
prox help
```

You can also run it directly without building:

```bash
go run . help
```

# 💻 Usage

> [!NOTE]
> If you haven't entered a valid API key in the setup command before, use the `prox setup setup-env` command to add your API key as an environment variable for using ai features.

Here are a few common examples using the current command set:

```bash
# Base64 encode
prox b64 encode "hello"

# Base64 decode
prox b64 decode SGVsbG8=

# Generate a secure random key
prox keygen 16

# Generate a hash from a string
prox hash -s test

# Generate a hash from a file
prox hash -f sample.txt -b

# Scan open ports on a target
prox portscan example.com -p 80,443 -w 50 -t 1000

# Generate lorem ipsum text
prox lorem -w 12 -p 2

# Use AI helpers (requires PROX_API_KEY)
prox ai cmd "list files in the current directory"

```

# 🧰 Available Commands

| Command | Description |
| :--- | :--- |
| `help` | Display the available commands |
| `b64` | Encode or decode strings to and from Base64 |
| `keygen` | Generate random secure keys |
| `hash` | Compute the hash of a given input string or file securely |
| `myip` | Display public IP information |
| `portscan` | Scan a target host for open ports concurrently |
| `lorem` | Generate dummy Lorem Ipsum text for testing and placeholders |
| `ai` | Generate terminal commands, discover relevant modules, or explain logs and payloads |
| `setup` | Make it easier to access Prox, save your API key, and enable autocomplete |
| `version` | Show the current version of the CLI |

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

import (
	"prox-cli/core"
)

type NameCommand struct{}

func (c NameCommand) Execute(args []string) error {
	return nil
}

func (c NameCommand) Description() string {
	return "Description"
}
func (c NameCommand) Help() string {
	help := "Usage: prox Name"
	return help
}
func (c NameCommand) SubCommands() []string {
	return []string{"--argument", "help"} // This is for AI agents, list only those with meaningful names (Example: --argument instead of -a)
}
func init() {
	Register("name", NameCommand{})
}
```

# 📝 License

This project is licensed under the Apache License 2.0.

See [LICENSE](LICENSE) for more details.
