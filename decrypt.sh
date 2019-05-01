#!/bin/bash

aws-vault exec jemurai-mkonda-admin -- go run main.go decrypt --debug true --bucket s3s2-demo --region us-east-1 --destination ./test/s3s2/s3s2-down/ --privkey ./test/s3s2/s3s2-keys/test1.privkey --pubkey ./test/s3s2/s3s2-keys/test1.pubkey --file $1