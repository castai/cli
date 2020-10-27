#!/bin/bash

set -e

echo 'Test login'
./bin/cast login --token $CI_CAST_API_ACCESS_TOKEN

echo 'Test list clusters'
first_cluster_id=$(./bin/cast cluster list -o jsonpath='{[0].id}')

echo 'Test get cluster kubeconfig'
./bin/cast cluster get-kubeconfig $first_cluster_id

