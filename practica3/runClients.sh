#!/bin/sh

if [ $# -ne 1 ]; then
  echo "Usage: ./runClients.sh <n>"
  exit 1
fi

for i in (1 $1); do
    sleep(5)
    go run client.go 155.210.154.210:30000 &
done
exit 0
