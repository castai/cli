#!/bin/bash

set -e

./bin/cast login --token $CI_CAST_API_ACCESS_TOKEN
./bin/cast cluster list -o jsonpath='{[0].name}'
