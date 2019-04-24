# TODO

List of items remaining to do.

* Check bucket configuration
* Turn print statements into debug logging and make flag for debug.
* Write documentation.
* Error handling.
* Build process.
* Godoc

## Backlog

* Bucket audit
* Decrypt

--

GPG versus symmetric.

- 1. Code cleanup
- 1. Let's measure current state.
1. Generate symmetric - symmetric key.
1. Pick a zip that has AES-256.
    Zip time == Encryption time is ok.
    If Encryption time is 10x then need to revisit.

- 1. Zip individually
- 1. Manifest separate (not gpg)
1. Parallel unzipping
- 1. Create a folder.
- 1. Append extension.  .csv.zip.gpg

1. Decrypt
1. Build instructions (Linux, Mac, Windows)
1. Include public key url
- 1. Double check nested folders

--

Secondary: 
1. Bring your own key.
1. Okta ID, how would that integration work?  Get temp aws access keys.
1. AWS Vault + Okta.
1. Policy
1. Checking Bucket

