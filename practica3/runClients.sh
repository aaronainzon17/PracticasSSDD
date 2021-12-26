#!/bin/sh

if [ $# -ne 1 ]; then
  echo "Usage: ./runClients.sh <n>"
  exit 1
fi
max=$1

for i in `seq 1 $max`
do
    sleep 9
    go run client.go 155.210.154.210:30000 &
done
echo "ALL CLIENTS DEPLOYED"
exit 0
