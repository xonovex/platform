# locals: Computed Values with Locals

Centralize computed values, transformations, and conditionals in a `locals` block instead of repeating expressions across resources. Reference as `local.name`.

```hcl
locals {
  common_tags   = merge(var.tags, { Environment = var.environment, ManagedBy = "terraform" })
  instance_type = var.environment == "prod" ? "t3.large" : "t3.micro"

  # map az => derived cidr, for for_each
  subnet_cidrs = { for i, az in var.availability_zones : az => cidrsubnet(var.vpc_cidr, 8, i) }

  # flatten nested loops into a single list
  security_rules = flatten([
    for sg in var.security_groups : [
      for rule in sg.rules : { sg_name = sg.name, from_port = rule.from_port }
    ]
  ])
}
```

- Avoid `timestamp()` / `uuid()` in locals used by resource arguments: they change every plan and force perpetual diffs.
