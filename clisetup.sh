# Show the kms key
aws-vault exec jemurai-mkonda -- aws kms create-key --description "Testing s3s2 2"
aws-vault exec jemurai-mkonda -- aws kms describe-key --key-id (key_id)

# S3
aws-vault exec jemurai-mkonda-admin -- aws s3 mb s3://s3s2-demo1

# Show what is in S3
aws-vault exec jemurai-mkonda -- aws s3 ls
aws-vault exec jemurai-mkonda -- aws s3 ls s3s2-demo



## s3s2


## gpg  http://irtfweb.ifa.hawaii.edu/~lockhart/gpg/gpg-cs.html

gpg --gen-key
gpg --list-keys
gpg2 --export --armor [KEY ID] > /tmp/pubKey.asc

(Point to this key file in the command line)

## Decrypt
gpg -o azip.zip -d s3s2_14dc40c8-9c9d-47e1-a9be-104d12abc0a2.zip.gpg

