# go-ssh-web-client

A simple SSH web client using Go, WebSocket and Xterm.js.

## Getting Started

There are two ways to install and run the project, using Go and using Docker.

### Go

After cloning the project, go into its `front` folder and install npm packages:

```bash
cd go-ssh-web-client/front
npm install --production
```

Then go back to main folder, add configuration file and modify it:

```bash
cd ..
cp config.toml.sample config.toml
vim config.toml
```

Modify the host, port, user and password attributes to match the target SSH server, then save the file. Finally, run the program:

```bash
go run .
```

Now, the HTTP server is running on port 8080, open http://localhost:8080 to use it (use http at your own risk).

### Docker

First, prepare a configuration file, like [config.toml.sample](config.toml.sample). After preparing `config.toml` in current directory, run the prebuilt image:

```bash
docker run --name go-ssh -d \
    -v `pwd`/config.toml:/root/config.toml \
    -p 8080:8080 \
    wuchihsu/go-ssh-web-client
```

Now, the HTTP server is running on port 8080, open http://localhost:8080 to use it (use http at your own risk).
