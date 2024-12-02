#! /usr/bin/env bash
create_bucket() {
    BUCKET=$1
    # Create Private Bucket
    awslocal s3api create-bucket \
        --acl private \
        --region eu-west-1 \
        --create-bucket-configuration LocationConstraint=eu-west-1 \
        --bucket "$BUCKET"

    # Add Public Access Block
    awslocal s3api put-public-access-block \
        --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true" \
        --bucket "$BUCKET"

    # Add Default Encryption
    awslocal s3api put-bucket-encryption \
        --bucket "$BUCKET" \
        --server-side-encryption-configuration '{ "Rules": [ { "ApplyServerSideEncryptionByDefault": { "SSEAlgorithm": "AES256" } } ] }'

    # Add Encryption Policy
    awslocal s3api put-bucket-policy \
        --policy '{ "Statement": [ { "Sid": "DenyUnEncryptedObjectUploads", "Effect": "Deny", "Principal": { "AWS": "*" }, "Action": "s3:PutObject", "Resource": "arn:aws:s3:::'${BUCKET}'/*", "Condition":  { "StringNotEquals": { "s3:x-amz-server-side-encryption": "AES256" } } }, { "Sid": "DenyUnEncryptedObjectUploads", "Effect": "Deny", "Principal": { "AWS": "*" }, "Action": "s3:PutObject", "Resource": "arn:aws:s3:::'${BUCKET}'/*", "Condition":  { "Bool": { "aws:SecureTransport": false } } } ] }' \
        --bucket "$BUCKET"
}

awslocal sqs create-queue --queue-name ddc.fifo --attributes FifoQueue=true,ContentBasedDeduplication=true,VisibilityTimeout=30,ReceiveMessageWaitTimeSeconds=0
awslocal sqs create-queue --queue-name notify-dead-letter-queue --attributes VisibilityTimeout=30,ReceiveMessageWaitTimeSeconds=0
awslocal sqs create-queue --queue-name notify --attributes file:///etc/localstack/init/ready.d/notify-queue-attributes.json

# Set secrets in Secrets Manager
awslocal secretsmanager create-secret --name local/jwt-key \
    --description "JWT secret for Go services authentication" \
    --secret-string "mysupersecrettestkeythatis128bits"

awslocal ssm put-parameter --name "/local/local-credentials" --type "SecureString" --value '{"opg_document_and_d@publicguardian.gsi.gov.uk":"$2y$10$Xlq5mrdU6ZSh7kU5Yi.vpuCOrWCekNl9BwLcAg5G5bwr22ehTEpEa"}' --overwrite

# S3
create_bucket "opg-backoffice-datastore-local"
create_bucket "opg-backoffice-jobsqueue-local"