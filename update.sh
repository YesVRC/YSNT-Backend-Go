#!/bin/bash

git pull
go build .
cp -f ysnt-backend.service /etc/systemd/system/ysnt-backend.service
systemctl daemon-reload
systemctl restart ysnt-backend.service