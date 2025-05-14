#!/bin/bash
# Script to run the PocketBase integration example with various configurations

# Default values
PORT=8080
DATA_DIR="./pb_data"
KILL_EXISTING=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -p|--port)
      PORT="$2"
      shift 2
      ;;
    -d|--data)
      DATA_DIR="$2"
      shift 2
      ;;
    -k|--kill)
      KILL_EXISTING=true
      shift
      ;;
    -h|--help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  -p, --port PORT       Server port (default: 8080)"
      echo "  -d, --data DIR        Data directory (default: ./pb_data)"
      echo "  -k, --kill            Kill any existing processes using the specified port"
      echo "  -h, --help            Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Kill existing processes if requested
if [ "$KILL_EXISTING" = true ]; then
  echo "Checking for processes using port $PORT..."
  if command -v lsof >/dev/null 2>&1; then
    PIDS=$(lsof -ti:$PORT)
    if [ -n "$PIDS" ]; then
      echo "Killing processes: $PIDS"
      kill -9 $PIDS
    else
      echo "No processes found using port $PORT"
    fi
  else
    echo "lsof command not found, cannot check for existing processes"
  fi
fi

# Run the server
echo "Starting PocketBase integration example on port $PORT with data directory $DATA_DIR"
go run main.go --port $PORT --data $DATA_DIR

# Exit with the same status as the server
exit $? 