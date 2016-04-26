#!/bin/bash

# get revel framework and cli tool
go get -u github.com/revel/revel
go get -u github.com/revel/cmd/revel

# create new revel app
revel new github.com/myaccount/my-app
