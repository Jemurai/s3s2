# Demo of Expected Use of S3S2

1. Set up a user with AWS Access to create a key and S3 bucket.
1. Create a kms key for S3S2 to use.
1. Assign permissions to use the key appropriately.
1. Create an S3 bucket for S3S2 to use.
1. Run `s3s2 config` to build the default config file
1. Run `s3s2 share --bucket s3s2-demo --directory ~/Desktop/s3s2/` (Fails because no encryption)
1. Run `s3s2 share --bucket s3s2-demo --directory ~/Desktop/s3s2/ --awskey 933c...9d9 --region us-east-1` (Succeed)
1. Pull file from console to show encryption
1. Notice manifest
1. Run `s3s2 share --bucket s3s2-demo --directory ~/Desktop/s3s2/ --awskey 933c...9d9 --region us-east-1 --pubkey ./s3s2-pub.asc` (Succeed)
1. Pull file to show gpg encryption
1. GPG decrypt
1. Run `s3s2 share --bucket s3s2-demo --directory ~/Desktop/s3s2/ --awskey 933c...9d9 --region us-east-1 --pubkey https://s3s2.jemurai.com/.well_known/s3s2-pub.asc` (Succeed)
1. Run `s3s2 share`

## Cleanup

1. `aws-vault exec jemurai-mkonda-admin -- aws s3 rm s3://s3s2-demo/ --recursive`