#!/bin/sh

NodeA=base-sepolia-devnet-sequencer-donotuse.cbhq.net:50050
NodeB=10.242.49.206:50050
NodeC=10.242.176.235:50050

sequencer_active() {
    curl -s \
        --location $1 \
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
    raftadmin $1 leader
}

echo "NodeA status:"
echo "sequencer active: $(sequencer_active $NodeA)"
echo "raft leader: $(raft_leader $NodeA)"

echo ""
echo "NodeB status:"
echo "NodeB sequencer active: $(sequencer_active $NodeB)"
echo "raft leader: $(raft_leader $NodeB)"

echo ""
echo "NodeC status:"
echo "NodeC sequencer active: $(sequencer_active $NodeC)"
echo "raft leader: $(raft_leader $NodeC)"
