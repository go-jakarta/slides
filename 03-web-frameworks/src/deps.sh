#!/bin/bash

for i in "github.com/mattn/go-sqlite3 github.com/gin-gonic/gin github.com/labstack/echo/... goji.io goji.io/pat golang.org/x/net/context"; do
  go get -u -v -x $i
done
