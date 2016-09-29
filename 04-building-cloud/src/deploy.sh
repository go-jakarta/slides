#!/bin/bash

# chunk assets
go-bindata -o assets/assets.go -ignore=\\.go\$ assets/...

# build
go build

# deploy
rsync -avP myBinary user@host:/path/to/bin

# kick
ssh user@host '/path/to/restart.sh'
