# Container Provisioning Engine

API service to provision, manage, and expose Docker containers.

It's written in Go and uses the Docker SDK for container management, golang channels and goroutines for concurrency, and Traefik as a reverse proxy to expose the containers.

## Features

- Provisioning of Docker containers from a specified image tag.
- Support for pulling images from authenticated registries using a username and password.
- Exposing provisioned containers on a subdomain with a Let's Encrypt SSL certificate.
- An async task queue system for managing deployment tasks.

## Setting Up

Setup .env from the default values in [.env.development](.env.development)

```
cp .env.development .env
```

Start dependencies in docker

```
make dx-start
```

Build and run the application

```
make start
```

# Improvement Ideas

Robustness
- Improve fault-tolerance
  - Add Persistence to the queue
  - Add a recovery mechanism to queue system in case of crashes and other irrecoverable failures
- Add multi-tenancy

Features
- Access to the realtime docker container logs
- Zero downtime updates to deployments
- Horizontally scalable deployments

Admin
- API rate limiting
- Enforce Resource limits for individual deployments
- Track compute resources used by each user (tenant)
  - Limit compute resources allocated to each tenant
