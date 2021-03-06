#!/bin/bash

SRC=$(realpath $(cd -P "$( dirname "${BASH_SOURCE[0]}" )" && pwd ))

DB=$1

if [ -z "$DB" ]; then
  DB=postgres://postgres:P4ssw0rd@localhost
fi

set -e -x

usql $DB \
  -c "DROP DATABASE IF EXISTS booktest"

usql $DB \
  -f $SRC/schema.sql

xo $DB \
  -o $SRC
