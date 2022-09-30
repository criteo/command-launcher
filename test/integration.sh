#!/bin/bash

SCRIPT_DIR=${SCRIPT_DIR:-$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )}
echo "integration test directory: $SCRIPT_DIR"


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

    OUTPUT_DIR=$OUTPUT_DIR \
    CL_PATH=$OUTPUT_DIR/cl \
    CL_HOME=$CL_HOME \
      $SCRIPT_DIR/integration/${test}.sh

    if [ $? -eq 0 ]; then
      echo "- DONE"
    else
      echo "- FAILED"
      exit 1
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

    OUTPUT_DIR=$OUTPUT_DIR \
    CL_PATH=$OUTPUT_DIR/cl \
    CL_HOME=$CL_HOME \
      $f

    echo ""
  done
fi

##
# remove the output folder
##
echo "clean up"
rm -rf $OUTPUT_DIR
