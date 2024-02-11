#!/usr/bin/env sh

if [ -z "$PORT" ]; then
  PORT=3000
fi
export PORT

if [ ! -f data/0_0.tigerbeetle ]; then
  echo "[*] Setting up tigerbeetle database..."
  /tigerbeetle format --cluster=0 --replica=0 --replica-count=1 /data/0_0.tigerbeetle
fi

echo "[*] Starting tigerbeetle on port $PORT..."
/tigerbeetle start --addresses=0.0.0.0:$PORT --cache-grid=512MB /data/0_0.tigerbeetle 
