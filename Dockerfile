FROM golang:1.17-alpine AS go-env
WORKDIR /go/src/github.com/wuchihsu/go-ssh-web-client/
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app .

FROM node:14.17-alpine AS node-env
WORKDIR /usr/src/
COPY front ./front
RUN cd front && npm install --production

# FROM alpine:latest
FROM scratch
WORKDIR /root/
COPY --from=go-env /go/src/github.com/wuchihsu/go-ssh-web-client/app ./
COPY --from=node-env /usr/src/front ./front
EXPOSE 8080/tcp
ENTRYPOINT ["./app"]
