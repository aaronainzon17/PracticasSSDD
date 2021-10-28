#!/bin/bash

kill_list="$( ps -eo pid,comm | sed -e 's/ \+//' | tr -s " " | grep 'lector' | cut -d' ' -f1)"

for i in $kill_list
do
        kill -9 ${i} > /dev/null 2>&1
done

kill_list="$( ps -eo pid,comm | sed -e 's/ \+//' | tr -s " " | grep 'escritor' | cut -d' ' -f1)"

for i in $kill_list
do
        kill -9 ${i} > /dev/null 2>&1
done

kill_list="$( ps -eo pid,comm | sed -e 's/ \+//' | tr -s " " | grep 'gestorfichero$' | cut -d' ' -f1)"

for i in $kill_list
do
        kill -9 ${i} > /dev/null 2>&1
done

kill_list="$( ps -eo pid,comm | sed -e 's/ \+//' | tr -s " " | grep 'ra$' | cut -d' ' -f1)"

for i in $kill_list
do
        kill -9 ${i} > /dev/null 2>&1
done

echo "SSDD Stopped"

exit 0
