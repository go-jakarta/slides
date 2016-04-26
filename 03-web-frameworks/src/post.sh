#!/bin/bash

set -x

curl -v --data 'name=matt%20mikkelsen' http://localhost:8000/add
