#!/bin/bash

# Exit immediately if a command fails
set -e

# Path to the app's folder (resolve relative to script location)
APP_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Move into the folder
cd "$APP_DIR"

echo "========================================="
echo "   GhostGate Production Test Suite       "
echo "========================================="

# Configuration backup & restore logic to prevent overwriting user config
CONFIG_DIR="$HOME/Library/Application Support/GhostGate"
CONFIG_FILE="$CONFIG_DIR/config.json"
BACKUP_FILE=""

if [ -f "$CONFIG_FILE" ]; then
  BACKUP_FILE=$(mktemp)
  cp "$CONFIG_FILE" "$BACKUP_FILE"
  echo "[*] Backed up existing user config to $BACKUP_FILE"
fi

# Define cleanup function to kill background servers and clean up folders
cleanup() {
  echo ""
  echo "[*] Running test suite cleanup..."
  
  # Kill background jobs started by this script (suppress output)
  jobs -p | xargs kill -9 2>/dev/null || true
  
  # Remove temp files/folders
  rm -rf test_staging_source test_staging_dir test_uploads secret.txt ghostgate_bin config.json
  
  # Restore user config if backed up
  if [ -n "$BACKUP_FILE" ] && [ -f "$BACKUP_FILE" ]; then
    mkdir -p "$CONFIG_DIR"
    cp "$BACKUP_FILE" "$CONFIG_FILE"
    rm "$BACKUP_FILE"
    echo "[*] Restored original user config"
  else
    rm -f "$CONFIG_FILE"
    rmdir "$CONFIG_DIR" 2>/dev/null || true
  fi
  echo "[*] Cleanup complete."
}
trap cleanup EXIT

# 1. Build the Binary
echo "[*] Building GhostGate binary..."
go build -o ghostgate_bin .
echo "[+] Binary built successfully."

# 2. Test Help Functionality
echo "[*] Testing help outputs..."
./ghostgate_bin -h > /dev/null
./ghostgate_bin stage -h > /dev/null
./ghostgate_bin upload -h > /dev/null
./ghostgate_bin tunnel -h > /dev/null
./ghostgate_bin audit -h > /dev/null
./ghostgate_bin init -h > /dev/null
echo "[+] Help output tests passed."

# 3. Test Interactive Init Subcommand
echo "[*] Testing interactive config initialization..."
# Remove any existing config so the overwrite prompt does not appear
rm -f "$CONFIG_FILE"
# Answer the prompts: Port: 9091, Staging Dir: test_staging_dir, Upload Dir: test_uploads, Exfil path: /test_exfil, TLS: n
printf "9091\ntest_staging_dir\ntest_uploads\n/test_exfil\nn\n" | ./ghostgate_bin init > /dev/null

if [ ! -f "$CONFIG_FILE" ]; then
  echo "[-] Fail: config.json was not created by init command."
  exit 1
fi

# Assert values in created config
grep -q '"default_port": "9091"' "$CONFIG_FILE"
grep -q '"default_payloads_directory": "test_staging_dir"' "$CONFIG_FILE"
grep -q '"default_uploads_directory": "test_uploads"' "$CONFIG_FILE"
grep -q '"default_url_path": "/test_exfil"' "$CONFIG_FILE"
grep -q '"default_tls_enabled": false' "$CONFIG_FILE"
echo "  -> Config values verified correctly in $CONFIG_FILE"
echo "[+] Interactive config initialization test passed."

# Remove the config file created so fallback/default tests run cleanly
rm -f "$CONFIG_FILE"

# 4. Test Argument Validation (expected to fail)
echo "[*] Testing input validation edge cases..."

set +e # Allow commands to fail for validation checks

# A. Invalid port (out of range)
./ghostgate_bin stage -p 999999 2>/dev/null
if [ $? -eq 0 ]; then
  echo "[-] Fail: Managed to run stage with invalid port 999999"
  exit 1
fi

# B. Invalid port (non-numeric)
./ghostgate_bin stage -p abc 2>/dev/null
if [ $? -eq 0 ]; then
  echo "[-] Fail: Managed to run stage with non-numeric port 'abc'"
  exit 1
fi

# C. Invalid staging directory (only special characters - no alphabetical chars)
./ghostgate_bin stage -p 9091 -d "123456" 2>/dev/null
if [ $? -eq 0 ]; then
  echo "[-] Fail: Managed to run stage with numeric-only directory"
  exit 1
fi

# D. Invalid URL path for upload (not starting with / or containing space)
./ghostgate_bin upload -p 9091 -u "invalid path" 2>/dev/null
if [ $? -eq 0 ]; then
  echo "[-] Fail: Managed to run upload with invalid path spaces"
  exit 1
fi

# E. Invalid tunnel target URL
./ghostgate_bin tunnel -p 9091 -t "ftp://google.com" 2>/dev/null
if [ $? -eq 0 ]; then
  echo "[-] Fail: Managed to run tunnel with unsupported protocol scheme (ftp)"
  exit 1
fi

# F. Invalid audit target URL (could not parse)
./ghostgate_bin audit -t ":" 2>/dev/null
if [ $? -eq 0 ]; then
  echo "[-] Fail: Managed to run audit with unparseable URL"
  exit 1
fi

# G. Audit connection fail (should exit gracefully without panicking)
./ghostgate_bin audit -t "http://127.0.0.1:65530" 2>&1 | grep -q "Connection failed"
if [ $? -ne 0 ]; then
  echo "[-] Fail: Audit failed to output connection failure message cleanly."
  exit 1
