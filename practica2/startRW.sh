#!/bin/sh

if [ $# -ne 1 ]; then
  echo "Usage: ./startRW.sh <mode>"
  exit 1
fi
echo $0 $1

go run gestorfichero.go 4 ./ms/users.txt &

if [ $1=="GoVec" ]; then 
    go run lectorGoVec.go 1 3 ./ms/users.txt &
    go run lectorGoVec.go 2 3 ./ms/users.txt &
    go run escritorGoVec.go 3 3 ./ms/users.txt &
else 
    go run lector.go 1 3 ./ms/users.txt &
    go run lector.go 2 3 ./ms/users.txt &
    go run escritor.go 3 3 ./ms/users.txt &
fi

exit 0
