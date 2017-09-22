#!/bin/bash
set -e

go build

./frontend -collector='127.0.0.1:7701' -backend='http://127.0.0.1:8200'