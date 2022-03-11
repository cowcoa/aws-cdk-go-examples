#!/bin/bash

script_name=$(basename $0)
arg_count=$#

if test $arg_count -eq 0; then
    echo "Script for testing lambda function in local."
    echo ""
    echo "NOTE: Please run 'run_local_test.sh' script first."
    echo ""
    echo "Usage:"
    echo ""
    echo "      $script_name <json body>"
    echo ""
    echo "Examples:"
    echo ""
    echo "      $script_name -d '{\"name\":\"hello world!\"}'"
    echo ""
    echo "      $script_name -d @json_file"
    echo ""
    exit 0
fi

curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" "$@"