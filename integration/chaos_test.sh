#!/bin/sh
set -e

# Chaos Monkey for Pandacea Protocol
# Randomly kills and restarts the agent-backend service

SERVICES="agent-backend"

while true; do
  # Sleep for a random interval between 10 and 30 seconds
  SLEEP_TIME=$((10 + RANDOM % 21))
  echo "[chaos-monkey] Sleeping for $SLEEP_TIME seconds..."
  sleep $SLEEP_TIME

  # Randomly select a service (for now, only agent-backend)
  TARGET_SERVICE="agent-backend"
  echo "[chaos-monkey] Killing $TARGET_SERVICE..."
  docker-compose kill $TARGET_SERVICE

  # Sleep for a few seconds before restart
  RESTART_SLEEP=$((3 + RANDOM % 5))
  echo "[chaos-monkey] Waiting $RESTART_SLEEP seconds before restart..."
  sleep $RESTART_SLEEP

  echo "[chaos-monkey] Restarting $TARGET_SERVICE..."
  docker-compose up -d $TARGET_SERVICE

done 