#!/bin/sh

ZN=zn
ZN_DEV=znt

# build Zn
rm -f ./$ZN ./$ZN_DEV

echo '=== build [Zn] ==='
go build -o zn

echo '=== build [Zn-dev] ==='
go build -o znt ./cmd/znt
