#!/bin/bash
dir=$(dirname "$0")
cd "$dir"
CGO_ENABLED=0 go build

