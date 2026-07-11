# remote-state: Remote State Configuration

Store state remotely, encrypted, and locked, with one isolated state key per environment (never share a state file across envs).

```hcl
# environments/prod/backend.tf
terraform {
  backend "s3" {
    bucket         = "myapp-terraform-state-prod"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-state-lock-prod" # lock table prevents concurrent applies
  }
}
```

- `dynamodb_table` gives distributed locking; without it concurrent applies corrupt state.
- Enable S3 bucket versioning to recover a clobbered state file.
- `lifecycle`/`prevent_destroy` belong on resources, not inside the `backend` block.