fi

set -e # Re-enable exit on command error
echo "[+] Validation checks passed."

# 5. Test Configuration Fallbacks
echo "[*] Testing configuration default fallbacks..."
# Run stage in the background without flags (should start on port 8080 using 'payloads' directory)
mkdir -p payloads
./ghostgate_bin stage > stage_fallback.log 2>&1 &
STAGE_FALLBACK_PID=$!

# Wait for server to bind
sleep 1.5

# Check if port 8080 is listening
if ! nc -z 127.0.0.1 8080; then
  echo "[-] Fail: Fallback stage server is not listening on port 8080"
  cat stage_fallback.log
  exit 1
fi

# Kill fallback staging server
kill -9 $STAGE_FALLBACK_PID 2>/dev/null || true
rm -rf payloads stage_fallback.log
echo "[+] Staging configuration fallback checks passed."

# 6. Test Staging Server Functionality
echo "[*] Testing Staging Server functionality..."
mkdir -p test_staging_source
echo "Hello from GhostGate!" > test_staging_source/hello.txt

# Run stage command
./ghostgate_bin stage -p 9091 -d test_staging_dir -s test_staging_source > stage.log 2>&1 &
STAGE_PID=$!

# Wait for server to copy files and spin up
sleep 1.5

# Verify source files were copied to staging directory
if [ ! -f "test_staging_dir/hello.txt" ]; then
  echo "[-] Fail: Staging server failed to copy files from source directory."
  exit 1
fi

# Retrieve file via HTTP curl
CONTENT=$(curl -s http://127.0.0.1:9091/hello.txt)
if [ "$CONTENT" != "Hello from GhostGate!" ]; then
  echo "[-] Fail: Curl response does not match expected staging content. Got: $CONTENT"
  exit 1
fi

# Kill the staging server (trigger defer cleanup)
kill -15 $STAGE_PID 2>/dev/null || true
sleep 1.5

# Verify staging directory was automatically cleaned up
if [ -d "test_staging_dir" ]; then
  echo "[-] Fail: Staging directory was not cleaned up automatically."
  exit 1
fi
echo "[+] Staging Server functionality test passed."

# 7. Test Upload/Exfil Server Functionality
echo "[*] Testing Upload/Exfil Server functionality..."
echo "My Secret Exfiltration Data" > secret.txt

# Run upload command
./ghostgate_bin upload -p 9092 -u /exfil -d test_uploads > upload.log 2>&1 &
UPLOAD_PID=$!

sleep 1.5

# Send exfiltration payload
UPLOAD_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST --data-binary @secret.txt -H 'X-File-Name: secret.txt' http://127.0.0.1:9092/exfil)
if [ "$UPLOAD_STATUS" != "201" ]; then
  echo "[-] Fail: Upload server returned HTTP status: $UPLOAD_STATUS (expected 201)"
  exit 1
fi

# Assert file uploaded correctly
if [ ! -f "test_uploads/secret.txt" ]; then
  echo "[-] Fail: Uploaded file does not exist on disk."
  exit 1
fi

UPLOADED_CONTENT=$(cat test_uploads/secret.txt)
if [ "$UPLOADED_CONTENT" != "My Secret Exfiltration Data" ]; then
  echo "[-] Fail: Uploaded file content mismatch. Got: $UPLOADED_CONTENT"
  exit 1
fi

kill -9 $UPLOAD_PID 2>/dev/null || true
echo "[+] Upload/Exfil Server functionality test passed."

# 8. Test Pivot Tunnel Server Functionality
echo "[*] Testing Pivot/Tunnel Server functionality..."

# Start target mock backend (use stage server on port 9093)
mkdir -p test_staging_source
echo "Backend Content" > test_staging_source/hello.txt
./ghostgate_bin stage -p 9093 -d test_staging_dir -s test_staging_source > stage_backend.log 2>&1 &
BACKEND_PID=$!

# Start tunnel server forwarding traffic on port 9094 to mock backend
./ghostgate_bin tunnel -p 9094 -t http://127.0.0.1:9093 > tunnel.log 2>&1 &
TUNNEL_PID=$!

sleep 1.5

# Test request forwarding through tunnel
TUNNEL_CONTENT=$(curl -s http://127.0.0.1:9094/hello.txt)
if [ "$TUNNEL_CONTENT" != "Backend Content" ]; then
  echo "[-] Fail: Tunnel server failed to route request correctly. Got: $TUNNEL_CONTENT"
  exit 1
fi
echo "[+] Pivot/Tunnel Server functionality test passed."

# 9. Test Configuration Audit Functionality
echo "[*] Testing Configuration Audit functionality..."
# Audit the mock backend running on port 9093
AUDIT_OUTPUT=$(./ghostgate_bin audit -t http://127.0.0.1:9093)

# Check if target audit section and response headings are present
if echo "$AUDIT_OUTPUT" | grep -q "TARGET HTTP SERVER AUDIT"; then
  echo "[+] Configuration Audit functionality test passed."
else
  echo "[-] Fail: Configuration audit did not output expected analysis report. Output: $AUDIT_OUTPUT"
  exit 1
fi

# Stop target mock backend and tunnel server
kill -9 $BACKEND_PID $TUNNEL_PID 2>/dev/null || true

echo "========================================="
echo "   All tests passed successfully!        "
echo "========================================="
