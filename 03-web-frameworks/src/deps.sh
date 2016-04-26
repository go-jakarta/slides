#!/bin/bash

for i in "github.com/mattn/go-sqlite3 github.com/gin-gonic/gin github.com/labstack/echo/... goji.io goji.io/pat golang.org/x/net/context github.com/revel/revel github.com/revel/cmd/revel"; do
  go get -u -v -x $i
done
