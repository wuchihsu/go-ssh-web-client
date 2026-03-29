# Go SSH Web Client

A simple SSH web client using Go, WebSocket and Xterm.js.

## Getting Started

After cloning the project, copy [config.toml.sample](config.toml.sample) to `config.toml`. Modify the host, port, user and password attributes to match the target SSH server in `config.toml`.

There are two ways to install and run the project, using build script and using Docker.

### 1. Script

Prerequisites:
- [Go](https://golang.org/doc/install)
- [Node.js](https://nodejs.org/en/download/)
- [npm](https://www.npmjs.com/get-npm)

First, execute script to run frontend dev server:

```bash
./scripts/frontend-dev.sh
```

Then run the backend program:

```bash
./scripts/backend-dev.sh
```

Now, the HTTP server is running on port 8080, open http://localhost:8080 to use it (use http at your own risk).

### 2. Docker

Prerequisites:
- [Docker](https://docs.docker.com/get-docker/)

In current directory, run the prebuilt image:

```bash
./scripts/build-docker.sh
```

Then, run the container:

```bash
./scripts/docker-dev.sh
```

Now, the HTTP server is running on port 8080, open http://localhost:8080 to use it (use http at your own risk).
