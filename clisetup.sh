# AWS
# We generally recommend using aws-vault (https://github.com/99designs/aws-vault)
# to help you manage sessions and logins across roles and accounts.

# Show the kms key
aws-vault exec <role> -- aws kms create-key --description "Testing s3s2 2"
aws-vault exec <role> -- aws kms describe-key --key-id (key_id)

# S3
aws-vault exec <role> -- aws s3 mb s3://s3s2-demo

# Show what is in S3
aws-vault exec jemurai-mkonda -- aws s3 ls
aws-vault exec jemurai-mkonda -- aws s3 ls s3s2-demo

## gpg  http://irtfweb.ifa.hawaii.edu/~lockhart/gpg/gpg-cs.html

gpg --gen-key
gpg --list-keys
gpg2 --export --armor [KEY ID] > /tmp/publicKey.asc

(Point to this key file in the command line)

## Decrypt
gpg -o azip.zip -d s3s2_14dc40c8-9c9d-47e1-a9be-104d12abc0a2.zip.gpg

# Now use s3s2