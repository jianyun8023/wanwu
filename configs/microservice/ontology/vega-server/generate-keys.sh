#!/bin/bash

# RSA Key Generation Script for Vega Services
# Usage: ./generate-keys.sh [output_dir]
# Default output directory: current directory
# Skips key generation if target files are non-empty (delete or clear to regenerate).
# Forward-compatible: copies from data-connection/ if old pem files exist there.
# Also derives state.json from the public key.

set -e

OUTPUT_DIR="${1:-.}"

echo "=== RSA Key Generation Script ==="
echo "Output directory: $OUTPUT_DIR"
echo ""

PRIV="$OUTPUT_DIR/private_key.pem"
PUB="$OUTPUT_DIR/public_key.pem"
STATE_JSON="$OUTPUT_DIR/state.json"

# Forward-compat: old paths under data-connection/
OLD_PRIV="$OUTPUT_DIR/data-connection/private_key.pem"
OLD_PUB="$OUTPUT_DIR/data-connection/public_key.pem"

# Helper: check if a file exists and is non-empty
is_non_empty() { [ -s "$1" ]; }

# Helper: refuse to regenerate if file is non-empty
refuse_if_non_empty() {
  if is_non_empty "$1"; then
    echo "Existing non-empty file detected, refusing to overwrite: $1"
    echo "Delete or clear this file first if you want to regenerate."
    return 0  # true = refused
  fi
  return 1  # false = ok to proceed
}

# --- Private key ---
if refuse_if_non_empty "$PRIV"; then
  echo "  Using existing: $PRIV"
elif is_non_empty "$OLD_PRIV"; then
  echo "Found existing key in old path, copying: $OLD_PRIV -> $PRIV"
  cp "$OLD_PRIV" "$PRIV"
else
  echo "Generating RSA private key..."
  openssl genrsa -out "$PRIV" 2048 2>/dev/null
  echo "  Created: $PRIV"
fi

# --- Public key ---
if refuse_if_non_empty "$PUB"; then
  echo "  Using existing: $PUB"
elif is_non_empty "$OLD_PUB"; then
  echo "Found existing key in old path, copying: $OLD_PUB -> $PUB"
  cp "$OLD_PUB" "$PUB"
elif is_non_empty "$PRIV"; then
  echo "Deriving RSA public key from private key..."
  openssl rsa -in "$PRIV" -pubout -out "$PUB" 2>/dev/null
  echo "  Created: $PUB"
else
  echo "ERROR: No private key available to derive public key, and no existing public key found." >&2
  exit 1
fi

# --- Permissions ---
if [ -f "$PRIV" ]; then chmod 644 "$PRIV"; fi
if [ -f "$PUB" ]; then chmod 644 "$PUB"; fi

echo ""
echo "=== RSA Key Generation Complete ==="

# --- state.json ---
if refuse_if_non_empty "$STATE_JSON"; then
  echo "  Using existing: $STATE_JSON"
elif is_non_empty "$PUB"; then
  echo "Generating state.json from public key..."
  PEM_ESCAPED=$(awk 'BEGIN { ORS="" } { sub(/\r$/, ""); if (NR>1) printf "\\n"; printf "%s", $0 }' "$PUB")
  cat > "$STATE_JSON" << EOF
{
  "publicKey": "${PEM_ESCAPED}"
}
EOF
  echo "  Created: $STATE_JSON"
else
  echo "Public key not found or empty ($PUB), skipping state.json generation."
fi

echo ""
echo "IMPORTANT: Do NOT commit generated key/state files to version control!"
