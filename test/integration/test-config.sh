#!/bin/bash

# required environment varibale
# CL_PATH
# CL_HOME
# OUTPUT_DIR
SCRIPT_DIR=${1:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}

##
# test config
##
echo "> test config"
RESULT=$($OUTPUT_DIR/cl config)
echo "$RESULT" | grep 'local_command_repository_dirname' | grep 'home' | grep -q 'current'
if [ $? -eq 0 ]; then
  # ok
  echo "OK"
else
  echo "KO - wrong config: local_command_repository_dirname"
  exit 1
fi

echo "> test get all config in json"
RESULT=$($OUTPUT_DIR/cl config --json)
VALUE=$(echo "$RESULT" | jq -r '.log_enabled')
if [ $VALUE == "false" ]; then
  echo "OK"
else
  echo "KO - incorrect config value"
  exit 1
fi

echo "> test get one config in json"
RESULT=$($OUTPUT_DIR/cl config log_enabled --json)
VALUE=$(echo "$RESULT" | jq -r '.log_enabled')
if [ $VALUE == "false" ]; then
  echo "OK"
else
  echo "KO - incorrect config value"
  exit 1
fi

echo "> test group_help_by_registry config exist, and default true"
RESULT=$($OUTPUT_DIR/cl config --json)
VALUE=$(echo "$RESULT" | jq -r '.group_help_by_registry')
if [ $VALUE == "true" ]; then
  echo "OK"
else
  echo "KO - incorrect config value"
  exit 1
fi

echo "> test concurrent config read and write"
# writers replace config.json while readers load it: no process may see a
# partial file ("Cannot read configuration file") and the file must stay valid
CONCURRENT_LOGS=$OUTPUT_DIR/concurrent-config
rm -rf $CONCURRENT_LOGS && mkdir -p $CONCURRENT_LOGS
PIDS=()
NAMES=()
for i in $(seq 50); do
  $OUTPUT_DIR/cl config log_level debug > /dev/null 2> $CONCURRENT_LOGS/write-$i.err &
  PIDS+=($!)
  NAMES+=(write-$i)
  $OUTPUT_DIR/cl config --json > /dev/null 2> $CONCURRENT_LOGS/read-$i.err &
  PIDS+=($!)
  NAMES+=(read-$i)
done
FAILED=0
for idx in "${!PIDS[@]}"; do
  if ! wait ${PIDS[$idx]}; then
    FAILED=$((FAILED+1))
    echo "  ${NAMES[$idx]} failed: $(cat $CONCURRENT_LOGS/${NAMES[$idx]}.err)"
  fi
done
if [ $FAILED -eq 0 ]; then
  echo "OK"
else
  echo "KO - $FAILED of ${#PIDS[@]} concurrent config commands failed"
  exit 1
fi

if jq . $CL_HOME/config.json > /dev/null; then
  echo "OK"
else
  echo "KO - config.json is not valid JSON after concurrent writes"
  exit 1
fi

RESULT=$($OUTPUT_DIR/cl config log_level)
if [ "$RESULT" == "debug" ]; then
  echo "OK"
else
  echo "KO - incorrect config value after concurrent writes: $RESULT"
  exit 1
fi

LEFTOVER=$(ls -a $CL_HOME | grep -c '\.tmp$')
if [ $LEFTOVER -eq 0 ]; then
  echo "OK"
else
  echo "KO - $LEFTOVER temporary config files left in $CL_HOME"
  exit 1
fi
