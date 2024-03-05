#!/bin/sh

docker network create traefik_default 2>/dev/null || true

docker compose -f docker/docker-compose.yml --env-file .env up -d
