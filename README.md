# s3s2

Safe and secure (s2) file sharing with S3.

Simple Secure S3 Share - Share files securely with S3.

The purpose of s3s2 is to make it easy for end users that may not be familiar with S3 or GPG to do the right thing when they share files.  The tool makes some simplifying assumptions that help to make it easy and provides just enough control to prevent bad things from happening.

Anything that can be done in this tool could be done by hand with tools like keybase + the S3 CLI.  The value this project intends to bring is to have a single small distributable binary (for any mainstream platform) that just does the work.

## Running s3s2

`s3s2 share --bucket sharing-bucket --key https://jemurai.com/.well_known/id_rsa.pub --directory .`

This will take the current working directory, list the files to build a manifest.json file, put them all in a Zip file, encrypt that with the public key of the receiving party (so that only they, with the private key can read it) and drop the file in an S3 bucket.

## An Example of Using S3 as an Organization that Wants to Receive Incoming Data Securely

1. Set up your AWS KMS key, S3 bucket and GPG key (if desired).
1. Run `s3s2 config` to build your reference config.
1. Run `s3s2 share --directory /dir/to/share --org OrgName`

## Setting Up AWS

See the `clisetup.sh` script for snippets you can use that might be helpful when setting up the AWS environment and GPG keys.

Other references:

- https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html
- https://docs.aws.amazon.com/cli/latest/reference/kms/create-key.html

## Building a Configuration

s3s2 makes it easy to build a default configuration for clients to use.  By using this, you can distribute a JSON file with your organization's default information so that using the tool is very easy.

```bash
om:s3s2 mk$ ./s3s2 config --file ~/s3s2-demo.json
Using config file: /Users/mk/.s3s2.json
Please specify a bucket.
> demo-incoming
Please specify a region.
> us-east-1
Please specify an org.
> Jemurai
Please specify a working directory.
> ~/Desktop/s3s2/
Please specify a file prefix (nothing sensitive).
> jemurai_
Please specify a public key to use (file path or url).
> https://s3s2.jemurai.com/.well_known/s3s2-pub.asc
Your config was written to /Users/mk/s3s2-demo.json . You can invoke with s3s2 --config /Users/mk/s3s2-demo.json
```

# Get Help

Feel free to create issues on [the project](https://github.com/jemurai/s3s2) to ask questions or come find us on [Gitter](https://gitter.im/jemurai-oss/s3s2) to have a chat. 
