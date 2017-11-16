#!/bin/bash
go build -o kumareport *.go
kill $(ps aux | grep 'kumareport' | awk '{print $2}')
nohup ./kumareport > access.log &