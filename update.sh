#!/bin/bash

git pull &&
go build -o gologserver &&
./gologserver start
