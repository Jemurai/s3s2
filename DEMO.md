# Demo of Expected Use of S3S2

## Setup and Background

1. Set up a user with AWS Access to create a key and S3 bucket.
1. Create a kms key for S3S2 to use.
1. Assign permissions to use the key appropriately.
1. Create an S3 bucket for S3S2 to use.
1. Run `s3s2 config` to build the default config file
1. Run `s3s2 share --bucket s3s2-demo --directory ~/Desktop/s3s2/` (Fails because no encryption - neither gpg or s3-kms)
1. Run `s3s2 share --bucket s3s2-demo --directory ~/Desktop/s3s2/ --awskey <kms-key-we-have-permissions-on> --region us-east-1` (Succeed)
1. Pull file from console to show encryption

## Demo Using Shell Scripts

1. `s3s2 genkey --keydir ./test/s3s2/s3s2-keys/ --keyprefix test1`  - Generates keys to use.  Note these keynames need to match the scripts.
1. `preptest.sh` - Cleans up the directories and S3 buckets used.
1. `sanity.sh` - Shows where the current files are.
1. `share.sh` - Shares the data up to S3 encrypted with the pgp files.
1. `decrypt.sh <filename>` - Pulls the files back down from S3 based on the manifest and decrypts.

## Demo By Hand

1. Generate keys to use:  `s3s2 genkey --keydir ./test/s3s2/s3s2-keys/ --keyprefix test1`
1. Set up data to use.  For the purpose of this demo, we'll put the data we want to process in test/s3s2/s3s2-up/
1. Share the directory: `aws-vault exec <role>s3s2 share --debug true --bucket <your-bucket> --region <your-region> --directory test/s3s2/s3s2-up/ --org YourOrg --prefix <optional-prefix> --pubkey test/s3s2/s3s2-keys/test1.pubkey --privkey test/s3s2/s3s2-keys/test1.privkey`  (Keys and directories per setup)
1. Check your bucket for the files:  `aws-vault exec <role> -- aws s3 ls <your-bucket>`
1. Download and decrypt the files:  `aws-vault exec <role> -- s3s2 decrypt --debug true --bucket <your-bucket> --region <your-region> --destination ./test/s3s2/s3s2-down/ --privkey ./test/s3s2/s3s2-keys/test1.privkey --pubkey ./test/s3s2/s3s2-keys/test1.pubkey --file <the manifest.json file from the share step>`
1. Check the local files:  `ls -al test/s3s2/s3s2-down/`

## Cleanup

1. `aws-vault exec <role> -- aws s3 rm s3://<your-test-bucket>/ --recursive`
1. Cleanup directories with s3s2-down and s3s2-up.