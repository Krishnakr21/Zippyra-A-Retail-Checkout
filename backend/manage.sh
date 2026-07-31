#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/.bin"
LOG_DIR="$SCRIPT_DIR/.logs"
PID_FILE="$SCRIPT_DIR/.pids"

cmd=$1

case $cmd in
  start-all)
    echo "=== Starting Zippyra Microservices ==="
    mkdir -p "$BIN_DIR" "$LOG_DIR"
    > "$PID_FILE"

    for svc_path in "$SCRIPT_DIR"/services/*; do
      if [ -d "$svc_path" ]; then
        svc_name=$(basename "$svc_path")
        echo "--> Building & Starting $svc_name..."
        go build -o "$BIN_DIR/$svc_name" "$svc_path"/*.go 2>/dev/null || go build -o "$BIN_DIR/$svc_name" "$svc_path"
        "$BIN_DIR/$svc_name" > "$LOG_DIR/$svc_name.log" 2>&1 &
        pid=$!
        echo "$pid" >> "$PID_FILE"
        echo "    Started $svc_name (PID: $pid)"
      fi
    done
    echo "=== All 24 Microservices Running! Logs in backend/.logs/ ==="
    ;;

  stop-all)
    echo "=== Stopping Zippyra Microservices ==="
    if [ -f "$PID_FILE" ]; then
      while read -r pid; do
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
          kill "$pid" 2>/dev/null || true
          echo "Stopped PID $pid"
        fi
      done < "$PID_FILE"
      rm -f "$PID_FILE"
    fi
    echo "=== All Microservices Stopped ==="
    ;;

  status)
    echo "=== Zippyra Microservices Status ==="
    if [ -f "$PID_FILE" ]; then
      while read -r pid; do
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
          echo "PID $pid: RUNNING"
        else
          echo "PID $pid: STOPPED"
        fi
      done < "$PID_FILE"
    else
      echo "No services running."
    fi
    ;;

  *)
    echo "Usage: ./manage.sh [start-all|stop-all|status]"
    ;;
esac
