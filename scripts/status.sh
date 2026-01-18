#!/bin/bash

# status.sh - Check status of the AI SSH/FTP Proxy Service
cd "$(dirname "$0")/.."

PID_FILE="server.pid"

if [ ! -f "$PID_FILE" ]; then
    echo "Status: Stopped (PID file not found)"
    exit 0
fi

PID=$(cat "$PID_FILE")

if ps -p "$PID" > /dev/null 2>&1; then
    echo "Status: Running (PID: $PID)"
    echo "Process Info:"
    ps -fp "$PID"
else
    echo "Status: Stopped (PID file exists but process missing)"
    # Optional: cleanup
    # rm "$PID_FILE"
fi
