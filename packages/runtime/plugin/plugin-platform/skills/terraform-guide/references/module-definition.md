# module-definition: Module Definition Pattern

Split a module into `variables.tf` (typed + validated inputs), `main.tf` (locals + resources), `outputs.tf` (described outputs). Validate inputs so errors surface before plan.

```hcl
# variables.tf
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "Must be valid IPv4 CIDR"
  }
}

# main.tf
locals {
  common_tags = merge(var.tags, { Module = "network", Environment = var.environment })
}

resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  tags                 = merge(local.common_tags, { Name = "${var.environment}-vpc" })
}

# outputs.tf
output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.this.id
}
```
