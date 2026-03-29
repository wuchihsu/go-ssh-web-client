FROM golang:1.25.3-alpine AS go-env
WORKDIR /go/src/github.com/wuchihsu/go-ssh-web-client/
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM node:18.20-alpine AS node-env
WORKDIR /usr/src/
COPY front ./front
RUN cd front && npm install --production

# FROM alpine:latest
FROM scratch

# set workdir to non-root user home
WORKDIR /root/

# copy app and front to non-root user home
COPY --from=go-env /go/src/github.com/wuchihsu/go-ssh-web-client/app ./
COPY --from=node-env /usr/src/front ./front

# expose and run app as non-root user
EXPOSE 8080/tcp
ENTRYPOINT ["./app"]
