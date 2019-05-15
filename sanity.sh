#!/bin/bash

echo "Checking data"
find ./test/s3s2/s3s2-data -type f
echo "Checking down"
find ./test/s3s2/s3s2-down -type f
echo "Checking up"
find ./test/s3s2/s3s2-up -type f
echo "Checking keys"
find ./test/s3s2/s3s2-keys -type f
echo "Checking s3"
aws-vault exec jemurai-mkonda-admin -- aws s3 ls --recursive s3s2-demo
