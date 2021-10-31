#!/bin/bash

PROC="$( pgrep 'lector' )"

for i in $PROC
do
        kill -9 ${i} > /dev/null 2>&1
done

PROC="$( pgrep 'escritor' )"

for i in $PROC
do
        kill -9 ${i} > /dev/null 2>&1
done

PROC="$( pgrep 'gestorfichero' )"

for i in $PROC
do
        kill -9 ${i} > /dev/null 2>&1
done

PROC="$( pgrep 'ra' )"

for i in $PROC
do
        kill -9 ${i} > /dev/null 2>&1
done

echo "Lectores y Escritores detenido "

exit 0
