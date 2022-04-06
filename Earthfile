VERSION 0.6
FROM golang:1.17-bullseye
WORKDIR /app

deps:
    COPY . .
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
    ARG GOOS=linux
    ARG GOARCH=amd64
    FROM +deps
    RUN GOOS=$GOOS GOARCH=$GOARCH go build -o restake-authz-ledger
    RUN tar -czvf restake-authz-ledger-$GOOS-$GOARCH.tar.gz restake-authz-ledger
    SAVE ARTIFACT ./restake-authz-ledger AS LOCAL dist/restake-authz-ledger
    SAVE ARTIFACT ./restake-authz-ledger-$GOOS-$GOARCH.tar.gz AS LOCAL dist/restake-authz-ledger-$GOOS-$GOARCH.tar.gz

build-release:
    BUILD +build --GOOS=linux --GOARCH=amd64
    BUILD +build --GOOS=linux --GOARCH=arm64
    BUILD +build --GOOS=darwin --GOARCH=amd64
    BUILD +build --GOOS=darwin --GOARCH=arm64
    BUILD +build --GOOS=windows --GOARCH=amd64
    BUILD +build --GOOS=windows --GOARCH=arm64
