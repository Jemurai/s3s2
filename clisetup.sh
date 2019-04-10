# Show the kms key
aws-vault exec jemurai-mkonda -- aws kms create-key --description "Testing s3s2 2"
aws-vault exec jemurai-mkonda -- aws kms describe-key --key-id (key_id)

# S3
aws-vault exec jemurai-mkonda-admin -- aws s3 mb s3://s3s2-demo1

# Show what is in S3
aws-vault exec jemurai-mkonda -- aws s3 ls
aws-vault exec jemurai-mkonda -- aws s3 ls s3s2-demo



## s3s2
