#!/bin/bash
set +e  # Continue on failures

echo "=== AWS Resource Cleanup ==="

# Function to safely run AWS commands with retries
safe_aws_command() {
  local max_attempts=3
  local attempt=1
  local command="$@"
  while [ $attempt -le $max_attempts ]; do
    if eval "$command"; then
      return 0
    fi
    echo "Retry $attempt/$max_attempts failed: $command"
    sleep 5
    attempt=$((attempt + 1))
  done
  echo "Command failed after $max_attempts attempts: $command"
  return 1
}

# Clean up S3 buckets from SAM CLI test stacks
echo "=== Cleaning S3 buckets ==="
TEST_PATTERNS=("sam-app" "test-" "integration-test" "samcli" "aws-sam-cli-managed")

for pattern in "${TEST_PATTERNS[@]}"; do
  STACKS=$(aws cloudformation list-stacks --region $AWS_DEFAULT_REGION --stack-status-filter CREATE_COMPLETE UPDATE_COMPLETE ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE --query "StackSummaries[?contains(StackName, '$pattern')].[StackName]" --output text 2>/dev/null || true)

  for stack in $STACKS; do
    echo "Processing stack: $stack"
    
    # Get S3 buckets from stack
    BUCKET_NAMES=$(aws cloudformation describe-stacks --stack-name "$stack" --region $AWS_DEFAULT_REGION --query 'Stacks[0].Outputs[?contains(OutputKey, `Bucket`) || contains(OutputKey, `bucket`)].OutputValue' --output text 2>/dev/null || true)
    RESOURCE_BUCKETS=$(aws cloudformation describe-stack-resources --stack-name "$stack" --region $AWS_DEFAULT_REGION --query 'StackResources[?ResourceType==`AWS::S3::Bucket`].PhysicalResourceId' --output text 2>/dev/null || true)

    # Empty buckets (don't delete them)
    for bucket in $BUCKET_NAMES $RESOURCE_BUCKETS; do
      if [ -n "$bucket" ] && [ "$bucket" != "None" ]; then
        echo "Emptying S3 bucket: $bucket"
        if aws s3api head-bucket --bucket "$bucket" 2>/dev/null; then
          safe_aws_command "aws s3 rm s3://$bucket --recursive --quiet" || true
          echo "✅ Emptied bucket: $bucket"
        fi
      fi
    done
  done
done

# Clean up Lambda functions
echo "=== Cleaning Lambda functions ==="
LAMBDA_PATTERNS=("sam-app" "test-" "HelloWorld")
for pattern in "${LAMBDA_PATTERNS[@]}"; do
  FUNCTIONS=$(aws lambda list-functions --region $AWS_DEFAULT_REGION --query "Functions[?contains(FunctionName, '$pattern')].FunctionName" --output text 2>/dev/null || true)
  for func in $FUNCTIONS; do
    echo "Deleting Lambda function: $func"
    safe_aws_command "aws lambda delete-function --function-name '$func' --region $AWS_DEFAULT_REGION" || true
  done
done

# Clean up API Gateway APIs
echo "=== Cleaning API Gateway APIs ==="
APIS=$(aws apigateway get-rest-apis --region $AWS_DEFAULT_REGION --query 'items[?contains(name, `sam-app`) || contains(name, `test-`) || contains(name, `Test`)].id' --output text 2>/dev/null || true)
for api in $APIS; do
  echo "Deleting API Gateway API: $api"
  safe_aws_command "aws apigateway delete-rest-api --rest-api-id '$api' --region $AWS_DEFAULT_REGION" || true
done

echo "✅ Cleanup completed"