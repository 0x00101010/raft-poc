#!/bin/sh

PROJECT_NAME=raft-poc

# clean up first
docker ps -a --filter "name=$PROJECT_NAME" -q | xargs -r docker rm
docker volume ls --filter "name=$PROJECT_NAME" -q | xargs -r docker volume rm
docker network ls --filter "name=$PROJECT_NAME" -q | xargs -r docker network rm

docker-compose -f docker-compose-bootstrap.yml up -d --build --force-recreate

# Add other participants
raftadmin localhost:50051 add_voter NodeB elector2:50052 0
# most likely the leader will remain the same, otherwise we should use
# >> raftadmin localhost:50052 add_voter NodeC elector3:50053 0
raftadmin localhost:50051 add_voter NodeC elector3:50053 0
