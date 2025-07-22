#! /usr/bin/env bash
create_bucket() {
    BUCKET=$1

    # Create Private Bucket
    awslocal s3api create-bucket \
        --acl private \
        --region eu-west-1 \
        --create-bucket-configuration LocationConstraint=eu-west-1 \
        --bucket "$BUCKET" || { echo "Failed to create bucket $BUCKET"; exit 1; }

    # Add Public Access Block
    awslocal s3api put-public-access-block \
        --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true" \
        --bucket "$BUCKET" || { echo "Failed to add public access block to bucket $BUCKET"; exit 1; }

    # Add Default Encryption
    awslocal s3api put-bucket-encryption \
        --bucket "$BUCKET" \
        --server-side-encryption-configuration '{ "Rules": [ { "ApplyServerSideEncryptionByDefault": { "SSEAlgorithm": "AES256" } } ] }' || { echo "Failed to set encryption for bucket $BUCKET"; exit 1; }

    # Add Encryption Policy
    awslocal s3api put-bucket-policy \
        --policy '{
            "Statement": [
                {
                    "Sid": "AllowListingObjects",
                    "Effect": "Allow",
                    "Principal": { "AWS": "*" },
                    "Action": ["s3:ListBucket"],
                    "Resource": "arn:aws:s3:::'${BUCKET}'"
                },
                {
                    "Sid": "DenyUnEncryptedObjectUploads",
                    "Effect": "Deny",
                    "Principal": { "AWS": "*" },
                    "Action": "s3:PutObject",
                    "Resource": "arn:aws:s3:::'${BUCKET}'/*",
                    "Condition": {
                        "StringNotEquals": {
                            "s3:x-amz-server-side-encryption": "AES256"
                        }
                    }
                },
                {
                    "Sid": "DenyUnEncryptedObjectUploads",
                    "Effect": "Deny",
                    "Principal": { "AWS": "*" },
                    "Action": "s3:PutObject",
                    "Resource": "arn:aws:s3:::'${BUCKET}'/*",
                    "Condition": {
                        "Bool": { "aws:SecureTransport": false }
                    }
                }
            ]
        }' \
        --bucket "$BUCKET" || { echo "Failed to set bucket policy for $BUCKET"; exit 1; }
}

awslocal sqs create-queue --queue-name ddc.fifo --attributes FifoQueue=true,ContentBasedDeduplication=true,VisibilityTimeout=30,ReceiveMessageWaitTimeSeconds=0

# Set secrets in Secrets Manager
awslocal secretsmanager create-secret --name local/jwt-key \
    --description "JWT secret for Go services authentication" \
    --secret-string "mysupersecrettestkeythatis128bits"

awslocal ssm put-parameter --name "/local/local-credentials" --type "SecureString" --value '{"opg_document_and_d@publicguardian.gsi.gov.uk":"$2y$10$Xlq5mrdU6ZSh7kU5Yi.vpuCOrWCekNl9BwLcAg5G5bwr22ehTEpEa"}' --overwrite

# S3
create_bucket "opg-backoffice-datastore-local"
create_bucket "opg-backoffice-jobsqueue-local"

# DynamoDB
awslocal dynamodb create-table \
 --region eu-west-1 \
 --table-name Documents \
 --attribute-definitions AttributeName=PK,AttributeType=S \
 --key-schema AttributeName=PK,KeyType=HASH \
 --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000
