#!/bin/bash

echo "Checking data"
ls -al ./test/s3s2/s3s2-data
echo "Checking down"
ls -al ./test/s3s2/s3s2-down
echo "Checking up"
ls -al ./test/s3s2/s3s2-up
echo "Checking keys"
ls -al ./test/s3s2/s3s2-keys
echo "Checking s3"
aws-vault exec jemurai-mkonda-admin -- aws s3 ls --recursive s3s2-demo
