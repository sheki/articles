#!/bin/bash -e

rm -rf docs
mkdir docs
go run ./cmd/main.go 
