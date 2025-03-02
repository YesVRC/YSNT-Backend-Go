#!/bin/bash

export GOCACHE=/usr/go/cache
git pull
go build .
cp -f ysnt-backend.service /etc/systemd/system/ysnt-backend.service
systemctl daemon-reload