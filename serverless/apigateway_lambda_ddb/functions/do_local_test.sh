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
    echo "      $script_name '{\"body\":\"{\\\"name\\\":\\\"Cow\\\",\\\"comment\\\":\\\"Sample comment!\\\",\\\"chatRoom\\\":\\\"101\\\"}\"}'"
    echo ""
    echo "      $script_name @put-chat-records/sample_body.json"
    echo ""
    echo "      $script_name '{\"queryStringParameters\":{\"chatroom\":\"101\"}}'"
    echo ""
    echo "      $script_name @get-chat-records/sample_query_string.json"
    echo ""
    exit 0
fi

curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d "$@"
echo ""
