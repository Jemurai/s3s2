#!/bin/bash

aws-vault exec jemurai-mkonda-admin -- go run main.go decrypt --debug true --bucket s3s2-demo --region us-east-1 --destination ./test/s3s2/s3s2-down/ --my-private-key ./test/s3s2/s3s2-keys/test.privkey --my-public-key ./test/s3s2/s3s2-keys/test.pubkey --file $1