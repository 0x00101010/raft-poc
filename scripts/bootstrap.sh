#!/bin/sh

PROJECT_NAME=raft-poc
FILE_NAME=docker-compose.yml

# stop running services if any
docker-compose down

# clean up first
docker ps -a --filter "name=$PROJECT_NAME" -q | xargs -r docker rm
docker volume ls --filter "name=$PROJECT_NAME" -q | xargs -r docker volume rm
docker network ls --filter "name=$PROJECT_NAME" -q | xargs -r docker network rm

# change elector1 BOOTSTRAP to true
awk '!found && /BOOTSTRAP=false/ {sub(/BOOTSTRAP=false/, "BOOTSTRAP=true"); found=1} 1' $FILE_NAME >temp && mv temp $FILE_NAME

# start raft servers
docker-compose up -d --build --force-recreate

# change elector 1 BOOTSTRAP to false
awk '!found && /BOOTSTRAP=true/ {sub(/BOOTSTRAP=true/, "BOOTSTRAP=false"); found=1} 1' $FILE_NAME >temp && mv temp $FILE_NAME

sleep 5
# Add other participants
raftadmin localhost:50051 add_voter NodeB elector2:50052 0
# most likely the leader will remain the same, otherwise we should use
# >> raftadmin localhost:50052 add_voter NodeC elector3:50053 0
raftadmin localhost:50051 add_voter NodeC elector3:50053 0
