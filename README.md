# go-ssh-web-client

A simple SSH web client using Go, WebSocket and Xterm.js.

## Getting Started

After cloning the project, go into its `front` folder and install npm packages:

```bash
cd go-ssh-web-client/front
npm install --production
```

Then go back to main folder, add config file and modify it:

```bash
cd ..
cp config.toml.sample config.toml
vim config.toml
```

Modify the host, port, user and password attributes to match the target SSH server, then save the file. Finally, run the program:

```bash
go run .
```

Now, the web server is running on port 8080, open http://localhost:8080 to use it (use http at your own risk).
