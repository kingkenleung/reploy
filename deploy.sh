#!/bin/bash
set -e

VM_USER="root"
VM_IP="34.96.183.161"

echo "==> Pushing code to GitHub..."
git add -A
git commit -m "deploy: $(date '+%Y-%m-%d %H:%M')" || echo "Nothing new to commit"
git push origin master

echo "==> Deploying on VM..."
ssh "$VM_USER@$VM_IP" bash << 'EOF'
  set -e
  cd /opt/reploy
  git pull
  docker compose up -d --build --no-deps reploy
  echo "==> Done"
EOF

echo "==> Deploy complete"
