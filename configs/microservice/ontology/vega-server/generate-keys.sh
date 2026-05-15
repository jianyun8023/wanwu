#!/bin/bash

# RSA Key Generation Script for Vega Services
# Usage: ./generate-keys.sh [output_dir]
# Default output directory: current directory

set -e

OUTPUT_DIR="${1:-.}"

echo "=== RSA Key Generation Script ==="
echo "Output directory: $OUTPUT_DIR"
echo ""

# Create directories if not exist
mkdir -p "$OUTPUT_DIR/data-connection"
mkdir -p "$OUTPUT_DIR/vega-gateway-pro"

# Generate RSA private key
echo "Generating RSA private key..."
openssl genrsa -out "$OUTPUT_DIR/data-connection/private_key.pem" 2048 2>/dev/null
echo "  Created: $OUTPUT_DIR/data-connection/private_key.pem"

# Generate RSA public key
echo "Generating RSA public key..."
openssl rsa -in "$OUTPUT_DIR/data-connection/private_key.pem" -pubout -out "$OUTPUT_DIR/data-connection/public_key.pem" 2>/dev/null
echo "  Created: $OUTPUT_DIR/data-connection/public_key.pem"

# Copy private key to vega-gateway-pro
echo "Copying private key to vega-gateway-pro..."
cp "$OUTPUT_DIR/data-connection/private_key.pem" "$OUTPUT_DIR/vega-gateway-pro/private_key.pem"
echo "  Created: $OUTPUT_DIR/vega-gateway-pro/private_key.pem"

# Set permissions
echo "Setting file permissions ..."
chmod 644 "$OUTPUT_DIR/data-connection/private_key.pem"
chmod 644 "$OUTPUT_DIR/data-connection/public_key.pem"
chmod 644 "$OUTPUT_DIR/vega-gateway-pro/private_key.pem"

echo ""
echo "=== RSA Key Generation Complete ==="
echo ""
echo "Generated files:"
echo "  - data-connection/private_key.pem    (解密数据源密码)"
echo "  - data-connection/public_key.pem     (加密数据源密码)"
echo "  - vega-gateway-pro/private_key.pem   (解密数据源密码)"
echo ""
echo "IMPORTANT: Do NOT commit these files to version control!"
