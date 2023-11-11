#!/bin/sh

set -ex

HSH=$1

curl \
    --location http://node:8545 \
    --header 'Content-Type: application/json' \
    --data '{
    "jsonrpc":"2.0",
    "method":"admin_startSequencer",
    "params":[
        "'$HSH'"
    ],
    "id": 1
}' | jq

curl \
    --location http://batcher:8545 \
    --header 'Content-Type: application/json' \
    --data '{
    "jsonrpc":"2.0",
    "method":"admin_startBatcher",
    "params":[
    ],
    "id": 1
}' | jq
