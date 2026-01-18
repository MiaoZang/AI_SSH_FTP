#!/bin/bash

# start.sh - Start the AI SSH/FTP Proxy Service
cd "$(dirname "$0")/.."

BINARY_NAME="ssh-ftp-proxy"
LOG_FILE="config/server.log"
PID_FILE="server.pid"

if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo "Service is already running (PID: $PID)"
        exit 1
    else
        echo "PID file exists but process not found. Cleaning up."
        rm "$PID_FILE"
    fi
fi

echo "Starting $BINARY_NAME..."
nohup ./$BINARY_NAME > "$LOG_FILE" 2>&1 &
PID=$!
echo "$PID" > "$PID_FILE"

echo "Service started (PID: $PID)"
echo "Logs: $LOG_FILE"
