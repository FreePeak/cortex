#!/bin/bash
# Script to kill processes using a specific port

PORT=${1:-8080}

echo "Attempting to find and kill processes using port $PORT..."

# For Mac/Linux
if command -v lsof >/dev/null 2>&1; then
  PIDS=$(lsof -ti:$PORT)
  if [ -n "$PIDS" ]; then
    echo "Found processes: $PIDS"
    kill -9 $PIDS
    echo "Processes killed"
  else
    echo "No processes found using port $PORT"
  fi
# For Windows
elif command -v netstat >/dev/null 2>&1 && command -v taskkill >/dev/null 2>&1; then
  PID=$(netstat -ano | grep ":$PORT" | awk '{print $5}' | head -n 1)
  if [ -n "$PID" ]; then
    echo "Found process: $PID"
    taskkill /F /PID $PID
    echo "Process killed"
  else
    echo "No processes found using port $PORT"
  fi
else
  echo "Could not find appropriate command to check for processes"
  exit 1
fi

echo "Done!" 