#!/bin/bash

git add -A
git commit -m "deploy: $(date '+%Y-%m-%d %H:%M')" || echo "Nothing new to commit"
git push origin master


set -e
cd /opt/reploy
git pull
docker compose up -d --build --no-deps reploy