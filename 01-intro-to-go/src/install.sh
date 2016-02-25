#!/bin/bash

sudo -s
cd /usr/local/

export GOROOT_BOOTSTRAP=/usr/local/go-1.6

git clone https://go.googlesource.com/go
cd go/src && ./all.bash

export PATH=$PATH:/usr/local/go/bin
