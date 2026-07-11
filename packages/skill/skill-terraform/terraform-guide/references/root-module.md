# root-module: Root Module Pattern

Keep root modules thin: they pin `required_version`/`required_providers`, configure the provider, and wire child modules together — no resource creation of their own. Apply cross-cutting tags via provider `default_tags`. Feed module outputs into dependent modules as inputs.

```hcl
# environments/prod/main.tf
terraform {
  required_version = ">= 1.12"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
  }
}

provider "aws" {
  region      = var.aws_region
  default_tags { tags = { Environment = "prod", ManagedBy = "terraform" } }
}

module "network" {
  source      = "../../modules/network"
  vpc_cidr    = "10.0.0.0/16"
  environment = "prod"
}

module "compute" {
  source      = "../../modules/compute"
  vpc_id      = module.network.vpc_id       # wire output -> input
  subnet_ids  = module.network.public_subnet_ids
  environment = "prod"
}
```
