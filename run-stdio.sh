#! /bin/bash

go build -o bin/echo-stdio-server cmd/echo-stdio-server/main.go
./bin/echo-stdio-server