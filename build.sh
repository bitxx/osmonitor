#!/usr/bin/env bash

# macos arm64 => new macos (M1 M2 M3...)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o osmonitor-client client/client.go
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o osmonitor-server server/server.go
mkdir -p ./osmonitor/client/
mkdir -p ./osmonitor/server/
mv osmonitor-client ./osmonitor/client/
mv osmonitor-server ./osmonitor/server/
cp client/settings.yml ./osmonitor/client/
cp server/settings.yml ./osmonitor/server/
tar czvf "osmonitor-macos-arm64".tar.gz ./osmonitor
rm -rf ./osmonitor

sleep 3

# macos amd64
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o osmonitor-client client/client.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o osmonitor-server server/server.go
mkdir -p ./osmonitor/client/
mkdir -p ./osmonitor/server/
mv osmonitor-client ./osmonitor/client/
mv osmonitor-server ./osmonitor/server/
cp client/settings.yml ./osmonitor/client/
cp server/settings.yml ./osmonitor/server/
tar czvf "osmonitor-macos-amd64".tar.gz ./osmonitor
rm -rf ./osmonitor

sleep 3


# 交叉编译windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o osmonitor-client.exe client/client.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o osmonitor-server.exe server/server.go
mkdir -p ./osmonitor/client/
mkdir -p ./osmonitor/server/
mv osmonitor-client.exe ./osmonitor/client/
mv osmonitor-server.exe ./osmonitor/server/
cp client/settings.yml ./osmonitor/client/
cp server/settings.yml ./osmonitor/server/
tar czvf "osmonitor-windows".tar.gz ./osmonitor
rm -rf ./osmonitor

sleep 3


# 交叉编译linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o osmonitor-client client/client.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o osmonitor-server server/server.go
mkdir -p ./osmonitor/client/
mkdir -p ./osmonitor/server/
mv osmonitor-client ./osmonitor/client/
mv osmonitor-server ./osmonitor/server/
cp client/settings.yml ./osmonitor/client/
cp server/settings.yml ./osmonitor/server/
tar czvf "osmonitor-linux-amd64".tar.gz ./osmonitor
rm -rf ./osmonitor

sleep 3
