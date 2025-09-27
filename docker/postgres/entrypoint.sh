#!/bin/bash

# Start the second process
docker-entrypoint.sh "$@"
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start second process: postgres engine: $status"
  exit $status
fi

# The container exits with an error
# if it detects that either of the processes has exited.
# Otherwise it loops forever, waking up every 60 seconds

while sleep 60; do
  ps aux |grep postgres |grep -q -v grep
  PROCESS_1_STATUS=$?
  # If the greps above find anything, they exit with 0 status
  # If they are not both 0, then something is wrong
  if [ $PROCESS_1_STATUS -ne 0 ]; then
    exit 1
  fi
done
