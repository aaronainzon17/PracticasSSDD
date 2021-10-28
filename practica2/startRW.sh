#!/bin/sh

if [ $# -ne 1 ]; then
  echo "Usage: ./startRW.sh <mode>"
  exit 1
fi
echo $0 $1

go run gestorfichero.go 5 ./ms/users.txt &
if [ $1=="GoVec" ]; then 
    go run escritorGoVec.go 1 4 ./ms/users.txt &
    go run escritorGoVec.go 2 4 ./ms/users.txt &
    go run escritorGoVec.go 3 4 ./ms/users.txt &
    go run escritorGoVec.go 4 4 ./ms/users.txt &
else 
    go run lector.go 1 4 ./ms/users.txt &
    go run lector.go 2 4 ./ms/users.txt &
    go run lector.go 3 4 ./ms/users.txt &
    go run lector.go 4 4 ./ms/users.txt &
fi

exit 0
