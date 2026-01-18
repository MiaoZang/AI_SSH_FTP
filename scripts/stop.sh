#!/bin/bash

# stop.sh - Stop the AI SSH/FTP Proxy Service
cd "$(dirname "$0")/.."

PID_FILE="server.pid"

if [ ! -f "$PID_FILE" ]; then
    echo "PID file not found. Is the service running?"
    exit 1
fi

PID=$(cat "$PID_FILE")

if ps -p "$PID" > /dev/null 2>&1; then
    echo "Stopping service (PID: $PID)..."
    kill "$PID"
    
    # Wait for process to exit
    TIMEOUT=5
    while [ $TIMEOUT -gt 0 ]; do
        if ps -p "$PID" > /dev/null 2>&1; then
            sleep 1
            ((TIMEOUT--))
        else
            break
        fi
    done
    
    if ps -p "$PID" > /dev/null 2>&1; then
        echo "Service did not stop gracefuly. Force killing..."
        kill -9 "$PID"
    fi
    
    echo "Service stopped."
else
    echo "Process $PID not found."
fi

rm "$PID_FILE"
