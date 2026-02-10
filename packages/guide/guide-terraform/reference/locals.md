# locals: Computed Values with Locals

**Guideline:** Use locals block for computed values, transformations, and conditional logic

**Rationale:** Locals reduce duplication, improve readability, and centralize computation logic

**Example:**

```hcl
locals {
  # Merge tags
  common_tags = merge(
    var.tags,
    {
      Environment = var.environment
      ManagedBy   = "terraform"
      Timestamp   = timestamp()
    }
  )

  # Calculate values
  subnet_count = length(var.availability_zones)

  # Conditional values
  instance_type = var.environment == "prod" ? "t3.large" : "t3.micro"

  # Complex transformations
  subnet_cidrs = {
    for i, az in var.availability_zones :
    az => cidrsubnet(var.vpc_cidr, 8, i)
  }

  # Flatten nested structures
  security_rules = flatten([
    for sg in var.security_groups : [
      for rule in sg.rules : {
        sg_name   = sg.name
        from_port = rule.from_port
        to_port   = rule.to_port
        protocol  = rule.protocol
      }
    ]
  ])
}
```

**Techniques:**

- **Tag merging:** Combine common tags with resource-specific tags using merge()
- **Calculated values:** Compute derived values like counts and indices once
- **Conditional logic:** Use ternary operators for environment-specific resource sizing
- **CIDR subnetting:** Transform VPC CIDR into subnet ranges using cidrsubnet()
- **Flattening:** Collapse nested structures for iteration with flatten()
