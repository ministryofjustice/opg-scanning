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

awslocal dynamodb create-table --table-name frontend-sessions --attribute-definitions AttributeName=id,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000
awslocal dynamodb create-table --table-name last-login-local --attribute-definitions AttributeName=email,AttributeType=S --key-schema AttributeName=email,KeyType=HASH --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000

awslocal dynamodb create-table --table-name events-dedupe --attribute-definitions AttributeName=id,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000 --
awslocal dynamodb update-time-to-live --table-name events-dedupe --time-to-live-specification "Enabled=true, AttributeName=expires"

awslocal sqs create-queue --queue-name ddc.fifo --attributes FifoQueue=true,ContentBasedDeduplication=true,VisibilityTimeout=30,ReceiveMessageWaitTimeSeconds=0
awslocal sqs create-queue --queue-name notify-dead-letter-queue --attributes VisibilityTimeout=30,ReceiveMessageWaitTimeSeconds=0
awslocal sqs create-queue --queue-name notify --attributes file:///etc/localstack/init/ready.d/notify-queue-attributes.json

awslocal dynamodb create-table \
--table-name file-service-requests \
--attribute-definitions \
 AttributeName=Ref,AttributeType=S \
--key-schema \
 AttributeName=Ref,KeyType=HASH \
--provisioned-throughput \
 ReadCapacityUnits=1000,WriteCapacityUnits=1000

awslocal dynamodb update-time-to-live --table-name file-service-requests --time-to-live-specification "Enabled=true, AttributeName=Ttl"

# Set secrets in Secrets Manager
awslocal secretsmanager create-secret --name local/jwt-key \
    --description "JWT secret for Go services authentication" \
    --secret-string "mysupersecrettestkeythatis128bits"

awslocal secretsmanager create-secret --name local/user-hash-salt \
    --description "Email salt for Go services authentication" \
    --secret-string "ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0"

awslocal secretsmanager create-secret --name local/azure-oauth/client-secret \
    --description "Azure OAuth Client Secret" \
    --secret-string "..."

awslocal ssm put-parameter --name "/local/unreturned-lpas-start-date" --type "String" --value "1970-01-01T00:00:00.000Z" --overwrite
awslocal ssm put-parameter --name "/local/local-credentials" --type "SecureString" --value '{"opg_document_and_d@publicguardian.gsi.gov.uk":"$2y$10$Xlq5mrdU6ZSh7kU5Yi.vpuCOrWCekNl9BwLcAg5G5bwr22ehTEpEa"}' --overwrite

awslocal ssm put-parameter --name "/local/api-key/poas-eventbridge" --type "SecureString" --value 'my_auth_token' --overwrite
awslocal ssm put-parameter --name "/local/notify-templates" --type "String" --value '{"SMS_PAYMENT_REQUEST":"46946c8c-5f8b-4293-863b-43d245efb93b"}' --overwrite


aws configure set cli_follow_urlparam false
awslocal events create-event-bus --name local-poas
awslocal events create-connection --name local-poas-api --authorization-type API_KEY --auth-parameters '{
  "ApiKeyAuthParameters": {
    "ApiKeyName": "Authorization",
    "ApiKeyValue": "Bearer my_auth_token"
  }
}'
awslocal events create-api-destination --name local-poas-api --invocation-endpoint "http://api/api/v1/handle-event" --http-method "POST" \
  --connection-arn $(awslocal events describe-connection --name=local-poas-api --output=text --query="ConnectionArn")

awslocal events put-rule --name local-poas-api --event-bus-name local-poas --event-pattern '{}'
awslocal events put-targets --rule local-poas-api --event-bus-name local-poas --targets Id=sirius-api,Arn=$(awslocal events describe-api-destination --name=local-poas-api --output=text --query="ApiDestinationArn")
aws configure set cli_follow_urlparam true

# S3
create_bucket "opg-backoffice-casrec-exports-local"
create_bucket "opg-backoffice-datastore-local"
create_bucket "opg-backoffice-jobsqueue-local"
create_bucket "opg-backoffice-public-api-local"
create_bucket "opg-backoffice-async-uploads-local"
create_bucket "opg-backoffice-finance-local"
create_bucket "opg-backoffice-digideps-local"
create_bucket "opg-backoffice-reduced-fees-uploads-local"
