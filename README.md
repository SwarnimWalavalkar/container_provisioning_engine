# Container Provisioning Engine

API service to provision and manage docker containers and expose them with TLS behind a Traefik reverse proxy

- Provision containers given an image tag
  - Supports pulling from authenticated registries with a username and password
- Expose on a subdomain with a letsencrypt SSL cert

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
