#!/bin/bash
cd /usr/local/go
git fetch && git checkout go1.6.2
cd src
./make.bash
