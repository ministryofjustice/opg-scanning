#!/usr/bin/env bash

# SQS
queues=$(awslocal sqs list-queues)
echo $queues | grep '"http://sqs.eu-west-1.localhost.localstack.cloud:4566/000000000000/ddc.fifo"' || exit 1
echo $queues | grep '"http://sqs.eu-west-1.localhost.localstack.cloud:4566/000000000000/notify"' || exit 1
echo $queues | grep '"http://sqs.eu-west-1.localhost.localstack.cloud:4566/000000000000/notify-dead-letter-queue"' || exit 1

# S3
buckets=$(awslocal s3 ls)

echo $buckets | grep "opg-backoffice-datastore-local" || exit 1
echo $buckets | grep "opg-backoffice-jobsqueue-local" || exit 1
