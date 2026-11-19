# remote-state: Remote State Configuration

Store state remotely, encrypted, and locked, with one isolated state key per environment. Never share a state file across environments.

```hcl
# environments/prod/backend.tf
terraform {
  backend "s3" {
    bucket       = "myapp-terraform-state-prod"
    key          = "prod/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true
  }
}
```

- `use_lockfile = true` enables S3 state locking; grant read/write/delete access to the `.tflock` object.
- DynamoDB-based locking is deprecated. Configure `dynamodb_table` only while migrating clients that cannot yet use S3 lockfiles, then remove it.
- Enable S3 bucket versioning to recover a clobbered state file.
- `lifecycle`/`prevent_destroy` belong on resources, not inside the `backend` block.
