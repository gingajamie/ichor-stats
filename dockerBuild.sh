#!/bin/bash

docker stop csgo-stats
docker rm csgo-stats
docker rmi csgo-stats:latest --force
docker build -t csgo-stats:latest .
docker run --publish 5001:5001 --detach --name csgo-stats csgo-stats:latest
