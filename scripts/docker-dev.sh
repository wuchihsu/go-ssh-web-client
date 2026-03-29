#!/bin/bash
# docker down and delete container
docker stop go-ssh && docker rm go-ssh || echo "go-ssh container not found"

docker run --name go-ssh \
    -v `pwd`/config.toml:/root/config.toml \
    -p 8080:8080 \
    wuchihsu/go-ssh-web-client