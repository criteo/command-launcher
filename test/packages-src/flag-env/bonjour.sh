#!/bin/sh

echo "bonjour!"
echo $CL_LOG_LEVEL
echo $CL_FLAG_NAME
echo $CL_FLAG_LANGUAGE
echo $CL_ARG_1
echo "number of args: $CL_NARGS"

echo "cola flag: $COLA_FLAG_NAME"
echo "cola flag: $COLA_FLAG_LANGUAGE"
echo "cola arg: $COLA_ARG_1"
echo "cola nargs: $COLA_NARGS"


echo "$@"
