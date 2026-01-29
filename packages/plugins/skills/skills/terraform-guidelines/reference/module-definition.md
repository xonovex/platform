# module-definition: Module Definition Pattern

**Guideline:** Define reusable modules with typed variables, validation rules, computed locals, and explicit outputs

**Rationale:** Well-defined interfaces make modules composable and prevent configuration errors

**Example:**

```hcl
# modules/network/variables.tf
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  validation {
    condition     = can(cidrhost(var.vpc_cidr, 0))
    error_message = "Must be valid IPv4 CIDR"
  }
}

variable "environment" {
  description = "Environment name"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Must be dev, staging, or prod"
  }
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = []
}

# modules/network/main.tf
locals {
  common_tags = merge(
    var.tags,
    {
      Module      = "network"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  )

  subnet_cidrs = [
    for i, az in var.availability_zones :
    cidrsubnet(var.vpc_cidr, 8, i)
  ]
}

resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.common_tags, {
    Name = "${var.environment}-vpc"
  })
}

# modules/network/outputs.tf
output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.this.id
}

output "vpc_cidr" {
  description = "CIDR block of the VPC"
  value       = aws_vpc.this.cidr_block
}
```

**Techniques:**

- **Type constraints:** Specify all variable types to enable early validation
- **Validation blocks:** Add business logic constraints like CIDR format checking
- **Common locals:** Centralize shared tags and computed values for consistency
- **CIDR transformations:** Use cidrsubnet() in locals for flexible subnet allocation
- **Descriptive outputs:** Export all useful values with clear descriptions
- **Environment isolation:** Use tags and naming patterns for multi-environment support
