#!/bin/bash
cd /usr/local/go
git fetch && git checkout go1.7.1
cd src
./make.bash
