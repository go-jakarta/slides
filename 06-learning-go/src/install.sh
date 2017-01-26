#!/bin/bash

sudo -s
cd /usr/local/

export GOROOT_BOOTSTRAP=/usr/local/go-1.7.4

git clone https://go.googlesource.com/go
cd go/src && git checkout go1.8rc2
./all.bash

export PATH=$PATH:/usr/local/go/bin
