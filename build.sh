#! /bin/bash

# Build worker and master

cd /usr/local/GOPATH/src/github.com/yaltachen/crontab/master/main
env GOOS=linux GOARCH=amd64 go build -o ../../bin/master

cd /usr/local/GOPATH/src/github.com/yaltachen/crontab/worker/main
env GOOS=linux GOARCH=amd64 go build -o ../../bin/worker

