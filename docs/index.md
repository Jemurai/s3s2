# S3 Secure Sharing (S3S2)

## Use Case

S3 is a simple and useful tool for sharing information.  Time and time again, we see organizations using S3 in 
insecure ways.  This tool is built to try to make it very
easy to share files securely.


### What Does It Do?

Usually, the folks we're talking to have files in a directory
that they want to share.

What S3S2 makes easy is:
- Creating a manifest file describing the contents
- Zipping up all the files a directory
- Encrypting the zip file
- Sharing the encrypted file to S3
- Distributable configuration to make it easy to give a configuration to a partner that they can then use

The goal is that a partner can run a simple command like this:

`s3s2 share --config /your/config.json the_directory` 

### How Does It Work?

S3S2 is just using open and existing tools in a secure combination.
- S3 with AWS AES encryption
- File encryption using gpg

## Installation

We produce packages for common platforms that you can just download from 

## Running the Tool

This section describes how to use the tool.  Skip to Running to jump to the most useful stuff.

### Using with AWS-Vault

We generally recommend using aws-vault (https://github.com/99designs/aws-vault) for storing AWS credentials and setting 
environment variables.

S3S2 is designed to work seamlessly with aws-vault so that you can say: 

`aws-vault exec jemurai s3s2 share /directory`

and it will pick up the correct AWS credentials.

### Creating a Bucket

TODO - s3s2 will make it easy to provision a bucket.

### Setting a Configuration

Often, we want to share different data with the same partner.
In that case, many of the configuration variables are the same.  Sometimes we want to share a pre-baked configuration 
with the partner so that everything "just works".

To do that, run: 

`s3s2 config`

and you will be prompted for each parameter, which will then be written to a configuration file.

The default configuration file is a `.s3s2` file in your home directory.

You can specify a different configuration file to write by supplying a file parameter as follows:

`s3s2 config --file /Users/mk/mys3s2config.json`

The default .s3s2 file will be used unless when running you 
specify a config file like this:

`s3s2 --config /Users/mk/mys3s2config.json share`

## Running

Running s3s2 is as simple as invoking this from a terminal or command prompt:  
`s3s2 share /directory/to/share>`

## Getting Help

Feel free to create issues and work with us on GitHub: 
https://github.com/Jemurai/s3s2

You can find Jemurai folks to talk about S3S2 on Gitter:
https://gitter.im/jemurai-oss/s3s2