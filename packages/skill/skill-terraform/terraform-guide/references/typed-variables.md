# typed-variables: Variable Types and Validation

Give every variable an explicit `type`; add `validation` blocks for business constraints so errors surface before plan. Use `object()` for structured config and `optional(type, default)` (Terraform 1.3+) for flexible schemas.

```hcl
variable "environment" {
  type = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}

variable "instance_count" {
  type    = number
  default = 2
  validation {
    condition     = var.instance_count >= 1 && var.instance_count <= 10
    error_message = "Instance count must be between 1 and 10"
  }
}

variable "database_config" {
  type = object({
    instance_class  = string
    engine_version  = string
    backup_retention = optional(number, 7) # Terraform 1.3+ optional with default
  })
}
```
