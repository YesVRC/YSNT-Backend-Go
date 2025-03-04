#!/bin/bash

export GOCACHE=/usr/go/cache
git pull
go get
go build .
cp -f ysnt-backend.service /etc/systemd/system/ysnt-backend.service
systemctl daemon-reload
systemctl start ysnt-backend.service