#!/bin/bash
source env.env
trap "kill 0" SIGINT
go run cmd/orchestrator/main.go &
sleep 2
go run cmd/agent/main.go &
sleep 3
echo Ctrl-c to exit
wait
