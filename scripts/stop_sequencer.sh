#!/bin/sh

set -ex

curl \
    --location http://node:8545 \
    --header 'Content-Type: application/json' \
    --data '{
    "jsonrpc":"2.0",
    "method":"admin_stopSequencer",
    "params":[
    ],
    "id": 1
}' | jq

curl \
    --location http://batcher:8545 \
    --header 'Content-Type: application/json' \
    --data '{
    "jsonrpc":"2.0",
    "method":"admin_stopBatcher",
    "params":[
    ],
    "id": 1
}' | jq

curl \
    --location http://node:8545 \
    --header 'Content-Type: application/json' \
    --data '{
    "jsonrpc":"2.0",
    "method":"admin_sequencerActive",
    "params":[
    ],
    "id": 1
}' | jq
