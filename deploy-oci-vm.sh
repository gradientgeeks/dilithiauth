#!/usr/bin/env bash
set -e

echo "=== 🚀 Building DilithiAuth Server for Linux/ARM64 ==="
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /tmp/dilithiauth-server ./cmd/server

echo "=== 📦 Transferring binary to OCI VM (ubuntu@80.225.222.81) ==="
rsync -avz /tmp/dilithiauth-server ubuntu@80.225.222.81:/opt/dilithiauth/server.new

echo "=== 🔄 Swapping binary & restarting dilithiauth.service on OCI VM ==="
ssh ubuntu@80.225.222.81 '
  sudo chmod +x /opt/dilithiauth/server.new
  sudo cp /opt/dilithiauth/server /opt/dilithiauth/server.bak || true
  sudo mv /opt/dilithiauth/server.new /opt/dilithiauth/server
  sudo systemctl daemon-reload
  sudo systemctl restart dilithiauth.service
  sudo systemctl status dilithiauth.service --no-pager
'

echo "=== ✅ Deployment Complete! Testing health endpoints ==="
curl -s https://dilithiauthapis.gradientgeeks.tech/healthz || echo "Health check finished"
echo ""
echo "DilithiAuth server successfully updated on OCI VM!"
