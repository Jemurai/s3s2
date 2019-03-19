# s3s2

Safe and secure (s2) file sharing with S3.

Simple Secure S3 Share - Share files securely with S3.

The purpose of s3s2 is to make it easy for end users that may not be familiar with S3 or GPG to do the right thing when they share files.  The tool makes some simplifying assumptions that help to make it easy and provides just enough control to prevent bad things from happening.

Anything that can be done in this tool could be done by hand with tools like keybase + the S3 CLI.  The value this project intends to bring is to have a single small distributable binary (for any mainstream platform) that just does the work.

## Running s3s2

`s3s2 share -b sharing-bucket -k https://jemurai.com/.well_known/id_rsa.pub .`

This will take the current working directory, list the files to build a manifest.json file, put them all in a Zip file, encrypt that with the public key of the receiving party (so that only they, with the private key can read it) and drop the file in an S3 bucket.
