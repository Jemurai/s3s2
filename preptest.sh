#!/bin/bash

echo "Cleaning up files in $PWD"
rm -rf ./test/s3s2/s3s2-down/*
rm -rf ./test/s3s2/s3s2-up/*
aws-vault exec jemurai-mkonda-admin -- aws s3 rm --recursive s3://s3s2-demo
cp ./test/s3s2/s3s2-data/*.csv ./test/s3s2/s3s2-up/

echo "Done"