#!/usr/bin/env bash

CGO_ENABLED=0 GOOS=linux go build  -ldflags '-w' github.com/johnstcn/fakeadog/cmd/fakeadog
docker build -t johnstcn/fakeadog .