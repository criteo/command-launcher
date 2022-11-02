#!/bin/bash

SCRIPT_DIR=${SCRIPT_DIR:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}
# replace \ to / for windows
SCRIPT_DIR=${SCRIPT_DIR//\\//}
echo "integration test directory: $SCRIPT_DIR"

EXIT_CODE=0
TEST_COUNT=0
FAILURE_COUNT=0

# create output folder
OUTPUT_DIR=$SCRIPT_DIR/output
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

##
# build the binary
##
cd $SCRIPT_DIR
go build -o $OUTPUT_DIR/cl -ldflags='-X main.version=integration-test -X main.buildNum=1 -X main.appName=cl -X "main.appLongName=Command Launcher"' $SCRIPT_DIR/../main.go

# specify the app home
export CL_HOME=$OUTPUT_DIR/home

if [ $# -ne 0 ]; then
  # in case pass test as arguments, run test from the arguments
  for test in "$@"; do
    echo "------------------------------------------------------------"
    echo "- test/integration/${test}.sh"
    echo "------------------------------------------------------------"

    let TEST_COUNT++

    OUTPUT_DIR=$OUTPUT_DIR \
    CL_PATH=$OUTPUT_DIR/cl \
    CL_HOME=$CL_HOME \
      $SCRIPT_DIR/integration/${test}.sh

    if [ $? -eq 0 ]; then
      echo "- PASS"
    else
      echo "- FAIL"
      EXIT_CODE=1
      let FAILURE_COUNT++
    fi

    echo ""
  done
else
  # otherwise run all tests in integration/ folder
  TESTS=$(ls $SCRIPT_DIR/integration/*.sh)
  echo "find all tests in integration folder:"
  echo "$TESTS"
  for f in $TESTS; do
    echo "------------------------------------------------------------"
    echo "- $f"
    echo "------------------------------------------------------------"

    let TEST_COUNT++

    OUTPUT_DIR=$OUTPUT_DIR \
    CL_PATH=$OUTPUT_DIR/cl \
    CL_HOME=$CL_HOME \
    $f

    if [ $? -eq 0 ]; then
      echo "- PASS"
    else
      echo "- FAIL"
      EXIT_CODE=1
      let FAILURE_COUNT++
    fi

    echo ""
  done
fi

##
# remove the output folder
##
echo "clean up"
rm -rf $OUTPUT_DIR

echo ""
echo "Total test suits $TEST_COUNT, failure $FAILURE_COUNT"
exit $EXIT_CODE
