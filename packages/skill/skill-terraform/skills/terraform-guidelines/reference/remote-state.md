# remote-state: Remote State Configuration

**Guideline:** Store state remotely with encryption and locking, isolate per environment

**Rationale:** Remote state enables team collaboration and prevents concurrent modification conflicts

**Example:**

```hcl
# environments/prod/backend.tf
terraform {
  backend "s3" {
    bucket         = "myapp-terraform-state-prod"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock-prod"

    # Prevent accidental state deletion
    lifecycle {
      prevent_destroy = true
    }
  }
}

# Each environment has isolated state
# dev: s3://myapp-tf-state-dev/dev/terraform.tfstate
# staging: s3://myapp-tf-state-staging/staging/terraform.tfstate
# prod: s3://myapp-tf-state-prod/prod/terraform.tfstate
```

**Techniques:**

- **S3 backend:** Store state in versioned, encrypted S3 buckets
- **DynamoDB locking:** Use DynamoDB tables to prevent concurrent modifications
- **Encryption:** Enable server-side encryption on S3 backend buckets
- **Per-environment isolation:** Create separate state buckets for dev, staging, prod
- **State lifecycle:** Prevent accidental state deletion with versioning and backups
