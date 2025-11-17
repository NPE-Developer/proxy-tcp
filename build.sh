#!/bin/bash
set -e

if [ -n "$1" ]; then
    echo 'start building binaries'
    
    GOOS=windows GOARCH=386  go build -o ./bin-$1/proxy-win32.exe &
    GOOS=windows GOARCH=amd64  go build -o ./bin-$1/proxy-win64.exe &
    GOOS=linux GOARCH=386  go build -o ./bin-$1/proxy-linux32 &
    GOOS=linux GOARCH=amd64  go build -o ./bin-$1/proxy-linux64

    cp config.json ./bin-$1

    echo 'finished building binaries'
else 
    echo 'invalid version number'
fi
