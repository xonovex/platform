# root-module: Root Module Pattern

**Guideline:** Keep root modules thin by composing child modules with provider config and environment-specific values

**Rationale:** Root modules orchestrate infrastructure without duplicating logic, modules remain reusable

**Example:**

```hcl
# environments/prod/main.tf
terraform {
  required_version = ">= 1.12"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = "prod"
      Project     = "myapp"
      ManagedBy   = "terraform"
    }
  }
}

locals {
  availability_zones = ["${var.aws_region}a", "${var.aws_region}b", "${var.aws_region}c"]

  common_tags = {
    Project     = "myapp"
    Environment = "prod"
  }
}

module "network" {
  source = "../../modules/network"

  vpc_cidr           = "10.0.0.0/16"
  environment        = "prod"
  availability_zones = local.availability_zones
  tags               = local.common_tags
}

module "compute" {
  source = "../../modules/compute"

  vpc_id            = module.network.vpc_id
  subnet_ids        = module.network.public_subnet_ids
  environment       = "prod"
  instance_type     = "t3.medium"
  desired_capacity  = 3
  tags              = local.common_tags
}
```

**Techniques:**

- **Provider blocks:** Configure required providers with version constraints
- **Default tags:** Use provider default_tags to apply common tags to all resources
- **Thin root:** Root modules only compose child modules, no resource creation
- **Module composition:** Wire module outputs as inputs to dependent modules
- **Environment locals:** Compute availability zones and common tags once
- **Relative sourcing:** Use ../../modules/ paths for portable module references
