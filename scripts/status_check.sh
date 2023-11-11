#!/bin/sh

NODE_ADDR=http://node
SERVER_ADDR=http://localhost

sequencer_active() {
    curl -s \
        --location $NODE_ADDR:8545 \
        --header 'Content-Type: application/json' \
        --data '{
        "jsonrpc":"2.0",
        "method":"admin_sequencerActive",
        "params":[
        ],
        "id":1
    }' |
        jq -r .result
}

raft_leader() {
    raftadmin $SERVER_ADDR:50050 leader
}

echo "Node status:"
echo "sequencer active: $(sequencer_active)"
echo "raft leader: $(raft_leader)"
