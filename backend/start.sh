#!/bin/bash
set -e

go build

./backend -collector='127.0.0.1:7701'