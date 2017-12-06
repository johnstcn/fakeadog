#!/usr/bin/env bash

CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w' .
docker build -t johnstcn/fakeadog .