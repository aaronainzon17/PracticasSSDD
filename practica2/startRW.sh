#!/bin/sh

if [ $# -ne 1 ]; then
  echo "Usage: ./startRW.sh <govec/go>"
  exit 1
fi

if [ $1 != "go" ] || [ $1 != "govec" ];then 
  go run gestorfichero.go 4 ./ms/users.txt &
fi

if [ $1 = "govec" ]; then 
  go run lectorGoVec.go 1 3 ./ms/users.txt &
  go run lectorGoVec.go 2 3 ./ms/users.txt &
  go run escritorGoVec.go 3 3 ./ms/users.txt &
elif [ $1 = "go" ]; then
  go run lector.go 1 3 ./ms/users.txt &
  go run lector.go 2 3 ./ms/users.txt &
  go run escritor.go 3 3 ./ms/users.txt &
fi

exit 0
