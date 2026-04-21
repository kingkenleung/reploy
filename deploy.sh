#!/bin/bash
set -e

VM_USER="root"
VM_IP="34.96.183.161"  # update if your VM IP changes

echo "==> Building Linux binary on Mac..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o reploy-linux ./cmd/server

echo "==> Pushing code to GitHub..."
git add -A
git commit -m "deploy: $(date '+%Y-%m-%d %H:%M')" || echo "Nothing new to commit"
git push origin master

echo "==> Copying binary to VM..."
scp reploy-linux "$VM_USER@$VM_IP:/opt/reploy/reploy-linux"

echo "==> Deploying on VM..."
ssh "$VM_USER@$VM_IP" bash << 'EOF'
  set -e
  cd /opt/reploy
  git pull
  docker compose up -d --build --no-deps reploy
  echo "==> Done"
EOF

rm -f reploy-linux
echo "==> Deploy complete"
