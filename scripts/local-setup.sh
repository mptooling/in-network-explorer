#!/usr/bin/env bash
set -euo pipefail

DYNAMO_ENDPOINT="${DYNAMO_ENDPOINT:-http://dynamodb-local:8000}"
SECRETS_ENDPOINT="${SECRETS_ENDPOINT:-http://localstack:4566}"
TABLE_NAME="${DYNAMO_TABLE:-prospects-dev}"
SECRET_NAME="${LINKEDIN_COOKIES_SECRET:-li-cookies-local}"

echo "Creating DynamoDB table '${TABLE_NAME}'..."
aws dynamodb create-table \
  --endpoint-url "$DYNAMO_ENDPOINT" \
  --table-name "$TABLE_NAME" \
  --attribute-definitions \
    AttributeName=PK,AttributeType=S \
  --key-schema \
    AttributeName=PK,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  2>/dev/null && echo "  Table created" || echo "  Table already exists"

echo "Seeding Secrets Manager secret '${SECRET_NAME}'..."
aws secretsmanager create-secret \
  --endpoint-url "$SECRETS_ENDPOINT" \
  --name "$SECRET_NAME" \
  --secret-string "li_at=AQEDAQ_FAKE_COOKIE_FOR_DEV; JSESSIONID=ajax:fake123" \
  2>/dev/null && echo "  Secret created" || echo "  Secret already exists"

echo "Local setup complete."
