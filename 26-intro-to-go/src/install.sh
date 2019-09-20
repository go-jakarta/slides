#!/bin/bash

sudo -s
cd /usr/local/

export GOROOT_BOOTSTRAP=/usr/local/go-1.13

git clone https://go.googlesource.com/go
cd go/src && git checkout go1.13
./all.bash

export PATH=$PATH:/usr/local/go/bin
