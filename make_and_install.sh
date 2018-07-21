#!/bin/sh
echo "build go binary file..."
go build
echo "copy it to /bin"
cp -f home_ddns /bin/
